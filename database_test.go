package arangolite

import (
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConnect runs tests on the arangolite Connect method.
func TestConnect(t *testing.T) {
	db := New(true)
	db = New(false)
	db.Connect("http://localhost:8000", "dbName", "foo", "bar")
}

// TestRunQuery runs tests on the arangolite runQuery method.
func TestRunQuery(t *testing.T) {
	a := assert.New(t)
	r := require.New(t)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	db := New(false)
	db.Connect("", "", "", "")
	result, err := db.runQuery("/path", []byte(`{"query":"FOR c IN customer RETURN c"}`))
	r.Error(err)
	a.Nil(result)

	httpmock.RegisterResponder("POST", "http://arangodb:8000/_db/dbName/path",
		httpmock.NewStringResponder(200, `{"error": false, "errorMessage": "", "result": "[]"}`))

	db.Connect("http://arangodb:8000", "dbName", "foo", "bar")
	result, err = db.runQuery("/path", []byte{})
	r.Error(err)
	a.Nil(result)

	result, err = db.runQuery("/path", nil)
	r.Error(err)
	a.Nil(result)

	result, err = db.runQuery("/path", []byte(`{"query":"FOR c IN customer RETURN c"}`))
	r.NoError(err)
	a.Equal(`"[]"`, string(result))

	httpmock.RegisterResponder("POST", "http://arangodb:8000/_db/dbName/path",
		httpmock.NewStringResponder(200, `{"error": false, "errorMessage": "", "result": "[]"`))

	result, err = db.runQuery("/path", []byte(`{"query":"FOR c IN customer RETURN c"}`))
	r.Error(err)
	a.Nil(result)

	httpmock.RegisterResponder("POST", "http://arangodb:8000/_db/dbName/path",
		httpmock.NewStringResponder(500, `{"error": true, "errorMessage": "error !"}`))

	result, err = db.runQuery("/path", []byte(`{"query":"FOR c IN customer RETURN c"}`))
	r.Error(err)
	a.Nil(result)
}
