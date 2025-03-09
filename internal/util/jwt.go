package util

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrorUnexpectedSigningMethod = errors.New("unexpected signing method")
var ErrorExpiredToken = errors.New("token has expired")
var ErrorInvalidUser = errors.New("invalid user")
var ErrorInvalidToken = errors.New("invalid token")

type AuthContextKey string

// isAuthenticated checks if the incoming request has a valid JWT token.
func UserFromToken(jwt_secret, jwt_token string) (isAuthenticated bool, userId string, err error) {
	// Parse and validate the JWT token
	token, err := jwt.Parse(jwt_token, func(t *jwt.Token) (interface{}, error) {
		// Validate the algorithm
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrorUnexpectedSigningMethod
		}
		return []byte(jwt_secret), nil
	})

	if err != nil {
		return false, "", err
	}

	// Verify claims (e.g., expiration)
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Check the "exp" (expiration) claim
		if exp, ok := claims["exp"].(float64); ok {
			if time.Unix(int64(exp), 0).Before(time.Now()) {
				return false, "", ErrorExpiredToken
			}
		}

		// Retrieve the user ID or other relevant claim
		userID, ok := claims["user_id"].(float64)
		if !ok {
			return false, "", ErrorInvalidUser
		}

		// Authentication successful
		return true, strconv.Itoa(int(userID)), nil
	}

	return false, "", ErrorInvalidToken
}

func TokenString(jwt_secret string, userID int) (string, error) {
	// Create a JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(), // Token expires in 24 hours
	})

	// Sign the token
	tokenString, err := token.SignedString([]byte(jwt_secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
