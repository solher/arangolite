package arangolite_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"io/ioutil"

	"strings"

	"reflect"

	"encoding/json"

	"github.com/solher/arangolite"
	"github.com/solher/arangolite/requests"
)

// TestOptionsSend runs tests on the impact of options on the database Send method.
func TestOptionsSend(t *testing.T) {
	logger := log.New(ioutil.Discard, "", 0)
	client, server := httpMock()
	defer server.Close()
	// We add a handler that only return a 200 if all the request parameters are correctly set
	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status := 200
		if username, password, _ := r.BasicAuth(); username != "foo" || password != "bar" {
			status = 500
		}
		if r.URL.String() != fmt.Sprintf("http://foobar:80/_db/foobar/%s", requests.NewAQL("").Path()) {
			status = 500
		}
		w.WriteHeader(status)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, "{}")
	})

	var testCases = []struct {
		// Case description
		description string
		// Arguments
		username, password string
		dbName             string
		endpoint           string
		httpClient         *http.Client
		logger             *log.Logger
		verbosity          arangolite.LogVerbosity
		// Expected results
		testErr func(err error) bool
	}{
		{
			description: "zero/nil parameters",
			username:    "", password: "",
			dbName:     "",
			endpoint:   "",
			httpClient: nil,
			logger:     nil,
			verbosity:  0,
			testErr:    func(err error) bool { return err != nil },
		},
		{
			description: "parameters correctly set",
			username:    "foo", password: "bar",
			dbName:     "foobar",
			endpoint:   "http://foobar:80",
			httpClient: client,
			logger:     logger,
			verbosity:  arangolite.LogDebug,
			testErr:    func(err error) bool { return err == nil },
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			db := arangolite.NewDatabase(
				arangolite.OptBasicAuth(tc.username, tc.password),
				arangolite.OptDatabaseName(tc.dbName),
				arangolite.OptEndpoint(tc.endpoint),
				arangolite.OptHTTPClient(tc.httpClient),
				arangolite.OptLogging(tc.logger, tc.verbosity),
			)
			_, err := db.Send(ctx, requests.NewAQL(""))
			if ok := tc.testErr(err); !ok {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

// TestSend runs tests on the database Send method.
func TestRun(t *testing.T) {
	client, server := httpMock()
	defer server.Close()

	var testCases = []struct {
		// Case description
		description string
		// Arguments
		dbHandler http.HandlerFunc
		// Expected results
		testErr func(err error) bool
		result  interface{}
	}{
		{
			description: "database execution succeeds one page",
			dbHandler:   handler(200, `{"result": [{"_id":"1234"}], "hasMore": false, "cached": true}`),
			testErr:     func(err error) bool { return err == nil },
			result:      []arangolite.Document{{ID: "1234"}},
		},
		{
			description: "database execution succeeds two pages",
			dbHandler: cursorHandler(
				200,
				[]string{
					`{"result": [{"_id":"1234"}], "hasMore": true, "id": "foobar"}`,
					`{"result": [{"_id":"4321"}], "hasMore": false, "id": "foobar"}`,
				},
				"foobar",
			),
			testErr: func(err error) bool { return err == nil },
			result:  []arangolite.Document{{ID: "1234"}, {ID: "4321"}},
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			db := arangolite.NewDatabase(
				arangolite.OptHTTPClient(client),
			)
			server.Config.Handler = tc.dbHandler
			documents := []arangolite.Document{}
			err := db.Run(ctx, requests.NewAQL(""), &documents)
			if ok := tc.testErr(err); !ok {
				t.Errorf("unexpected error: %s", err)
			}
			if !reflect.DeepEqual(tc.result, documents) {
				t.Errorf("unexpected result. Expected %v, got %v", tc.result, documents)
			}
		})
	}
}

// TestSend runs tests on the database Send method.
func TestSend(t *testing.T) {
	client, server := httpMock()
	defer server.Close()

	var testCases = []struct {
		// Case description
		description string
		// Arguments
		dbHandler http.HandlerFunc
		// Expected results
		testErr   func(err error) bool
		raw       json.RawMessage
		rawResult json.RawMessage
		hasMore   bool
		cursor    string
	}{
		{
			description: "database returns a 500",
			dbHandler:   handler(500, "{}"),
			testErr:     func(err error) bool { return err != nil },
			raw:         nil,
		},
		{
			description: "database execution returns an error",
			dbHandler:   handler(200, `{"error": true, "errorMessage": "something happened"}`),
			testErr:     func(err error) bool { return strings.Contains(err.Error(), "something happened") },
			raw:         nil,
		},
		{
			description: "database execution returns a unique constraint error",
			dbHandler:   handler(200, `{"error": true, "errorMessage": "unique constraint violated", "errorNum": 1210}`),
			testErr:     func(err error) bool { return arangolite.IsErrUnique(err) },
			raw:         nil,
		},
		{
			description: "database execution returns a not found error",
			dbHandler:   handler(404, ``),
			testErr:     func(err error) bool { return arangolite.IsErrNotFound(err) },
			raw:         nil,
		},
		{
			description: "database execution returns an unauthorized error",
			dbHandler:   handler(401, ``),
			testErr:     func(err error) bool { return arangolite.IsErrUnauthorized(err) },
			raw:         nil,
		},
		{
			description: "database execution returns a forbidden error",
			dbHandler:   handler(403, ``),
			testErr:     func(err error) bool { return arangolite.IsErrForbidden(err) },
			raw:         nil,
		},
		{
			description: "database execution returns a bad request error",
			dbHandler:   handler(400, ``),
			testErr:     func(err error) bool { return arangolite.IsErrInvalidRequest(err) },
			raw:         nil,
		},
		{
			description: "database execution succeeds one page",
			dbHandler:   handler(200, `{"result": {}, "cached": true}`),
			testErr:     func(err error) bool { return err == nil },
			raw:         json.RawMessage(`{"result": {}, "cached": true}`),
			rawResult:   json.RawMessage(`{}`),
			hasMore:     false,
		},
		{
			description: "database execution succeeds multiple pages",
			dbHandler:   handler(200, `{"content": {}, "hasMore": true, "id": "foobar"}`),
			testErr:     func(err error) bool { return err == nil },
			raw:         json.RawMessage(`{"content": {}, "hasMore": true, "id": "foobar"}`),
			hasMore:     true,
			cursor:      "foobar",
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			db := arangolite.NewDatabase(
				arangolite.OptHTTPClient(client),
			)
			server.Config.Handler = tc.dbHandler
			result, err := db.Send(ctx, requests.NewAQL(""))
			if ok := tc.testErr(err); !ok {
				t.Errorf("unexpected error: %s", err)
			}
			if err != nil {
				return
			}
			if !reflect.DeepEqual(tc.raw, result.Raw()) {
				t.Errorf("unexpected raw result. Expected %v, got %v", string(tc.raw), string(result.Raw()))
			}
			if !reflect.DeepEqual(tc.rawResult, result.RawResult()) {
				t.Errorf("unexpected raw content. Expected %v, got %v", string(tc.rawResult), string(result.RawResult()))
			}
			if !reflect.DeepEqual(tc.hasMore, result.HasMore()) {
				t.Errorf("unexpected hasMore. Expected %v, got %v", tc.hasMore, result.HasMore())
			}
			if !reflect.DeepEqual(tc.cursor, result.Cursor()) {
				t.Errorf("unexpected cursor. Expected %v, got %v", tc.cursor, result.Cursor())
			}
		})
	}
}

func httpMock() (*http.Client, *httptest.Server) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}
	client := &http.Client{Transport: transport}
	return client, server
}

func handler(status int, body string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, body)
	})
}

func cursorHandler(status int, bodies []string, cursor string) http.HandlerFunc {
	i := 0
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if i > 0 {
			if strings.HasSuffix(r.URL.String(), "cursor/"+cursor) && r.Method == "PUT" {
				w.WriteHeader(200)
			} else {
				w.WriteHeader(404)
			}
		} else {
			w.WriteHeader(status)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, bodies[i])
		i++
	})
}
