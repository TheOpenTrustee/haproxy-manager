package main

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/jaitaiwan/haproxy-manager/internal/util"
	dpc "github.com/jaitaiwan/haproxy-manager/internal/util/dataplane-client"
)

// LoginPageHandler serves the login page.
func dashboardEndpoint(ff *util.FeatureFlags) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the index.html and login.partial.html templates from the embedded filesystem.
		tmpl, err := template.ParseFS(f, "static/index.html", "static/nav.partial.html", "static/dash.partial.html")
		if err != nil {
			http.Error(w, "Could not parse templates", http.StatusInternalServerError)
			return
		}

		// Execute the template and write the output to the response.
		w.Header().Set("Content-Type", "text/html")

		frontends, e, err := dpc.GetFrontends()
		if err != nil {
			http.Error(w, "Could not get frontends", http.StatusInternalServerError)
			return
		}

		if e != nil {
			log.Printf("Error: %v", e)
		}

		data := map[string]interface{}{
			"Title":     "Dashboard",
			"Content":   "dash.partial.html",
			"Flags":     ff,
			"Frontends": frontends,
		}

		if err := tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
			http.Error(w, "Could not render page", http.StatusInternalServerError)
		}
	}
}

// responseWriter is a wrapper around http.ResponseWriter that captures the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
	wroteHeader  bool
}

// WriteHeader captures the status code and writes the header.
func (rw *responseWriter) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.statusCode = code
		rw.ResponseWriter.WriteHeader(code)
		rw.wroteHeader = true
	}
}

// Write captures the number of bytes written and writes the response.
func (rw *responseWriter) Write(data []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(data)
	rw.bytesWritten += n
	return n, err
}

func loggingMiddleware(next http.Handler) http.Handler {
	log.Println("Server Started")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Optionally log headers (comment out if not needed).
		// for name, values := range r.Header {
		// 	for _, value := range values {
		// 		log.Printf("Header: %s=%s", name, value)
		// 	}
		// }

		// Pass the request to the next handler.
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)

		// Log request completion.
		duration := time.Since(start)

		referrer := r.Header.Get("Referer")
		if referrer == "" {
			referrer = "-" // Default value if Referer is empty
		}

		userAgent := r.Header.Get("User-Agent")
		if userAgent == "" {
			userAgent = "-" // Default value if User-Agent is empty
		}

		// Log request details.
		log.Printf(
			"%s - - \"%s %s %s\" %d %d \"%s\" \"%s\" %v",
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
			r.Proto,
			rw.statusCode,
			rw.bytesWritten,
			referrer,
			userAgent,
			duration,
		)
	})
}
