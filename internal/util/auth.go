package util

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"

	"golang.org/x/crypto/argon2"
)

// comparePassword verifies if the plain-text password matches the hashed password.
func ComparePassword(storedHash, providedPassword string) bool {
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

func HashNewPassword(password string) (string, error) {
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

func UserFromAuthCookie(r *http.Request, JWT_SECRET string) (string, bool) {
	// Retrieve the auth token from the cookie
	authHeader, err := r.Cookie("auth_token")
	if err != nil {
		log.Println("missing auth header")
		return "", false
	}

	// Check if the user is authenticated and retrieve the userID
	isAuth, userId, err := UserFromToken(JWT_SECRET, authHeader.Value)
	if err != nil {
		return "", false
	}

	return userId, isAuth
}
