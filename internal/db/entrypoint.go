package db

import (
	"database/sql"
	"fmt"

	"github.com/jaitaiwan/haproxy-manager/internal/errors"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

var ErrorGenericSQL = errors.New("generic SQL error")

func init() {
	var err error
	db, err = sql.Open("sqlite3", "./data/manager.db")
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to the database: %s", err))
	}

	setupDefaultUsers(db)
}
