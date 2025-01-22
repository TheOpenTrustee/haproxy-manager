package main

import (
	"html/template"
	"log"
	"net/http"
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

// responseWriter is a wrapper around http.ResponseWriter that captures the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

// WriteHeader captures the status code and writes the header.
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
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
