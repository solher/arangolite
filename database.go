// Package arangolite provides a lightweight ArangoDB driver.
package arangolite

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/eapache/go-resiliency/retrier"
	"gopkg.in/h2non/gentleman-retry.v1"
	"gopkg.in/h2non/gentleman.v1"
	"gopkg.in/h2non/gentleman.v1/context"
)

// DB represents an access to an ArangoDB database.
type DB struct {
	url, database, username, password string
	conn                              *gentleman.Client
	l                                 *logger
}

// New returns a new DB object.
func New() *DB {
	db := &DB{l: newLogger()}

	cli := gentleman.New()
	cli.Use(retry.New(retrier.New(retrier.ExponentialBackoff(3, 100*time.Millisecond), nil)))
	cli.UseRequest(func(ctx *context.Context, h context.Handler) {
		u, err := url.Parse(db.url)
		if err != nil {
			h.Error(ctx, err)
			return
		}

		ctx.Request.URL.Scheme = u.Scheme
		ctx.Request.URL.Host = u.Host
		ctx.Request.URL.Path = db.dbPath()
		h.Next(ctx)
	})
	cli.UseRequest(func(ctx *context.Context, h context.Handler) {
		ctx.Request.SetBasicAuth(db.username, db.password)
		h.Next(ctx)
	})

	db.conn = cli

	return db
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
	Description() string // Description shown in the logger
	Generate() []byte    // The body of the request
	Path() string        // The path where to send the request
	Method() string      // The HTTP method to use
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

	c, err := db.send(q.Description(), q.Method(), q.Path(), q.Generate())
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

	url, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	db.l.LogBegin(description, method, db.url+db.dbPath()+path, body)
	start := time.Now()
	path = url.EscapedPath()

	req := db.conn.Request().
		Method(method).
		AddPath(path).
		SetQueryParams(db.queryParams(url))

	if body != nil {
		req.Body(bytes.NewBuffer(body))
	}

	res, err := req.Send()
	if err != nil {
		db.l.LogError(err.Error(), start)
		return nil, err
	}

	if !res.Ok && len(res.Bytes()) == 0 {
		err := errors.New("the database returned a " + strconv.Itoa(res.StatusCode))

		switch res.StatusCode {
		case http.StatusUnauthorized:
			err = errors.New("unauthorized: invalid credentials")
		case http.StatusTemporaryRedirect:
			err = errors.New("the database returned a 307 to " + res.Header.Get("Location"))
		}

		db.l.LogError(err.Error(), start)

		return nil, err
	}

	result := &result{}
	json.Unmarshal(res.Bytes(), result)

	if result.Error {
		db.l.LogError(result.ErrorMessage, start)
		switch {
		case strings.Contains(result.ErrorMessage, "unique constraint violated"):
			return nil, &ErrUnique{result.ErrorMessage}
		case strings.Contains(result.ErrorMessage, "not found"):
			return nil, &ErrNotFound{result.ErrorMessage}
		case strings.Contains(result.ErrorMessage, "unknown collection"):
			return nil, &ErrNotFound{result.ErrorMessage}
		case strings.Contains(result.ErrorMessage, "duplicate name"):
			return nil, &ErrDuplicate{result.ErrorMessage}
		default:
			return nil, errors.New(result.ErrorMessage)
		}
	}

	go db.l.LogResult(result.Cached, start, in, out)

	if len(result.Content) != 0 {
		in <- result.Content
	} else {
		in <- json.RawMessage(res.Bytes())
	}

	if result.HasMore {
		go db.followCursor(path+"/"+result.ID, in)
	} else {
		in <- nil
	}

	return out, nil
}

// followCursor requests the cursor in database, put the result in the channel
// and follow while more batches are available.
func (db *DB) followCursor(path string, c chan interface{}) {
	req := db.conn.Request().
		Method("PUT").
		AddPath(path)

	res, err := req.Send()
	if err != nil {
		c <- err
		return
	}

	result := &result{}
	json.Unmarshal(res.Bytes(), result)

	if result.Error {
		c <- errors.New(result.ErrorMessage)
		return
	}

	c <- result.Content

	if result.HasMore {
		go db.followCursor(path, c)
	} else {
		c <- nil
	}
}

// syncResult synchronises the async result and returns all elements
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

func (db *DB) dbPath() string {
	return "/_db/" + db.database
}

func (db *DB) queryParams(url *url.URL) map[string]string {
	values := url.Query()
	queryParams := map[string]string{}

	for k, v := range values {
		if len(v) > 0 {
			queryParams[k] = v[0]
		}
	}

	return queryParams
}
