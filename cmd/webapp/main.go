package main

import (
	"io/fs"
	"net/http"
	"os"
)

var f fs.FS

func main() {
	rootMux := http.NewServeMux()
	appMux := http.NewServeMux()
	rootMux.Handle("/app/", http.StripPrefix("/app", appMux))
	rootMux.Handle("/", http.RedirectHandler("/app", http.StatusMovedPermanently))

	appMux.HandleFunc("/login", loginEndpoint)
	appMux.HandleFunc("/dashboard", protectedHandler("/app/login", dashboardEndpoint))
	appMux.Handle("/", http.RedirectHandler("/app/dashboard", http.StatusMovedPermanently))

	f = staticFiles
	if os.Getenv("ENV") != "production" {
		f = os.DirFS(".") // TODO fix this
	}

	http.ListenAndServe(":8080", combinedHandler(rootMux, http.FileServer(http.FS(f)).ServeHTTP))
}
