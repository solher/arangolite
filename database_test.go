package arangolite

import (
	"encoding/json"
	"testing"

	"github.com/h2non/gentleman-mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConnect runs tests on the arangolite Connect method.
func TestConnect(t *testing.T) {
	db := New().LoggerOptions(false, false, false)
	db.Connect("http://database:8000", "dbName", "foo", "bar").SwitchDatabase("dbName").SwitchUser("foo", "bar")
}

// TestRun runs tests on the arangolite Run and RunAsync methods.
func TestRun(t *testing.T) {
	a := assert.New(t)
	r := require.New(t)
	db := New().LoggerOptions(false, false, false)
	db.Connect("http://database:8000", "dbName", "foo", "bar")

	result, err := db.Run(nil)
	r.NoError(err)
	a.Equal(0, len(result))

	async, err := db.RunAsync(nil)
	r.NoError(err)
	a.Equal(false, async.HasMore())
}

// TestSend runs tests on the arangolite send methods.
func TestSend(t *testing.T) {
	a := assert.New(t)
	r := require.New(t)

	db := New().LoggerOptions(false, false, false)
	db.conn.Use(mock.Plugin)
	defer mock.Disable()
	m := mock.New("http://database:8000").Persist()

	// The connect params are incorrect
	db.Connect("", "", "", "")
	result, err := db.send("TEST", "POST", "/path", []byte(`{"query":"FOR c IN customer RETURN c"}`))
	r.Error(err)
	a.Nil(result)

	// The URL parsing returns an error
	db.Connect("http://[::1]:namedport", "", "", "")
	result, err = db.send("TEST", "POST", "/path", []byte(`{"query":"FOR c IN customer RETURN c"}`))
	r.Error(err)
	a.Nil(result)

	// A database returning an unauthorized status is created
	m.Post("/_db/dbName/path").
		Reply(401).
		BodyString(``)

	// Unauthorized access
	db.Connect("http://database:8000", "dbName", "bar", "foo")
	result, err = db.send("TEST", "POST", "/path", []byte(`{"query":"FOR c IN customer RETURN c"}`))
	r.Error(err)
	r.Contains(err.Error(), "unauthorized")
	a.Nil(result)

	// A valid database returning an empty result is created
	m.Post("/_db/dbName/path").
		Reply(200).
		BodyString(`{"error": false, "errorMessage": "", "result": []}`)

	// The query is empty and succeeds
	db.Connect("http://database:8000", "dbName", "foo", "bar")
	result, err = db.send("TEST", "POST", "/path", []byte{})
	r.NoError(err)
	a.Equal("[]", string((<-result).(json.RawMessage)))

	// The query is nil and succeeds
	result, err = db.send("TEST", "POST", "/path", nil)
	r.NoError(err)
	a.Equal("[]", string((<-result).(json.RawMessage)))

	// The query is not empty and succeeds
	resultByte, err := db.Send("TEST", "POST", "/path", struct{ query string }{query: "FOR c IN customer RETURN c"})
	r.NoError(err)
	a.Equal("[]", string(resultByte))

	// Mashalling fails
	resultByte, err = db.Send("TEST", "POST", "/path", func() {})
	r.Error(err)
	a.Nil(resultByte)

	// A valid database emulating a config response is created
	m.Post("/_db/dbName/path").
		Reply(200).
		BodyString(`{"error": false}`)

	// The query is empty and succeeds
	db.Connect("http://database:8000", "dbName", "foo", "bar")
	result, err = db.send("TEST", "POST", "/path", []byte{})
	r.NoError(err)
	a.Equal("{\"error\": false}", string((<-result).(json.RawMessage)))

	// The query is not empty and succeeds
	resultByte, err = db.Send("TEST", "POST", "/path", struct{ query string }{query: "FOR c IN customer RETURN c"})
	r.NoError(err)
	a.Equal("{\"error\": false}", string(resultByte))

	// A database returning an error is created
	m.Post("/_db/dbName/path").
		Reply(200).
		BodyString(`{"error": true, "errorMessage": "ERROR !"}`)

	// The database error is returned
	result, err = db.send("TEST", "POST", "/path", []byte(`{"query":"FOR c IN customer RETURN c"}`))
	r.Error(err)
	a.Equal("ERROR !", err.Error())
	a.Nil(result)

	resultByte, err = db.Send("TEST", "POST", "/path", struct{ query string }{query: "FOR c IN customer RETURN c"})
	r.Error(err)
	a.Equal("ERROR !", err.Error())
	a.Nil(resultByte)

	// A valid database returning a cursor is created
	m.Post("/_db/dbName/path").
		Reply(200).
		BodyString(`{"error": false, "errorMessage": "", "result": [], "hasMore":true, "id":"1000"}`)

	// send doesn't return error but one is returned in the channel as no responder
	// is listening for the PUT method
	result, err = db.send("TEST", "POST", "/path", []byte(`{"query":"FOR c IN customer RETURN c"}`))
	r.NoError(err)
	a.Equal("[]", string((<-result).(json.RawMessage)))
	a.Error((<-result).(error))

	// The PUT responder is set but returns an error
	m.Post("/_db/dbName/path").
		Reply(200).
		BodyString(`{"error": false, "errorMessage": "", "result": [], "hasMore":true, "id":"1000"}`)

	m2 := mock.New("http://database:8000").Persist()
	m2.Put("/_db/dbName/path/1000").
		Reply(200).
		BodyString(`{"error": true, "errorMessage": "ERROR !", "result": [], "hasMore":false, "id":"1000"}`)

	// The database error is returned in the channel
	result, err = db.send("TEST", "POST", "/path", []byte(`{"query":"FOR c IN customer RETURN c"}`))
	r.NoError(err)
	a.Equal("[]", string((<-result).(json.RawMessage)))
	a.Equal("ERROR !", (<-result).(error).Error())

	// The PUT responder is set and don't returns errors
	m.Post("/_db/dbName/path").
		Reply(200).
		BodyString(`{"error": false, "errorMessage": "", "result": [{}], "hasMore":true, "id":"1000"}`)

	m2.Put("/_db/dbName/path/1000").
		Reply(200).
		BodyString(`{"error": false, "errorMessage": "", "result": [{}], "hasMore":false}`)

	// The database error is returned in the channel
	result, err = db.send("TEST", "POST", "/path", []byte(`{"query":"FOR c IN customer RETURN c"}`))
	r.NoError(err)
	a.Equal("[{}]", string((<-result).(json.RawMessage)))
	a.Equal("[{}]", string((<-result).(json.RawMessage)))
}
