package arangolite

import (
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	shortQuery = `
    FOR d
    IN documents
    RETURN {
        document: d,
        metaData: (
            FOR m
            IN metaData
            FILTER m.documentId == d._id
            RETURN m
        )
    }
  `

	longQuery = `
      LET foo = (
        FOR f IN foo RETURN f._id
      )

      LET bar = (
        FOR b IN foo RETURN b._id
      )

      FOR d
      IN documents
      RETURN {
          document: d,
          metaData: (
              FOR m
              IN metaData
              FILTER m.documentId == d._id
              RETURN m
          ),
          foo: foo,
          bar: bar
      }
    `

	writeQuery = `
    FOR d
    IN documents
    INSERT d
    `
)

// TestFilter runs tests on the Query Filter method.
func TestFilter(t *testing.T) {
	a := assert.New(t)
	r := require.New(t)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "http://arangodb:8000/_db/dbName/_api/cursor",
		httpmock.NewStringResponder(200, `{}`))

	db := New(false)
	db.Connect("http://arangodb:8000", "dbName", "foo", "bar")

	result, err := NewQuery(shortQuery).Run(nil)
	r.Error(err)
	a.Nil(result)

	result, err = NewQuery("").Run(db)
	r.NoError(err)
	a.Nil(result)

	result, err = NewQuery(shortQuery).Run(db)
	r.NoError(err)
	a.Nil(result)
}
