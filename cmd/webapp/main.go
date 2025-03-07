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

const PROD_ENV = "production"

func main() {
	// Setup Feature flags
	configFile := flag.String("config", "", "Path to the feature flag config file (JSON or YAML)")
	featureFlags := flag.String("feature", "", "Comma-separated list of feature flags to enable (e.g., feature1,feature2)")
	disableFlags := flag.String("disable-feature", "", "Comma-separated list of feature flags to disable (e.g., feature1,feature2)")
	flag.Parse()
	ff := util.NewFeatureFlags()

	// Load feature flags from the file, if specified
	if *configFile != "" {
		if err := ff.LoadFlagsFromFile(*configFile); err != nil {
			fmt.Printf("Error loading feature flags from file: %v\n", err)
			os.Exit(1)
		}

		// Start watching for changes to the feature flags file
		go func() {
			if err := ff.WatchFile(*configFile); err != nil {
				fmt.Printf("Error watching feature flag file: %v\n", err)
			}
		}()
	}

	// Process feature flags to enable
	if *featureFlags != "" {
		flags := strings.Split(*featureFlags, ",")
		for _, flag := range flags {
			// You could also support enabling flag with specific values if needed
			ff.Set(flag, true)
		}
	}

	// Process disable flags
	if *disableFlags != "" {
		flags := strings.Split(*disableFlags, ",")
		for _, flag := range flags {
			ff.Set(flag, false)
		}
	}

	// Print the state of all feature flags
	fmt.Println("Feature flags:")
	for flag, enabled := range ff.GetAll() {
		fmt.Printf("%s=%v\n", flag, enabled)
	}

	rootMux := http.NewServeMux()
	appMux := http.NewServeMux()
	rootMux.Handle("/app/", http.StripPrefix("/app", appMux))

	appMux.HandleFunc("/login", loginEndpoint)
	appMux.HandleFunc("/dashboard", protectedHandler("/app/login", dashboardEndpoint(ff)))
	appMux.Handle("/", http.RedirectHandler("/app/dashboard", http.StatusMovedPermanently))

	f = staticFiles
	if os.Getenv("ENV") != PROD_ENV {
		_, file, _, ok := runtime.Caller(0)
		if !ok {
			fmt.Println("unable to get caller info")
			return
		}

		srcDir := filepath.Dir(file)
		f = os.DirFS(srcDir)
	}

	if ps := os.Getenv("HAPROXY_MANAGER_PASSWORD_HASH"); ps != "" {
		PASSWORD_SALT = ps
	}

	if ps := os.Getenv("HAPROXY_MANAGER_JWT_SECRET"); ps != "" {
		JWT_SECRET = ps
	}

	rootMux.Handle("/static/", http.FileServer(http.FS(f)))
	rootMux.Handle("/", http.RedirectHandler("/app", http.StatusMovedPermanently))

	log.Println("Server Starting")
	http.ListenAndServe(":8080", loggingMiddleware(rootMux))
}
