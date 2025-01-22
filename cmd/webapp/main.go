package main

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

var f fs.FS

func main() {
	rootMux := http.NewServeMux()
	appMux := http.NewServeMux()
	rootMux.Handle("/app/", http.StripPrefix("/app", appMux))

	appMux.HandleFunc("/login", loginEndpoint)
	appMux.HandleFunc("/dashboard", protectedHandler("/app/login", dashboardEndpoint))
	appMux.Handle("/", http.RedirectHandler("/app/dashboard", http.StatusMovedPermanently))

	f = staticFiles
	if os.Getenv("ENV") != "production" {
		_, file, _, ok := runtime.Caller(0)
		if !ok {
			fmt.Println("unable to get caller info")
			return
		}

		srcDir := filepath.Dir(file)
		f = os.DirFS(srcDir)
	}
	rootMux.Handle("/static/", http.FileServer(http.FS(f)))
	rootMux.Handle("/", http.RedirectHandler("/app", http.StatusMovedPermanently))

	http.ListenAndServe(":8080", rootMux)
}
