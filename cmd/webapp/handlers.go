package main

import (
	"html/template"
	"net/http"
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
