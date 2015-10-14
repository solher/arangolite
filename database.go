// Package arangolite provides a lightweight ArangoDB driver.
package arangolite

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
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

// runQuery executes a query at the path passed as argument.
func (db *DB) runQuery(path string, query []byte) ([]byte, error) {
	if query == nil || len(query) == 0 {
		return nil, errors.New("nil or empty query")
	}

	r, err := db.conn.Post(getFullURL(db, path), "application/json", bytes.NewBuffer(query))
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(r.Body)
}

func (db *DB) logBegin(msg, path string, jsonQuery []byte) {

	db.logger.Printf("%s %s %s | URL: %s\n    %s", blue, msg, reset, getFullURL(db, path), indentJSON(jsonQuery))
}

func (db *DB) logResult(result []byte, cached bool, execTime time.Duration) {
	if cached {
		db.logger.Printf("%s RESULT %s | %s CACHED %s | Execution: %v\n    %s",
			blue, reset, yellow, reset, execTime, string(indentJSON([]byte(result))))
	} else {
		db.logger.Printf("%s RESULT %s | Execution: %v\n    %s",
			blue, reset, execTime, string(indentJSON([]byte(result))))
	}
}

func (db *DB) logError(errMsg string, execTime time.Duration) {
	db.logger.Printf("%s RESULT %s | Execution: %v\n    ERROR: %s",
		blue, reset, execTime, errMsg)
}

var (
	green   = string([]byte{27, 91, 57, 55, 59, 52, 50, 109})
	white   = string([]byte{27, 91, 57, 48, 59, 52, 55, 109})
	yellow  = string([]byte{27, 91, 57, 55, 59, 52, 51, 109})
	red     = string([]byte{27, 91, 57, 55, 59, 52, 49, 109})
	blue    = string([]byte{27, 91, 57, 55, 59, 52, 52, 109})
	magenta = string([]byte{27, 91, 57, 55, 59, 52, 53, 109})
	cyan    = string([]byte{27, 91, 57, 55, 59, 52, 54, 109})
	reset   = string([]byte{27, 91, 48, 109})
)

func indentJSON(in []byte) []byte {
	b := &bytes.Buffer{}
	_ = json.Indent(b, in, "    ", "  ")

	return b.Bytes()
}

func getFullURL(db *DB, path string) string {
	url := bytes.NewBufferString(db.url)
	url.WriteString("/_db/")
	url.WriteString(db.database)
	url.WriteString(path)

	return url.String()
}
