package arangolite

import (
	"testing"

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
	q := NewQuery(shortQuery)

	err := q.Filter(nil)
	r.NoError(err)
	a.Contains(q.aql, "FOR d")
	a.Contains(q.aql, "FILTER m.documentId == d._id")
	a.Contains(q.aql, "}")

	err = q.Filter(&Filter{})
	r.NoError(err)
	a.Contains(q.aql, "FOR d")
	a.Contains(q.aql, "FILTER m.documentId == d._id")
	a.Contains(q.aql, "}")
	a.Contains(q.aql, "LET result = (FOR d")
	a.Contains(q.aql, "FOR var IN result RETURN var")

	err = q.Filter(&Filter{Limit: 2})
	r.NoError(err)
	a.Contains(q.aql, "FOR d")
	a.Contains(q.aql, "FILTER m.documentId == d._id")
	a.Contains(q.aql, "}")
	a.Contains(q.aql, "LET result = (FOR d")
	a.Contains(q.aql, "FOR var IN result LIMIT 2 RETURN var")

	q = NewQuery(longQuery)

	err = q.Filter(&Filter{})
	r.NoError(err)
	a.Contains(q.aql, "LET result = (FOR d")
	a.Contains(q.aql, "RETURN m")
	a.Contains(q.aql, "}")
	a.Contains(q.aql, "FOR var IN result RETURN var")

	err = q.Filter(&Filter{Where: map[string]interface{}{"and": []interface{}{"foo", "bar"}}})
	r.Error(err)

	q = NewQuery(writeQuery)

	err = q.Filter(&Filter{})
	r.Error(err)
}
