// Package arangolite provides a lightweight ArangoDatabase driver.
package arangolite

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/solher/arangolite/requests"
)

// Option sets an option for the database connection.
type Option func(db *Database)

// OptHost sets the host and the port used to access the database.
func OptHost(host, port string) Option {
	return func(db *Database) {
		db.host = host
		db.port = port
	}
}

// OptCredentials sets the username and password used to access the database.
func OptCredentials(username, password string) Option {
	return func(db *Database) {
		db.username = username
		db.password = password
	}
}

// OptDatabaseName sets the name of the targeted database.
func OptDatabaseName(dbName string) Option {
	return func(db *Database) {
		db.dbName = dbName
	}
}

// OptHTTPClient sets the HTTP client used to interact with the database.
func OptHTTPClient(cli *http.Client) Option {
	return func(db *Database) {
		if cli != nil {
			db.cli = cli
		}
	}
}

// OptLogging enables logging of the exchanges with the database.
func OptLogging(logger *log.Logger, verbosity LogVerbosity) Option {
	return func(db *Database) {
		if logger != nil {
			db.sender = newLoggingSender(db.sender, logger, verbosity)
		}
	}
}

// Runnable defines requests runnable by the Run and Send methods.
// A Runnable library is located in the 'requests' package.
type Runnable interface {
	// The body of the request.
	Generate() []byte
	// The path where to send the request.
	Path() string
	// The HTTP method to use.
	Method() string
}

// Response defines the response returned by the execution of a Runnable.
type Response interface {
	// The raw response from the database.
	Raw() json.RawMessage
	// The raw response result.
	RawResult() json.RawMessage
	// HasMore indicates if a next result page is available.
	HasMore() bool
	// The cursor ID if more result pages are available.
	Cursor() string
	// Unmarshal decodes the value of the Content field into the given object.
	Unmarshal(v interface{}) error
}

// Database represents an access to an ArangoDB database.
type Database struct {
	host, port         string
	username, password string
	dbName             string
	cli                *http.Client
	sender             sender
}

// NewDatabase returns a new Database object.
func NewDatabase(opts ...Option) *Database {
	db := &Database{
		host:   "localhost",
		port:   "8529",
		dbName: "_system",
		// These Transport parameters are derived from github.com/hashicorp/go-cleanhttp which is under Mozilla Public License.
		cli: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
			},
			Timeout: 10 * time.Minute,
		},
		sender: &basicSender{},
	}

	for _, opt := range opts {
		opt(db)
	}

	return db
}

// Run runs the Runnable, follows the query cursor if needed and unmarshal
// the result in the given object.
func (db *Database) Run(q Runnable, v interface{}) error {
	if q == nil {
		return nil
	}

	r, err := db.Send(q)
	if err != nil {
		return err
	}

	result, err := db.followCursor(r)
	if err != nil {
		return errors.Wrap(err, "could not follow the query cursor")
	}

	return json.Unmarshal(result, v)
}

// Send runs the Runnable and returns a Response that allows the user to
// have more control and handle pagination manually.
func (db *Database) Send(q Runnable) (Response, error) {
	if q == nil {
		return &response{}, nil
	}

	req, err := http.NewRequest(
		q.Method(),
		fmt.Sprintf("http://%s:%s/_db/%s/%s", db.host, db.port, db.dbName, q.Path()),
		bytes.NewBuffer(q.Generate()),
	)
	if err != nil {
		return nil, errors.Wrap(err, "the http request generation failed")
	}

	req.SetBasicAuth(db.username, db.password)

	res, err := db.sender.Send(db.cli, req)
	if err != nil {
		return nil, err
	}
	if res.parsed.Error {
		err = errors.Wrap(errors.New(res.parsed.ErrorMessage), "the database returned an error")
		switch {
		case strings.Contains(res.parsed.ErrorMessage, "unique constraint violated"):
			err = withErrUnique(err)
		case strings.Contains(res.parsed.ErrorMessage, "not found") || strings.Contains(res.parsed.ErrorMessage, "unknown collection"):
			err = withErrNotFound(err)
		case strings.Contains(res.parsed.ErrorMessage, "duplicate name"):
			err = withErrDuplicate(err)
		}
	}
	// We also return the response in the case of a database error so the user
	// can eventually do something with it
	return res, err
}

// followCursor follows the cursor of the given response and returns
// all elements of every batch returned by the database.
func (db *Database) followCursor(r Response) ([]byte, error) {
	// If the result isn't a JSON array, we only return the first batch.
	if len(r.RawResult()) == 0 || r.RawResult()[0] != '[' {
		return r.RawResult(), nil
	}

	q := &requests.FollowCursor{Cursor: r.Cursor()}
	buf := bytes.NewBuffer([]byte{'['})
	var err error

	for r.HasMore() {
		r, err = db.Send(q)
		if err != nil {
			return nil, err
		}
		buf.Write(r.RawResult()[1 : len(r.RawResult())-1])
		buf.WriteRune(',')
	}

	buf.Truncate(buf.Len() - 1)
	buf.WriteRune(']')

	return buf.Bytes(), nil
}
