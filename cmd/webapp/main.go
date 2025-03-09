package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jaitaiwan/haproxy-manager/internal/util"
)

var f fs.FS
var PASSWORD_SALT = "randompasswordsalt"
var JWT_SECRET = "randomjwtsecret"
var CONFIG_FILE = "config.yml"
var FEATURE_FLAGS = ""
var DISABLE_FLAGS = ""

const PROD_ENV = "production"

func main() {
	setupEnv()
	setupArgs()

	mux := setupMux()

	log.Println("Server Starting")
	http.ListenAndServe(":8080", loggingMiddleware(mux))
}

func setupArgs() {
	// Setup Feature flags
	flag.StringVar(&CONFIG_FILE, "config", "", "Path to the feature flag config file (JSON or YAML)")
	flag.StringVar(&FEATURE_FLAGS, "feature", "", "Comma-separated list of feature flags to enable (e.g., feature1,feature2)")
	flag.StringVar(&DISABLE_FLAGS, "disable-feature", "", "Comma-separated list of feature flags to disable (e.g., feature1,feature2)")
	flag.Parse()
}

func setupFlags() *util.FeatureFlags {
	ff := util.NewFeatureFlags()

	// Load feature flags from the file, if specified
	if CONFIG_FILE != "" {
		if err := ff.LoadFlagsFromFile(CONFIG_FILE); err != nil {
			fmt.Printf("Error loading feature flags from file: %v\n", err)
			os.Exit(1)
		}

		// Start watching for changes to the feature flags file
		go func() {
			if err := ff.WatchFile(CONFIG_FILE); err != nil {
				fmt.Printf("Error watching feature flag file: %v\n", err)
			}
		}()
	}

	// Process feature flags to enable
	if FEATURE_FLAGS != "" {
		flags := strings.Split(FEATURE_FLAGS, ",")
		for _, flag := range flags {
			// You could also support enabling flag with specific values if needed
			ff.Set(flag, true)
		}
	}

	// Process disable flags
	if DISABLE_FLAGS != "" {
		flags := strings.Split(DISABLE_FLAGS, ",")
		for _, flag := range flags {
			ff.Set(flag, false)
		}
	}

	// Print the state of all feature flags
	fmt.Println("Feature flags:")
	for flag, enabled := range ff.GetAll() {
		fmt.Printf("%s=%v\n", flag, enabled)
	}

	return ff
}

func setupEnv() {
	if ph := os.Getenv("HAPROXY_MANAGER_PASSWORD_SALT"); ph != "" {
		PASSWORD_SALT = ph
	}

	if js := os.Getenv("HAPROXY_MANAGER_JWT_SECRET"); js != "" {
		JWT_SECRET = js
	}
}

func setupFS() {
	var err error
	f, err = fs.Sub(staticFiles, "static")
	if err != nil {
		fmt.Println("unable to get static files")
		os.Exit(1)
	}

	if os.Getenv("ENV") != PROD_ENV {
		_, file, _, ok := runtime.Caller(0)
		if !ok {
			fmt.Println("unable to get caller info")
			os.Exit(1)
		}

		srcDir := filepath.Dir(file)
		f = os.DirFS(srcDir + string(os.PathSeparator) + "static")
	}
}

func setupMux() *http.ServeMux {
	ff := setupFlags()
	rootMux := http.NewServeMux()
	appMux := http.NewServeMux()
	hostsMux := newHostsMux(ff)
	rootMux.Handle("/app/", http.StripPrefix("/app", appMux))

	appMux.Handle("/login", loginHandler(ff))
	appMux.Handle("/dashboard", protect("/app/login", dashboardEndpoint(ff)))
	appMux.Handle("/hosts/", protect("/app/login", http.StripPrefix("/hosts", hostsMux)))
	appMux.Handle("/", http.RedirectHandler("/app/dashboard", http.StatusMovedPermanently))

	setupFS()

	rootMux.Handle("/assets/", http.FileServerFS(f))
	rootMux.Handle("/", http.RedirectHandler("/app", http.StatusMovedPermanently))
	return rootMux
}
