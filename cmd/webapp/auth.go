package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/argon2"
)

var db *sql.DB

type AuthContextKey string

var userIDKey AuthContextKey = "user_id"

// User represents a user in the database.
type User struct {
	ID       int
	Username string
	Password string // Stored as a hashed password
}

func init() {
	var err error
	db, err = sql.Open("sqlite3", "./data/manager.db")
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to the database: %s", err))
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY,
			username TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL -- Store hashed passwords (Argon2id).
		)
	`)
	if err != nil {
		panic(fmt.Sprintf("Failed to create table: %s", err))
	}

	username := "admin"
	password := "admin"
	if un := os.Getenv("HAPROXY_MANAGER_ADMIN_USER"); un != "" {
		username = un
	}

	if pw := os.Getenv("HAPROXY_MANAGER_ADMIN_PASS"); pw != "" {
		password = pw
	}

	// Check if the user already exists
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)", username).Scan(&exists)
	if err != nil {
		panic(fmt.Sprintf("Failed finding default user: %s", err))
	}
	if exists {
		log.Printf("Default user %s already exists", username)
		return
	}

	password, err = hashNewPassword(password)
	if err != nil {
		panic(fmt.Sprintf("Failed to has default admin user password: %s", err))
	}

	sql := fmt.Sprintf(`INSERT INTO users (username, password) VALUES ('%s', '%s')`, username, password)
	log.Printf("debug: %s", sql)
	_, err = db.Exec(sql)
	if err != nil {
		panic(fmt.Sprintf("Failed to create default admin user: %s", err))
	}
}

// LoginPageHandler serves the login page.
func loginEndpoint(w http.ResponseWriter, r *http.Request) {
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
		var user User
		err := db.QueryRow("SELECT id, username, password FROM users WHERE username = ?", username).Scan(&user.ID, &user.Username, &user.Password)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Invalid username or password", http.StatusUnauthorized)
				return
			}
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			log.Printf("Error querying user: %v\n", err)
			return
		}

		// Compare the hashed password.
		if !comparePassword(user.Password, password) {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}

		// Create a JWT
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": user.ID,
			"exp":     time.Now().Add(24 * time.Hour).Unix(), // Token expires in 24 hours
		})

		// Sign the token
		tokenString, err := token.SignedString([]byte(JWT_SECRET))
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

// ProtectedHandler wraps another handler and redirects unauthenticated users.
func protectedHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if the user is authenticated and retrieve the userID
		isAuth, userID, err := isAuthenticated(r)
		if !isAuth {
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Store the userID in the context
		ctx := context.WithValue(r.Context(), userIDKey, userID)

		// Call the next handler if authenticated.
		next(w, r.WithContext(ctx))
	}
}

// comparePassword verifies if the plain-text password matches the hashed password.
func comparePassword(storedHash, providedPassword string) bool {
	// Decode the stored hash into its salt and hash components
	if len(storedHash) < 32 {
		return false
	}

	// Extract salt (first 16 bytes) and hash (remaining bytes)
	salt, err := hex.DecodeString(storedHash[:32]) // First 32 hex characters (16 bytes)
	if err != nil {
		return false
	}

	storedHashBytes, err := hex.DecodeString(storedHash[32:]) // Remaining bytes
	if err != nil {
		return false
	}

	// Recreate the hash from the provided password using the same salt
	computedHash := argon2.IDKey([]byte(providedPassword), salt, 1, 64*1024, 4, 32)

	// Compare the computed hash with the stored hash
	return string(computedHash) == string(storedHashBytes)
}

// isAuthenticated checks if the incoming request has a valid JWT token.
func isAuthenticated(r *http.Request) (bool, string, error) {
	// Extract the JWT token from the Authorization header
	authHeader := r.Header.Get("auth_token")
	if authHeader == "" {
		return false, "", errors.New("authorization header missing")
	}

	// The format of the header should be "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return false, "", errors.New("invalid authorization header format")
	}

	tokenString := parts[1]

	// Parse and validate the JWT token
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		// Validate the algorithm
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return JWT_SECRET, nil
	})

	if err != nil {
		return false, "", err
	}

	// Verify claims (e.g., expiration)
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Check the "exp" (expiration) claim
		if exp, ok := claims["exp"].(float64); ok {
			if time.Unix(int64(exp), 0).Before(time.Now()) {
				return false, "", errors.New("token has expired")
			}
		}

		// Retrieve the user ID or other relevant claim
		userID, ok := claims["user_id"].(string)
		if !ok {
			return false, "", errors.New("user ID not found in token")
		}

		// Authentication successful
		return true, userID, nil
	}

	return false, "", errors.New("invalid token")
}

func hashNewPassword(password string) (string, error) {
	// Generate a random 16-byte salt
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// Hash the password using Argon2
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	// Combine salt and hash as a single string for storage
	return hex.EncodeToString(salt) + hex.EncodeToString(hash), nil
}
