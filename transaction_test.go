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

	result, err := db.Run(NewTransaction([]string{"foo"}, []string{"bar"}).
		AddQuery("var1", "FOR c IN customer RETURN c"))
	r.Error(err)
	a.Nil(result)

	httpmock.RegisterResponder("POST", "http://arangodb:8000/_db/dbName/_api/transaction",
		func(r *http.Request) (*http.Response, error) {
			buffer, _ := ioutil.ReadAll(r.Body)
			return httpmock.NewStringResponse(200, string(buffer)), nil
		})

	result, err = db.Run(NewTransaction(nil, nil))
	r.NoError(err)
	a.Equal("{\"collections\":{\"read\":[],\"write\":[]},\"action\":\"function () { var db = require(`internal`).db; }\"}", string(result))

	result, err = db.Run(NewTransaction([]string{"foo"}, []string{"bar"}).
		AddQuery("", "FOR c IN customer RETURN c"))
	r.NoError(err)
	a.Equal("{\"collections\":{\"read\":[\"foo\"],\"write\":[\"bar\"]},\"action\":\"function () { var db = require(`internal`).db; db._query(aqlQuery`FOR c IN customer RETURN c`).toArray(); }\"}", string(result))

	result, err = db.Run(NewTransaction([]string{"foo"}, []string{"bar"}).
		AddQuery("var1", "FOR c IN customer RETURN c").
		AddQuery("var2", "FOR c IN {{.var1}} RETURN c").
		Return("var1"))
	r.NoError(err)
	a.Equal("{\"collections\":{\"read\":[\"foo\"],\"write\":[\"bar\"]},\"action\":\"function () { var db = require(`internal`).db; var var1 = db._query(aqlQuery`FOR c IN customer RETURN c`).toArray(); var var2 = db._query(aqlQuery`FOR c IN ${var1} RETURN c`).toArray(); return var1;}\"}", string(result))

	transaction := NewTransaction([]string{"foo"}, []string{"bar"}).
		AddQuery("var1", "FOR c IN customer FILTER c._key == {{.key}} RETURN c").
		AddQuery("var2", "FOR c IN {{.var1}} RETURN c").
		Return("var2")
	transaction.Bind("key", 123)
	result, err = db.Run(transaction)

	r.NoError(err)
	a.Equal("{\"collections\":{\"read\":[\"foo\"],\"write\":[\"bar\"]},\"action\":\"function () { var db = require(`internal`).db; var key = '123'; var params = {key: key}; var var1 = db._query(aqlQuery`FOR c IN customer FILTER c._key == ${key} RETURN c`, params).toArray(); var var2 = db._query(aqlQuery`FOR c IN ${var1} RETURN c`, params).toArray(); return var2;}\"}", string(result))

	transaction = NewTransaction([]string{"foo"}, []string{"bar"}).
		AddQuery("var1", "FOR c IN customer FILTER c._key == @key RETURN c").
		AddQuery("var2", "FOR c IN {{.var1}} RETURN c").
		Return("var2")
	transaction.Bind("key", 123)
	result, err = db.Run(transaction)

	r.NoError(err)
	a.Equal("{\"collections\":{\"read\":[\"foo\"],\"write\":[\"bar\"]},\"action\":\"function () { var db = require(`internal`).db; var key = '123'; var params = {key: key}; var var1 = db._query(aqlQuery`FOR c IN customer FILTER c._key == @key RETURN c`, params).toArray(); var var2 = db._query(aqlQuery`FOR c IN ${var1} RETURN c`, params).toArray(); return var2;}\"}", string(result))

	httpmock.RegisterResponder("POST", "http://arangodb:8000/_db/dbName/_api/transaction",
		httpmock.NewStringResponder(500, `{"error": true, "errorMessage": "error !"}`))

	result, err = db.Run(NewTransaction([]string{"foo"}, []string{"bar"}).
		AddQuery("var1", "FOR c IN customer RETURN c"))
	r.Error(err)
	a.Nil(result)
}
