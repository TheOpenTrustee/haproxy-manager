package main

import (
	"context"
	"errors"
	"net/http"

	"github.com/jaitaiwan/haproxy-manager/internal/util"
	_ "github.com/mattn/go-sqlite3"
)

var userIDKey util.AuthContextKey = "user_id"
var ErrorMissingAuthHeader = errors.New("missing auth header")

func protect(loginUrl string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId, ok := util.UserFromAuthCookie(r, JWT_SECRET)
		if !ok {
			http.Redirect(w, r, loginUrl, http.StatusTemporaryRedirect)
			return
		}

		// Store the userID in the context
		ctx := context.WithValue(r.Context(), userIDKey, userId)

		// Call the next handler if authenticated.
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
