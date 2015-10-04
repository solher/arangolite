// Package arangolite provides a lightweight ArangoDB driver.
package arangolite

import (
	"fmt"
	"log"
	"os"
)

// DB represents an access to an ArangoDB database.
type DB struct {
	url, database, user, password string
	logger                        *log.Logger
}

// New returns a new DB object.
func New(logEnabled bool) *DB {
	var out *os.File

	if logEnabled {
		out = os.Stdout
	}

	return &DB{logger: log.New(out, fmt.Sprintf("\n[Arangolite] "), 0)}
}

// Connect initialize a DB object with the database url and credentials.
func (db *DB) Connect(url, database, user, password string) {
	db.url = url
	db.database = database
	db.user = user
	db.password = password
}

// RunAQL is a simple shortcut allowing to run AQL queries.
func (db *DB) RunAQL(aql string, params ...interface{}) ([]byte, error) {
	q := NewQuery(aql, params...)
	return q.Run(db)
}
