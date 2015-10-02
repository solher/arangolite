package arangolite

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildAQLQuery runs tests on the arangolite BuildAQLQuery method.
func TestBuildAQLQuery(t *testing.T) {
	a := assert.New(t)
	r := require.New(t)

	query := `
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

	aqlQuery, err := buildAQLQuery(nil, query)
	r.NoError(err)
	a.Contains(aqlQuery, "FOR d")
	a.Contains(aqlQuery, "FILTER m.documentId == d._id")
	a.Contains(aqlQuery, "}")

	aqlQuery, err = buildAQLQuery(&Filter{}, query)
	r.NoError(err)
	a.Contains(aqlQuery, "LET result = (FOR d")
	a.Contains(aqlQuery, "RETURN m")
	a.Contains(aqlQuery, "}")
	a.Contains(aqlQuery, "FOR var IN result RETURN var")

	query = `
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

	aqlQuery, err = buildAQLQuery(&Filter{}, query)
	r.NoError(err)
	a.Contains(aqlQuery, "LET result = (FOR d")
	a.Contains(aqlQuery, "RETURN m")
	a.Contains(aqlQuery, "}")
	a.Contains(aqlQuery, "FOR var IN result RETURN var")

	aqlQuery, err = buildAQLQuery(&Filter{Where: map[string]interface{}{"and": []interface{}{"foo", "bar"}}}, query)
	r.Error(err)
	a.EqualValues(0, len(aqlQuery))
}
