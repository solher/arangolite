// Package arangolite provides a lightweight ArangoDB driver.
package arangolite

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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

	db.logger.Printf("%s QUERY %s\n    %s", blue, reset, indentJSON(query))

	start := time.Now()
	r, err := db.conn.Post(db.url+"/_db/"+db.database+path, "application/json", bytes.NewBuffer(query))
	end := time.Now()

	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	result := &QueryResult{}

	if err := json.NewDecoder(r.Body).Decode(result); err != nil {
		return nil, err
	}

	resultLog := fmt.Sprintf("%s RESULT %s | Execution: %v\n    ",
		blue, reset, end.Sub(start))

	if result.Error {
		db.logger.Printf("%sERROR: %s", resultLog, result.ErrorMessage)
		return nil, errors.New(result.ErrorMessage)
	}

	db.logger.Printf(resultLog + string(indentJSON([]byte(result.Content))))

	return result.Content, nil
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
