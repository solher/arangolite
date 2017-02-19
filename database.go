// Package arangolite provides a lightweight ArangoDatabase driver.
package arangolite

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/solher/arangolite/requests"
)

// Option sets an option for the database connection.
type Option func(db *Database)

// Host sets the host and the port used to access the database.
func Host(host, port string) Option {
	return func(db *Database) {
		db.host = host
		db.port = port
	}
}

// Credentials sets the username and password used to access the database.
func Credentials(username, password string) Option {
	return func(db *Database) {
		db.username = username
		db.password = password
	}
}

// DatabaseName sets the name of the targeted database.
func DatabaseName(dbname string) Option {
	return func(db *Database) {
		db.dbname = dbname
	}
}

// HTTPClient sets the HTTP client used to interact with the database.
func HTTPClient(cli *http.Client) Option {
	return func(db *Database) {
		db.cli = cli
	}
}

// Logger sets the logger used by the drive to print debug information.
func Logger(logger *log.Logger) Option {
	return func(db *Database) {
		db.logger = logger
	}
}

// Database represents an access to an ArangoDB database.
type Database struct {
	host, port         string
	username, password string
	dbname             string
	cli                *http.Client
	logger             *log.Logger
}

// NewDatabase returns a new Database object.
func NewDatabase(opts ...Option) *Database {
	db := &Database{
		host:   "localhost",
		port:   "8529",
		dbname: "_system",
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
		logger: log.New(ioutil.Discard, "", 0),
	}

	for _, opt := range opts {
		opt(db)
	}

	return db
}

// Runnable defines requests runnable by the Run and RunAsync methods.
// Queries, transactions and everything in the requests.go file are Runnable.
type Runnable interface {
	// The body of the request.
	Generate() []byte
	// The path where to send the request.
	Path() string
	// The HTTP method to use.
	Method() string
}

// Result defines the result returned by the execution of a Runnable.
type Result interface {
	// The raw answer from the database.
	Raw() json.RawMessage
	// The raw answer content.
	RawContent() json.RawMessage
	// HasMore indicates if a next result page is available.
	HasMore() bool
	// The cursor ID if more result pages are available.
	Cursor() string
	// Unmarshal decodes the value of the Content field into the given object.
	Unmarshal(v interface{}) error
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

// Send runs the Runnable and returns a Result that allows the user to
// have more control and handle pagination manually.
func (db *Database) Send(q Runnable) (Result, error) {
	if q == nil {
		return &result{}, nil
	}

	r, err := db.send(q.Method(), q.Path(), q.Generate())
	if err != nil {
		return nil, err
	}

	return r, nil
}

type parsedResult struct {
	Error        bool            `json:"error"`
	ErrorMessage string          `json:"errorMessage"`
	Content      json.RawMessage `json:"result"`
	Cached       bool            `json:"cached"`
	HasMore      bool            `json:"hasMore"`
	ID           string          `json:"id"`
}

type result struct {
	raw    json.RawMessage
	parsed parsedResult
}

func (r *result) Raw() json.RawMessage {
	return r.raw
}

func (r *result) RawContent() json.RawMessage {
	return r.parsed.Content
}

func (r *result) HasMore() bool {
	return r.parsed.HasMore
}

func (r *result) Cursor() string {
	return r.parsed.ID
}

func (r *result) Unmarshal(v interface{}) error {
	return json.Unmarshal(r.raw, v)
}

func (db *Database) send(method, path string, body []byte) (*result, error) {
	u, err := url.Parse(fmt.Sprintf("http://%s:%s/_db/%s/%s", db.host, db.port, db.dbname, path))
	if err != nil {
		return nil, errors.Wrap(err, "the database URL generation failed")
	}

	req := &http.Request{
		Method: method,
		Body:   ioutil.NopCloser(bytes.NewBuffer(body)),
		URL:    u,
	}
	req.SetBasicAuth(db.username, db.password)

	res, err := db.cli.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "the database HTTP request failed")
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		res.Body.Close()
		return nil, errors.Errorf("the database HTTP request failed, status code %d", res.StatusCode)
	}

	raw, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, errors.Wrap(err, "could not read the database response")
	}
	parsed := parsedResult{}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, errors.Wrap(err, "database response decoding failed")
	}

	if parsed.Error {
		err := errors.Wrap(errors.New(parsed.ErrorMessage), "the database returned an error")
		switch {
		case strings.Contains(parsed.ErrorMessage, "unique constraint violated"):
			err = withErrUnique(err)
		case strings.Contains(parsed.ErrorMessage, "not found") || strings.Contains(parsed.ErrorMessage, "unknown collection"):
			err = withErrNotFound(err)
		case strings.Contains(parsed.ErrorMessage, "duplicate name"):
			err = withErrDuplicate(err)
		}
		return nil, err
	}

	return &result{raw: raw, parsed: parsed}, nil
}

// followCursor follows the cursor of the given result and returns
// all elements of every batch returned by the database.
func (db *Database) followCursor(r Result) ([]byte, error) {
	// If the result isn't a JSON array, we only return the first batch.
	if len(r.RawContent()) == 0 || r.RawContent()[0] != '[' {
		return r.RawContent(), nil
	}

	q := &requests.FollowCursor{Cursor: r.Cursor()}
	buf := bytes.NewBuffer([]byte{'['})
	var err error

	for r.HasMore() {
		r, err = db.Send(q)
		if err != nil {
			return nil, err
		}
		buf.Write(r.RawContent()[1 : len(r.RawContent())-1])
		buf.WriteRune(',')
	}

	buf.Truncate(buf.Len() - 1)
	buf.WriteRune(']')

	return buf.Bytes(), nil
}
