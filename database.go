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

type RunnableQuery interface {
	generate() []byte
}

// runQuery executes a query at the path passed as argument.
func (db *DB) runQuery(path string, query RunnableQuery) (chan interface{}, error) {
	if query == nil {
		return nil, errors.New("nil or empty query")
	}

	c := make(chan interface{}, 16)
	fullURL := getFullURL(db, path)

	req, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(query.generate()))
	if err != nil {
		return nil, err
	}

	r, err := db.conn.Do(req)
	if err != nil {
		return nil, err
	}

	result := &Result{}

	_ = json.NewDecoder(r.Body).Decode(result)
	r.Body.Close()

	if result.Error {
		return nil, errors.New(result.ErrorMessage)
	}

	c <- result.Content

	if result.HasMore {
		go db.followCursor(fullURL+"/"+result.ID, c)
	} else {
		c <- nil
	}

	return c, nil
}

func (db *DB) followCursor(url string, c chan interface{}) {
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(nil))
	if err != nil {
		c <- err
		return
	}

	r, err := db.conn.Do(req)
	if err != nil {
		c <- err
		return
	}

	result := &Result{}

	_ = json.NewDecoder(r.Body).Decode(result)
	r.Body.Close()

	if result.Error {
		c <- errors.New(result.ErrorMessage)
		return
	}

	c <- result.Content

	if result.HasMore {
		go db.followCursor(url, c)
	} else {
		c <- nil
	}
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
