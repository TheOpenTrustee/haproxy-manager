package main

import (
	"fmt"
	"html/template"
	"net/http"
)

// LoginPageHandler serves the login page.
func loginEndpoint(w http.ResponseWriter, r *http.Request) {
	// Parse the index.html and login.partial.html templates from the embedded filesystem.
	tmpl, err := template.ParseFS(f, "static/index.html", "static/login.partial.html")
	if err != nil {
		http.Error(w, "Could not parse templates", http.StatusInternalServerError)
		fmt.Printf("err: %s\n", err)
		return
	}

	// Example data to pass to the template (replace with actual data as needed).
	data := map[string]interface{}{
		"Title":   "Login Page",
		"Content": "login.partial.html",
	}

	// Execute the template and write the output to the response.
	w.Header().Set("Content-Type", "text/html")

	if err := tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, "Could not render page", http.StatusInternalServerError)
	}
}

func isAuthenticated(r *http.Request) bool {
	// Example check: Look for a "session_token" cookie.
	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		return false
	}
	// Additional validation of the session token can be added here.
	return true
}

// ProtectedHandler wraps another handler and redirects unauthenticated users.
func protectedHandler(loginPath string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isAuthenticated(r) {
			// Redirect to the login page if the user is not authenticated.
			http.Redirect(w, r, loginPath, http.StatusFound)
			return
		}
		// Call the next handler if authenticated.
		next(w, r)
	}
}
