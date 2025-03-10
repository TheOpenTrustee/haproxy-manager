package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/jaitaiwan/haproxy-manager/internal/db"
	"github.com/jaitaiwan/haproxy-manager/internal/util"
)

// LoginPageHandler serves the login page.
func loginHandler(ff *util.FeatureFlags) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Redirect to the dashboard if the user is already authenticated.
		if _, ok := util.UserFromAuthCookie(r, JWT_SECRET); ok {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		if r.Method == http.MethodPost {

			// Parse the form data.
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Invalid form data", http.StatusBadRequest)
				return
			}

			username := strings.TrimSpace(r.FormValue("username"))
			password := strings.TrimSpace(r.FormValue("password"))

			if username == "" || password == "" {
				http.Error(w, "Username and password are required", http.StatusBadRequest)
				return
			}

			// Check credentials against the database.
			// Retrieve the user from the database
			user, err := db.GetUserByUsername(username)
			if err != nil {
				if err == db.ErrorInvalidUser {
					http.Error(w, "Invalid username or password", http.StatusUnauthorized)
					return
				}
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				log.Printf("Error querying user: %v\n", err)
				return
			}

			// Compare the hashed password.
			if !util.ComparePassword(user.Password, password) {
				http.Error(w, "Invalid username or password", http.StatusUnauthorized)
				return
			}

			// Sign the token
			tokenString, err := util.TokenString(JWT_SECRET, user.ID)
			if err != nil {
				http.Error(w, "Failed to create token", http.StatusInternalServerError)
				log.Printf("Error signing token: %v\n", err)
				return
			}

			// Set the JWT as a cookie
			http.SetCookie(w, &http.Cookie{
				Name:     "auth_token",
				Value:    tokenString,
				Path:     "/",
				HttpOnly: true, // Prevent access via JavaScript
				Secure:   true, // Use Secure cookies in production (requires HTTPS)
				SameSite: http.SameSiteStrictMode,
			})
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		}

		// Parse the index.html and login.partial.html templates from the embedded filesystem.
		tmpl, err := template.ParseFS(f, "index.html", "auth/login.partial.html")
		if err != nil {
			http.Error(w, "Could not parse templates", http.StatusInternalServerError)
			fmt.Printf("err: %s\n", err)
			return
		}

		// Example data to pass to the template (replace with actual data as needed).
		data := DefaultViewData{
			Title: "Login Page",
			Flags: ff,
		}

		// Execute the template and write the output to the response.
		WriteTemplate(w, tmpl, data)
	})
}
