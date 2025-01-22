package main

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

var f fs.FS
var PASSWORD_SALT = "randompasswordsalt"
var JWT_SECRET = "randomjwtsecret"

const PROD_ENV = "production"

func main() {
	rootMux := http.NewServeMux()
	appMux := http.NewServeMux()
	rootMux.Handle("/app/", http.StripPrefix("/app", appMux))

	appMux.HandleFunc("/login", loginEndpoint)
	appMux.HandleFunc("/dashboard", protectedHandler(dashboardEndpoint))
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
