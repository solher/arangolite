package arangolite

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConnect runs tests on the arangolite Connect method.
func TestConnect(t *testing.T) {
	a := assert.New(t)
	r := require.New(t)

	db := New(true)
	db.Connect("http://localhost:8000", "dbName", "foo", "bar")
	result, err := db.RunAQL("")
	r.Error(err)
	a.Nil(result)

	db = New(false)
	db.Connect("", "", "", "")
	result, err = db.RunAQL("FOR c IN customer RETURN c")
	r.Error(err)
	a.Nil(result)
}
