package arangolite_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"strings"

	"github.com/solher/arangolite"
	"github.com/solher/arangolite/requests"
)

// TestConnect runs tests on the database Connect method.
func TestConnect(t *testing.T) {
	client, server := httpMock()
	defer server.Close()

	var testCases = []struct {
		// Case description
		description string
		// Arguments
		dbHandler http.HandlerFunc
		auth      arangolite.Option
		// Expected results
		testErr func(err error) bool
	}{
		{
			description: "database returns a 401",
			dbHandler:   connectHandler(0, ``),
			auth:        arangolite.OptBasicAuth("foo", "invalid"),
			testErr:     func(err error) bool { return arangolite.IsErrUnauthorized(err) },
		},
		{
			description: "database returns a 200",
			dbHandler:   connectHandler(0, ``),
			auth:        arangolite.OptBasicAuth("foo", "bar"),
			testErr:     func(err error) bool { return err == nil },
		},
		{
			description: "jwt login fails",
			dbHandler:   connectHandler(401, ``),
			auth:        arangolite.OptJWTAuth("foo", "bar"),
			testErr:     func(err error) bool { return arangolite.IsErrUnauthorized(err) },
		},
		{
			description: "database returns a 200 for jwt",
			dbHandler:   connectHandler(200, ``),
			auth:        arangolite.OptJWTAuth("foo", "bar"),
			testErr:     func(err error) bool { return err == nil },
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			server.Config.Handler = tc.dbHandler
			db := arangolite.NewDatabase(
				arangolite.OptHTTPClient(client),
				tc.auth,
			)
			err := db.Connect(ctx)
			if ok := tc.testErr(err); !ok {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

// TestOptionsSend runs tests on the impact of options on the database Send method.
func TestOptionsSend(t *testing.T) {
	var logger arangolite.Logger
	logger = log.New(ioutil.Discard, "", 0)
	client, server := httpMock()
	defer server.Close()
	// We add a handler that only return a 200 if all the request parameters are correctly set
	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status := 200
		if username, password, _ := r.BasicAuth(); username != "foo" || password != "bar" {
			status = 500
		}
		if r.URL.String() != fmt.Sprintf("http://foobar:80/_db/foobar%s", requests.NewAQL("").Path()) {
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
		logger             arangolite.Logger
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

// TestRun runs tests on the database Run method.
func TestRun(t *testing.T) {
	client, server := httpMock()
	defer server.Close()

	var testCases = []struct {
		// Case description
		description string
		// Query to pass to Run()
		query arangolite.Runnable
		// Where to store the result
		result interface{}
		// Arguments
		dbHandler http.HandlerFunc
		// Expected results
		testErr func(err error) bool
		expectedResult interface{}
	}{
		{
			description: "database execution succeeds one page",
			query: requests.NewAQL(""),
			result: &[]arangolite.Document{},
			dbHandler:   handler(200, `{"result": [{"_id":"1234"}], "hasMore": false, "cached": true}`),
			testErr:     func(err error) bool { return err == nil },
			expectedResult: &[]arangolite.Document{{ID: "1234"}},
		},
		{
			description: "database execution succeeds two pages",
			query: requests.NewAQL(""),
			result: &[]arangolite.Document{},
			dbHandler: cursorHandler(
				200,
				[]string{
					`{"result": [{"_id":"1234"}], "hasMore": true, "id": "foobar"}`,
					`{"result": [{"_id":"4321"}], "hasMore": false, "id": "foobar"}`,
				},
				"foobar",
			),
			testErr: func(err error) bool { return err == nil },
			expectedResult:  &[]arangolite.Document{{ID: "1234"}, {ID: "4321"}},
		},
		{
			description: "database execution test status code and error num",
			query: requests.NewAQL(""),
			result: &[]arangolite.Document{},
			dbHandler: handlerContentType(
				404,
				`{"error":true,"code":404,"errorNum":1203,"errorMessage":"unknown collection 'items'"}`,
				"application/json",
			),
			testErr: func(err error) bool {
				return arangolite.HasStatusCode(err, 404) && arangolite.HasErrorNum(err, 1203)
			},
			expectedResult:  &[]arangolite.Document{},
		},
		{
			description: "database execution requests.GetVersion",
			query: &requests.GetVersion{Details: false},
			result: &requests.GetVersionResult{},
			dbHandler: handler(200, `{"server":"arango","version":"3.0.12"}`),
			testErr: func(err error) bool { return err == nil },
			expectedResult: &requests.GetVersionResult{Server: "arango", Version: "3.0.12"},
		},
		{
			description: "database execution requests.GetVersion detailed",
			query: &requests.GetVersion{Details: true},
			result: &requests.GetVersionResult{},
			dbHandler: handler(
				200,
				`{"server":"arango","version":"3.0.12","details":{"architecture":"64bit","endianness":"little","sizeof int":"4","sizeof void*":"8","sse42":"false","mode":"server"}}`,
			),
			testErr: func(err error) bool { return err == nil },
			expectedResult: &requests.GetVersionResult{
				Server: "arango",
				Version: "3.0.12",
				Details: map[string]string{"architecture":"64bit","endianness":"little",
					"sizeof int":"4","sizeof void*":"8","sse42":"false","mode":"server"}},
		},
		{
			description: "database execution empty parsed result 1",
			query: requests.NewAQL(""),
			result: &struct {ID string}{ID: ""},
			dbHandler: handler(
				200,
				`{"id":"24292","name":"items","waitForSync":false,"isVolatile":false,"isSystem":false,"status":3,"type":2,"error":false,"code":200}`,
			),
			testErr: func(err error) bool { return err == nil },
			expectedResult: &struct {ID string}{ID: "24292"},
		},
		{
			description: "database execution empty parsed result 3",
			query: requests.NewAQL(""),
			result: &struct {ID string}{ID: ""},
			dbHandler: handler(
				200,
				`{"id":"node_port_relations/365","type":"hash","fields":["property"],"selectivityEstimate":0.03125,"unique":false,"sparse":false,"isNewlyCreated":false,"error":false,"code":200}`,
			),
			testErr: func(err error) bool { return err == nil },
			expectedResult: &struct {ID string}{ID: "node_port_relations/365"},
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			server.Config.Handler = tc.dbHandler
			db := arangolite.NewDatabase(arangolite.OptHTTPClient(client))
			err := db.Run(ctx, tc.result, tc.query)

			if ok := tc.testErr(err); !ok {
				t.Errorf("unexpected error: %s", err)
			}
			if !reflect.DeepEqual(tc.result, tc.expectedResult) {
				t.Errorf("unexpected result. Expected %v, got %v", tc.expectedResult, tc.result)
			}
		})
	}
}

// TestSend runs tests on the database Send method.
func TestSend(t *testing.T) {
	client, server := httpMock()
	defer server.Close()
	var logger arangolite.Logger
	logger = log.New(ioutil.Discard, "", 0)

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
			testErr:     func(err error) bool { return err != nil && strings.Contains(err.Error(), "something happened") },
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
			server.Config.Handler = tc.dbHandler
			db := arangolite.NewDatabase(
				arangolite.OptHTTPClient(client),
				arangolite.OptLogging(logger, arangolite.LogDebug),
			)
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
		if status == 200 {
			w.Header().Set("Content-Type", "application/json")
		}
		w.WriteHeader(status)
		fmt.Fprintln(w, body)
	})
}

func handlerContentType(status int, body string, contentType string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(status)
		fmt.Fprintln(w, body)
	})
}

func cursorHandler(status int, bodies []string, cursor string) http.HandlerFunc {
	i := 0
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if i > 0 {
			if strings.HasSuffix(r.URL.String(), "cursor/"+cursor) && r.Method == "PUT" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(200)
			} else {
				w.WriteHeader(404)
			}
		} else {
			if status == 200 {
				w.Header().Set("Content-Type", "application/json")
			}
			w.WriteHeader(status)
		}
		fmt.Fprintln(w, bodies[i])
		i++
	})
}

func connectHandler(jwtStatus int, body string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, (&requests.JWTAuth{}).Path()) {
			w.WriteHeader(jwtStatus)
			w.Write([]byte(`{"jwt":"foobar"}`))
			return
		}
		if strings.Contains(r.URL.Path, (&requests.CurrentDatabase{}).Path()) {
			if user, pass, ok := r.BasicAuth(); ok {
				if user == "foo" && pass == "bar" {
					w.WriteHeader(200)
					return
				}
			}
			if h := r.Header.Get("Authorization"); h == "bearer foobar" {
				w.WriteHeader(200)
				return
			}
			w.WriteHeader(401)
			return
		}
	})
}
