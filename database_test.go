package arangolite

import (
	"encoding/json"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type query struct {
	c string
}

func (q *query) description() string {
	return "TEST"
}

func (q *query) path() string {
	return "/path"
}

func (q *query) generate() []byte {
	return []byte(q.c)
}

// TestConnect runs tests on the arangolite Connect method.
func TestConnect(t *testing.T) {
	db := New().LoggerOptions(false, false, false)
	db.Connect("http://localhost:8000", "dbName", "foo", "bar")
}

// TestRun runs tests on the arangolite Run and RunAsync methods.
func TestRun(t *testing.T) {
	a := assert.New(t)
	r := require.New(t)
	db := New().LoggerOptions(false, false, false)
	db.Connect("http://localhost:8000", "dbName", "foo", "bar")

	result, err := db.Run(nil)
	r.NoError(err)
	a.Equal(0, len(result))

	async, err := db.RunAsync(nil)
	r.NoError(err)
	a.Equal(false, async.HasMore())
}

// TestRunQuery runs tests on the arangolite runQuery method.
func TestRunQuery(t *testing.T) {
	a := assert.New(t)
	r := require.New(t)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// The connect params are incorrect
	db := New().LoggerOptions(false, false, false)
	db.Connect("", "", "", "")
	result, err := db.runQuery(&query{c: `{"query":"FOR c IN customer RETURN c"}`})
	r.Error(err)
	a.Nil(result)

	// The URL parsing returns an error
	db.Connect("http://[::1]:namedport", "dbName", "foo", "bar")
	result, err = db.runQuery(&query{c: `{"query":"FOR c IN customer RETURN c"}`})
	r.Error(err)
	a.Nil(result)

	// A valid database returning an empty result is created
	setValidResponder()

	// // The query is empty
	db.Connect("http://arangodb:8000", "dbName", "foo", "bar")
	result, err = db.runQuery(&query{})
	r.NoError(err)
	a.Equal("[]", string((<-result).(json.RawMessage)))

	// The query can't be nil
	result, err = db.runQuery(nil)
	r.Error(err)
	a.Nil(result)

	// A database returning an error is created
	setErrorResponder()

	// The database error is returned
	result, err = db.runQuery(&query{c: `{"query":"FOR c IN customer RETURN c"}`})
	r.Error(err)
	a.Equal("ERROR !", err.Error())
	a.Nil(result)

	// A valid database returning a cursor is created
	setHasMoreResponder()

	// runQuery doesn't return error but one is returned in the channel as no responder
	// is listening for the PUT method
	result, err = db.runQuery(&query{c: `{"query":"FOR c IN customer RETURN c"}`})
	r.NoError(err)
	a.Equal("[]", string((<-result).(json.RawMessage)))
	a.Error((<-result).(error))

	// The PUT responder is set but returns an error
	setHasMoreResponderPutError()

	// The database error is returned in the channel
	result, err = db.runQuery(&query{c: `{"query":"FOR c IN customer RETURN c"}`})
	r.NoError(err)
	a.Equal("[]", string((<-result).(json.RawMessage)))
	a.Equal("ERROR !", (<-result).(error).Error())

	// The PUT responder is set and don't returns errors
	setHasMoreResponderPutValid()

	// The database error is returned in the channel
	result, err = db.runQuery(&query{c: `{"query":"FOR c IN customer RETURN c"}`})
	r.NoError(err)
	a.Equal("[]", string((<-result).(json.RawMessage)))
	a.Equal("[]", string((<-result).(json.RawMessage)))
}

func setValidResponder() {
	httpmock.RegisterResponder("POST", "http://arangodb:8000/_db/dbName/path",
		httpmock.NewStringResponder(200, `{"error": false, "errorMessage": "", "result": []}`))
}

func setErrorResponder() {
	httpmock.RegisterResponder("POST", "http://arangodb:8000/_db/dbName/path",
		httpmock.NewStringResponder(200, `{"error": true, "errorMessage": "ERROR !"}`))
}

func setHasMoreResponder() {
	httpmock.RegisterResponder("POST", "http://arangodb:8000/_db/dbName/path",
		httpmock.NewStringResponder(200, `{"error": false, "errorMessage": "", "result": [], "hasMore":true, "id":"1000"}`))
}

func setHasMoreResponderPutError() {
	httpmock.RegisterResponder("POST", "http://arangodb:8000/_db/dbName/path",
		httpmock.NewStringResponder(200, `{"error": false, "errorMessage": "", "result": [], "hasMore":true, "id":"1000"}`))

	httpmock.RegisterResponder("PUT", "http://arangodb:8000/_db/dbName/path/1000",
		httpmock.NewStringResponder(200, `{"error": true, "errorMessage": "ERROR !", "result": [], "hasMore":false, "id":"1000"}`))
}

func setHasMoreResponderPutValid() {
	httpmock.RegisterResponder("POST", "http://arangodb:8000/_db/dbName/path",
		httpmock.NewStringResponder(200, `{"error": false, "errorMessage": "", "result": [], "hasMore":true, "id":"1000"}`))

	httpmock.RegisterResponder("PUT", "http://arangodb:8000/_db/dbName/path/1000",
		httpmock.NewStringResponder(200, `{"error": false, "errorMessage": "", "result": [], "hasMore":false}`))
}
