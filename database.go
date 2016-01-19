// Package arangolite provides a lightweight ArangoDB driver.
package arangolite

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
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

// SwitchDatabase change the current database.
func (db *DB) SwitchDatabase(database string) *DB {
	db.database = database
	return db
}

// SwitchUser change the current user.
func (db *DB) SwitchUser(username, password string) *DB {
	db.username = username
	db.password = password
	return db
}

// Runnable defines requests runnable by the Run and RunAsync methods.
// Queries, transactions and everything in the requests.go file are Runnable.
type Runnable interface {
	description() string // Description shown in the logger
	generate() []byte    // The body of the request
	path() string        // The path where to send the request
	method() string      // The HTTP method to use
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

	c, err := db.send(q.description(), q.method(), q.path(), q.generate())

	if err != nil {
		return nil, err
	}

	return NewResult(c), nil
}

// Send runs a low level request in the database.
// The description param is shown in the logger.
// The req param is serialized in the body.
// The purpose of this method is to be a fallback in case the user wants to do
// something which is not implemented in the requests.go file.
func (db *DB) Send(description, method, path string, req interface{}) ([]byte, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	c, err := db.send(description, method, path, body)
	if err != nil {
		return nil, err
	}

	return db.syncResult(NewResult(c)), nil
}

// send executes a request at the path passed as argument.
// It returns a channel where the extracted content of each batch is returned.
func (db *DB) send(description, method, path string, body []byte) (chan interface{}, error) {
	in := make(chan interface{}, 16)
	out := make(chan interface{}, 16)
	fullURL := getFullURL(db, path)

	db.l.LogBegin(description, method, fullURL, body)
	start := time.Now()

	req, err := http.NewRequest(method, fullURL, bytes.NewBuffer(body))
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

	rawResult, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()

	if (r.StatusCode < 200 || r.StatusCode > 299) && len(rawResult) == 0 {
		err = errors.New("the database returned a " + strconv.Itoa(r.StatusCode))

		switch r.StatusCode {
		case http.StatusUnauthorized:
			err = errors.New("unauthorized: invalid credentials")
		case http.StatusTemporaryRedirect:
			err = errors.New("the database returned a 307 to " + r.Header.Get("Location"))
		}

		db.l.LogError(err.Error(), start)

		return nil, err
	}

	result := &result{}
	json.Unmarshal(rawResult, result)

	if result.Error {
		db.l.LogError(result.ErrorMessage, start)
		return nil, errors.New(result.ErrorMessage)
	}

	go db.l.LogResult(result.Cached, start, in, out)

	if len(result.Content) != 0 {
		in <- result.Content
	} else {
		in <- json.RawMessage(rawResult)
	}

	if result.HasMore {
		go db.followCursor(fullURL+"/"+result.ID, in)
	} else {
		in <- nil
	}

	return out, nil
}

// followCursor requests the cursor in database, put the result in the channel
// and follow while more batches are available.
func (db *DB) followCursor(url string, c chan interface{}) {
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(nil))
	req.SetBasicAuth(db.username, db.password)

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
		go db.followCursor(url, c)
	} else {
		c <- nil
	}
}

// syncResult	synchronises the async result and returns all elements
// of every batch returned by the database.
func (db *DB) syncResult(async *Result) []byte {
	r := async.Buffer()
	async.HasMore()

	// If the result isn't a JSON array, we only returns the first batch.
	if r.Bytes()[0] != '[' {
		return r.Bytes()
	}

	// If the result is a JSON array, we try to concatenate them all.
	result := []byte{'['}
	result = append(result, r.Bytes()[1:r.Len()-1]...)
	result = append(result, ',')

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
