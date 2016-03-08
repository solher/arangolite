package arangolite

import (
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/h2non/gentleman-mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTransactionRun runs tests on the Transaction Run method.
func TestTransactionRun(t *testing.T) {
	a := assert.New(t)
	r := require.New(t)
	db := New().LoggerOptions(false, false, false)
	db.Connect("http://transaction:8000", "dbName", "foo", "bar")

	db.conn.Use(mock.Plugin)
	defer mock.Disable()
	req := []byte{}
	m := mock.New("http://transaction:8000").Persist().
		Filter(func(re *http.Request) bool {
		req, _ = ioutil.ReadAll(re.Body)
		return true
	})

	result, err := db.Run(NewTransaction([]string{"foo"}, []string{"bar"}).
		AddQuery("var1", "FOR c IN customer RETURN c"))
	r.Error(err)
	a.Nil(result)

	m.Post("/_db/dbName/_api/transaction").
		Reply(200).
		BodyString("{}")

	result, err = db.Run(NewTransaction(nil, nil))
	r.NoError(err)
	a.Equal("{\"collections\":{\"read\":[],\"write\":[]},\"action\":\"function () { var db = require('internal').db; }\"}", string(req))

	result, err = db.Run(NewTransaction([]string{"foo"}, []string{"bar"}).
		AddQuery("", "FOR c IN customer RETURN c"))
	r.NoError(err)
	a.Equal("{\"collections\":{\"read\":[\"foo\"],\"write\":[\"bar\"]},\"action\":\"function () { var db = require('internal').db; db._query(aqlQuery`FOR c IN customer RETURN c`).toArray(); }\"}", string(req))

	result, err = db.Run(NewTransaction([]string{"foo"}, []string{"bar"}).
		AddQuery("var1", "FOR c IN customer RETURN c").
		AddQuery("var2", "FOR c IN {{.var1}} RETURN c").
		Return("var1"))
	r.NoError(err)
	a.Equal("{\"collections\":{\"read\":[\"foo\"],\"write\":[\"bar\"]},\"action\":\"function () { var db = require('internal').db; var var1 = db._query(aqlQuery`FOR c IN customer RETURN c`).toArray(); var var2 = db._query(aqlQuery`FOR c IN ${var1} RETURN c`).toArray(); return var1;}\"}", string(req))

	transaction := NewTransaction([]string{"foo"}, []string{"bar"}).
		AddQuery("var1", "FOR c IN customer FILTER c._key == {{.key}} RETURN c").
		AddQuery("var2", "FOR c IN {{.var1}} RETURN c").
		Return("var2")
	transaction.Bind("key", 123)
	result, err = db.Run(transaction)

	r.NoError(err)
	a.Equal("{\"collections\":{\"read\":[\"foo\"],\"write\":[\"bar\"]},\"action\":\"function () { var db = require('internal').db; var key = 123; var var1 = db._query(aqlQuery`FOR c IN customer FILTER c._key == ${key} RETURN c`).toArray(); var var2 = db._query(aqlQuery`FOR c IN ${var1} RETURN c`).toArray(); return var2;}\"}", string(req))

	transaction = NewTransaction([]string{"foo"}, []string{"bar"}).
		AddQuery("var1", "FOR c IN customer FILTER c._key == @key RETURN c").
		AddQuery("var2", "FOR c IN {{.var1}} RETURN c").
		Return("var2")
	transaction.Bind("key", "123")
	result, err = db.Run(transaction)

	r.NoError(err)
	a.Equal("{\"collections\":{\"read\":[\"foo\"],\"write\":[\"bar\"]},\"action\":\"function () { var db = require('internal').db; var key = '123'; var var1 = db._query(aqlQuery`FOR c IN customer FILTER c._key == ${key} RETURN c`).toArray(); var var2 = db._query(aqlQuery`FOR c IN ${var1} RETURN c`).toArray(); return var2;}\"}", string(req))

	m.Post("/_db/dbName/_api/transaction").
		Reply(500).
		BodyString(`{"error": true, "errorMessage": "error !"}`)

	result, err = db.Run(NewTransaction([]string{"foo"}, []string{"bar"}).
		AddQuery("var1", "FOR c IN customer RETURN c"))
	r.Error(err)
	a.Nil(result)
}
