package arangolite

import (
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTransactionRun runs tests on the Transaction Run method.
func TestTransactionRun(t *testing.T) {
	a := assert.New(t)
	r := require.New(t)
	db := New().LoggerOptions(false, false, false)
	db.Connect("http://arangodb:8000", "dbName", "foo", "bar")

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	result, err := NewTransaction([]string{"foo"}, []string{"bar"}).
		AddQuery("var1", "FOR c IN customer RETURN c").Run(db)
	r.Error(err)
	a.Nil(result)

	httpmock.RegisterResponder("POST", "http://arangodb:8000/_db/dbName/_api/transaction",
		func(r *http.Request) (*http.Response, error) {
			buffer, _ := ioutil.ReadAll(r.Body)
			return httpmock.NewStringResponse(200, `{"error": false, "errorMessage": "", "result": [`+string(buffer)+`]}`), nil
		})

	result, err = NewTransaction(nil, nil).Run(nil)
	r.Error(err)
	a.Nil(result)

	result, err = NewTransaction(nil, nil).Run(db)
	r.NoError(err)
	a.Equal("[]", string(result))

	result, err = NewTransaction([]string{"foo"}, []string{"bar"}).
		AddQuery("var1", "FOR c IN customer RETURN c").
		AddQuery("var2", "FOR c IN {{.var1}} RETURN c").
		AddQuery("", "FOR c IN customer").
		Return("var1").Run(db)
	r.NoError(err)
	a.Equal("[{\"collections\":{\"read\":[\"foo\"],\"write\":[\"bar\"]},\"action\":\"function () {var db = require(`internal`).db; var var1 = db._query(aqlQuery`FOR c IN customer RETURN c`).toArray(); var var2 = db._query(aqlQuery`FOR c IN ${var1} RETURN c`).toArray(); db._query(aqlQuery`FOR c IN customer`).toArray(); return var1;}\"}]", string(result))

	httpmock.RegisterResponder("POST", "http://arangodb:8000/_db/dbName/_api/transaction",
		httpmock.NewStringResponder(500, `{"error": true, "errorMessage": "error !"}`))

	result, err = NewTransaction([]string{"foo"}, []string{"bar"}).
		AddQuery("var1", "FOR c IN customer RETURN c").Run(db)
	r.Error(err)
	a.Nil(result)
}
