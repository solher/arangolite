// Package arangolite provides a lightweight ArangoDatabase driver.
package arangolite

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"runtime"
	"time"

	"github.com/pkg/errors"
	"github.com/solher/arangolite/requests"
)

// Option sets an option for the database connection.
type Option func(db *Database)

// OptEndpoint sets the endpoint used to access the database.
func OptEndpoint(endpoint string) Option {
	return func(db *Database) {
		db.endpoint = endpoint
	}
}

// OptBasicAuth sets the username and password used to access the database
// using basic authentication.
func OptBasicAuth(username, password string) Option {
	return func(db *Database) {
		db.auth = &basicAuth{username: username, password: password}
	}
}

// OptJWTAuth sets the username and password used to access the database
// using JWT authentication.
func OptJWTAuth(username, password string) Option {
	return func(db *Database) {
		db.auth = &jwtAuth{username: username, password: password}
	}
}

// OptDatabaseName sets the name of the targeted database.
func OptDatabaseName(dbName string) Option {
	return func(db *Database) {
		db.dbName = dbName
	}
}

// OptHTTPClient sets the HTTP client used to interact with the database.
// It is also the current solution to set a custom TLS config.
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
	// The raw response result, if present.
	RawResult() json.RawMessage
	// The response HTTP status code.
	StatusCode() int
	// HasMore indicates if a next result page is available.
	HasMore() bool
	// The cursor ID if more result pages are available.
	Cursor() string
	// Unmarshal decodes the response into the given object.
	Unmarshal(v interface{}) error
	// UnmarshalResult decodes the value of the Result field into the given object, if present.
	UnmarshalResult(v interface{}) error
}

// Database represents an access to an ArangoDB database.
type Database struct {
	endpoint string
	dbName   string
	cli      *http.Client
	sender   sender
	auth     authentication
}

// NewDatabase returns a new Database object.
func NewDatabase(opts ...Option) *Database {
	db := &Database{
		endpoint: "http://localhost:8529",
		dbName:   "_system",
		// These Transport parameters are derived from github.com/hashicorp/go-cleanhttp which is under Mozilla Public License.
		cli: &http.Client{
			Transport: &http.Transport{
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
		auth:   &basicAuth{},
	}

	db.Options(opts...)

	return db
}

// Connect setups the database connection and check the connectivity.
func (db *Database) Connect(ctx context.Context) error {
	if err := db.auth.Setup(ctx, db); err != nil {
		return err
	}
	if _, err := db.Send(ctx, &requests.CurrentDatabase{}); err != nil {
		return err
	}
	return nil
}

// Options apply options to the database.
func (db *Database) Options(opts ...Option) {
	for _, opt := range opts {
		opt(db)
	}
}

// Run runs the Runnable, follows the query cursor if any and unmarshal
// the result in the given object.
func (db *Database) Run(ctx context.Context, v interface{}, q Runnable) error {
	if q == nil {
		return nil
	}

	r, err := db.Send(ctx, q)
	if err != nil {
		return err
	}

	result, err := db.followCursor(ctx, r)
	if err != nil {
		return errors.Wrap(err, "could not follow the query cursor")
	}
	if v == nil || result == nil || len(result) == 0 {
		return nil
	}
	if err := json.Unmarshal(result, v); err != nil {
		return errors.Wrap(err, "run result unmarshalling failed")
	}

	return nil
}

// Send runs the Runnable and returns a "raw" Response object.
func (db *Database) Send(ctx context.Context, q Runnable) (Response, error) {
	if q == nil {
		return &response{}, nil
	}

	req, err := http.NewRequest(
		q.Method(),
		fmt.Sprintf("%s/_db/%s%s", db.endpoint, db.dbName, q.Path()),
		bytes.NewBuffer(q.Generate()),
	)
	if err != nil {
		return nil, errors.Wrap(err, "the http request generation failed")
	}

	if err := db.auth.Apply(req); err != nil {
		return nil, errors.Wrap(err, "authentication returned an error")
	}

	res, err := db.sender.Send(ctx, db.cli, req)
	if err != nil {
		return nil, err
	}
	if res.parsed.Error {
		err = errors.Wrap(errors.New(res.parsed.ErrorMessage), "the database execution returned an error")
		err = withErrorNum(err, res.parsed.ErrorNum)
	}
	if res.statusCode < 200 || res.statusCode >= 300 {
		if err == nil {
			err = errors.Errorf("the database HTTP request failed: status code %d", res.statusCode)
		}
		err = withStatusCode(err, res.statusCode)
	}
	if err != nil {
		// We also return the response in the case of a database error so the user
		// can eventually do something with it
		return res, err
	}

	return res, nil
}

// followCursor follows the cursor of the given response and returns
// all elements of every batch returned by the database.
func (db *Database) followCursor(ctx context.Context, r Response) ([]byte, error) {
	// If the result only has one page
	if !r.HasMore() {
		if len(r.RawResult()) != 0 {
			// Parsed result is not empty, so return this
			return r.RawResult(), nil
		} else {
			// Return the raw result
			return r.Raw(), nil
		}
	}

	buf := bytes.NewBuffer(r.RawResult()[:len(r.RawResult())-1])
	buf.WriteRune(',')

	q := &requests.FollowCursor{Cursor: r.Cursor()}
	var err error

	for r.HasMore() {
		r, err = db.Send(ctx, q)
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
