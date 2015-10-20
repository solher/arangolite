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
	url, database, username, password string
	conn                              *http.Client
	l                                 *logger
}

// New returns a new DB object.
func New() *DB {
	return &DB{conn: &http.Client{}, l: newLogger()}
}

// LoggerOptions sets the Arangolite logger options.
func (db *DB) LoggerOptions(enabled, printQuery, printResult bool) *DB {
	db.l.Options(enabled, printQuery, printResult)
	return db
}

// Connect initialize a DB object with the database url and credentials.
func (db *DB) Connect(url, database, username, password string) *DB {
	db.url = url
	db.database = database
	db.username = username
	db.password = password

	return db
}

// Run runs the Runnable synchronously and returns the JSON array of all elements
// of every batch returned by the database.
func (db *DB) Run(q Runnable) ([]byte, error) {
	if q == nil {
		return []byte{}, nil
	}

	r, err := db.RunAsync(q)
	if err != nil {
		return nil, err
	}

	return db.syncResult(r), nil
}

// RunAsync runs the Runnable asynchronously and returns an async Result object.
func (db *DB) RunAsync(q Runnable) (*Result, error) {
	if q == nil {
		return NewResult(nil), nil
	}

	c, err := db.runQuery(q)

	if err != nil {
		return nil, err
	}

	return NewResult(c), nil
}

// Runnable defines queries runnable by the "runQuery" method.
type Runnable interface {
	description() string
	generate() []byte
	path() string
}

// runQuery executes a query at the path passed as argument.
// It returns a channel where the extracted content of each batch is returned.
func (db *DB) runQuery(q Runnable) (chan interface{}, error) {
	if q == nil {
		return nil, errors.New("nil or empty query")
	}

	in := make(chan interface{}, 16)
	out := make(chan interface{}, 16)
	fullURL := getFullURL(db, q.path())
	jsonQuery := q.generate()

	db.l.LogBegin(q.description(), fullURL, jsonQuery)
	start := time.Now()

	req, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(jsonQuery))
	if err != nil {
		db.l.LogError(err.Error(), start)
		return nil, err
	}

	req.SetBasicAuth(db.username, db.password)

	r, err := db.conn.Do(req)
	if err != nil {
		db.l.LogError(err.Error(), start)
		return nil, err
	}
	if r.StatusCode == http.StatusUnauthorized {
		err = errors.New("unauthorized: invalid credentials")
		db.l.LogError(err.Error(), start)
		return nil, err
	}

	result := &result{}
	json.NewDecoder(r.Body).Decode(result)
	r.Body.Close()

	if result.Error {
		db.l.LogError(result.ErrorMessage, start)
		return nil, errors.New(result.ErrorMessage)
	}

	go db.l.LogResult(result, start, in, out)

	in <- result.Content

	if result.HasMore {
		go db.followCursor(fullURL+"/"+result.ID, q, in)
	} else {
		in <- nil
	}

	return out, nil
}

// followCursor requests the cursor in database, put the result in the channel
// and follow while more batches are available.
func (db *DB) followCursor(url string, q Runnable, c chan interface{}) {
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(nil))

	r, err := db.conn.Do(req)
	if err != nil {
		c <- err
		return
	}

	result := &result{}
	json.NewDecoder(r.Body).Decode(result)
	r.Body.Close()

	if result.Error {
		c <- errors.New(result.ErrorMessage)
		return
	}

	c <- result.Content

	if result.HasMore {
		go db.followCursor(url, q, c)
	} else {
		c <- nil
	}
}

// syncResult	synchronise the async result and return a JSON array of all elements
// of every batch returned by the database.
func (db *DB) syncResult(async *Result) []byte {
	result := []byte{'['}
	r := async.Buffer()

	for async.HasMore() {
		if r.Len() == 0 {
			continue
		}
		result = append(result, r.Bytes()[1:r.Len()-1]...)
		result = append(result, ',')
	}

	if len(result) <= 1 {
		return []byte{'[', ']'}
	}

	result = append(result[:len(result)-1], ']')

	return result
}

func getFullURL(db *DB, path string) string {
	url := bytes.NewBufferString(db.url)
	url.WriteString("/_db/")
	url.WriteString(db.database)
	url.WriteString(path)
	return url.String()
}
