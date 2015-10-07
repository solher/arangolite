// Package arangolite provides a lightweight ArangoDB driver.
package arangolite

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

// DB represents an access to an ArangoDB database.
type DB struct {
	url, database, user, password string
	conn                          *http.Client
	logger                        *log.Logger
}

// New returns a new DB object.
func New(logEnabled bool) *DB {
	var out *os.File

	if logEnabled {
		out = os.Stdout
	}

	return &DB{conn: &http.Client{}, logger: log.New(out, fmt.Sprintf("\n[Arangolite] "), 0)}
}

// Connect initialize a DB object with the database url and credentials.
func (db *DB) Connect(url, database, user, password string) {
	db.url = url
	db.database = database
	db.user = user
	db.password = password
}
