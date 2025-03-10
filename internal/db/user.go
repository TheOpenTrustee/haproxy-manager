package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/jaitaiwan/haproxy-manager/internal/errors"

	"github.com/jaitaiwan/haproxy-manager/internal/util"
)

var ErrorInvalidUser = errors.New("invalid username or password")

// User represents a user in the database.
type User struct {
	ID       int
	Username string
	Password string // Stored as a hashed password
}

func setupDefaultUsers(db *sql.DB) {
	var err error
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

	password, err = util.HashNewPassword(password)
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

func GetUserByUsername(username string) (*User, error) {
	var user User
	err := db.QueryRow("SELECT id, username, password FROM users WHERE username = ?", username).Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrorInvalidUser
		}
		log.Printf("Error querying user: %v\n", err)
		return nil, errors.Join(ErrorGenericSQL, err)
	}

	return &user, nil
}
