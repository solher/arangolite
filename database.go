// Package arangolite provides a lightweight ArangoDB driver.
package arangolite

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

// DB represents an access to an ArangoDB database.
type DB struct {
	url, database, user, password string
	conn                          *http.Client
	l                             *logger
}

// New returns a new DB object.
func New(logEnabled bool) *DB {
	return &DB{conn: &http.Client{}, l: newLogger(logEnabled)}
}

// Connect initialize a DB object with the database url and credentials.
func (db *DB) Connect(url, database, user, password string) {
	db.url = url
	db.database = database
	db.user = user
	db.password = password
}

type RunnableQuery interface {
	description() string
	generate() []byte
}

// runQuery executes a query at the path passed as argument.
func (db *DB) runQuery(path string, query RunnableQuery) (chan interface{}, error) {
	if query == nil {
		return nil, errors.New("nil or empty query")
	}

	in := make(chan interface{}, 16)
	out := make(chan interface{}, 16)
	fullURL := getFullURL(db, path)
	jsonQuery := query.generate()

	db.l.logBegin(query.description(), fullURL, jsonQuery)
	start := time.Now()

	req, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(jsonQuery))
	if err != nil {
		db.l.logError(err.Error(), time.Now().Sub(start))
		return nil, err
	}

	r, err := db.conn.Do(req)
	if err != nil {
		db.l.logError(err.Error(), time.Now().Sub(start))
		return nil, err
	}

	result := &result{}

	_ = json.NewDecoder(r.Body).Decode(result)
	r.Body.Close()

	if result.Error {
		db.l.logError(result.ErrorMessage, time.Now().Sub(start))
		return nil, errors.New(result.ErrorMessage)
	}

	go db.l.logResult(result, start, in, out)

	in <- result.Content

	if result.HasMore {
		go db.followCursor(fullURL+"/"+result.ID, in)
	} else {
		in <- nil
	}

	return out, nil
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

	result := &result{}

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

func getFullURL(db *DB, path string) string {
	url := bytes.NewBufferString(db.url)
	url.WriteString("/_db/")
	url.WriteString(db.database)
	url.WriteString(path)
	return url.String()
}
