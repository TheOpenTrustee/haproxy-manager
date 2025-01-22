package main

import (
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"time"
)

// LoginPageHandler serves the login page.
func dashboardEndpoint(w http.ResponseWriter, r *http.Request) {
	// Parse the index.html and login.partial.html templates from the embedded filesystem.
	tmpl, err := template.ParseFS(staticFiles, "static/index.html", "static/dash.partial.html")
	if err != nil {
		http.Error(w, "Could not parse templates", http.StatusInternalServerError)
		return
	}

	// Example data to pass to the template (replace with actual data as needed).
	data := map[string]interface{}{
		"Title":   "Login Page",
		"Content": "dash.partial.html",
	}

	// Execute the template and write the output to the response.
	w.Header().Set("Content-Type", "text/html")

	if err := tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, "Could not render page", http.StatusInternalServerError)
	}
}

// ServeStaticWithNoCache is a handler that serves static files and disables caching.
func serveStaticWithNoCache(staticFS fs.FS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set cache control headers to prevent caching
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		// Serve the file dynamically from the directory.
		file, err := staticFS.Open(r.URL.Path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer file.Close()

		// Check if the file implements io.ReadSeeker, which is required by ServeContent
		if readSeeker, ok := file.(*os.File); ok {
			// Serve the file content using http.ServeContent
			http.ServeContent(w, r, r.URL.Path, time.Now(), readSeeker)
		} else {
			http.Error(w, "File does not implement io.ReadSeeker", http.StatusInternalServerError)
		}
	}
}

func combinedHandler(mux *http.ServeMux, staticHandler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the handler and pattern from the mux.
		handler, pattern := mux.Handler(r)

		// If a pattern is matched, call the handler.
		if pattern != "" {
			handler.ServeHTTP(w, r)
			return
		}

		// Fall back to the static file handler for unmatched routes.
		if strings.HasPrefix(r.URL.Path, "/static/") {
			staticHandler(w, r)
			return
		}

		// Default 404 for other unmatched routes.
		http.NotFound(w, r)
	}
}
