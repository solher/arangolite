package requests_test

import (
	"testing"

	"encoding/json"

	"github.com/solher/arangolite/requests"
)

const indentedQuery = `
    FOR x
    IN documents
    RETURN {
        document: x,
        foo: bar
    }
`

type aql struct {
	Query     string                 `json:"query"`
	BindVars  map[string]interface{} `json:"bindVars,omitempty"`
	Cache     bool                   `json:"cache"`
	BatchSize int                    `json:"batchSize,omitempty"`
}

// TestAQL runs tests on the AQL request.
func TestAQL(t *testing.T) {
	var testCases = []struct {
		// Case description
		description string
		// Arguments
		query     string
		params    []interface{}
		bind      map[string]interface{}
		batchSize int
		cache     bool
		// Expected results
		output aql
	}{
		{
			description: "basic query",
			query:       "FOR x IN documents RETURN x",
			cache:       false,
			output: aql{
				Query: "FOR x IN documents RETURN x",
				Cache: false,
			},
		},
		{
			description: "cache",
			query:       "FOR x IN documents RETURN x",
			cache:       true,
			output: aql{
				Query: "FOR x IN documents RETURN x",
				Cache: true,
			},
		},
		{
			description: "batch size",
			query:       "FOR x IN documents RETURN x",
			cache:       false,
			batchSize:   1000,
			output: aql{
				Query:     "FOR x IN documents RETURN x",
				Cache:     false,
				BatchSize: 1000,
			},
		},
		{
			description: "bind parameters",
			query:       "FOR x IN documents FILTER x.attr1 == @attr1 AND x.attr2 == @attr2 RETURN x",
			cache:       false,
			bind: map[string]interface{}{
				"attr1": "foobar",
				"attr2": 100,
			},
			output: aql{
				Query: "FOR x IN documents FILTER x.attr1 == @attr1 AND x.attr2 == @attr2 RETURN x",
				Cache: false,
				BindVars: map[string]interface{}{
					"attr1": "foobar",
					"attr2": 100,
				},
			},
		},
		{
			description: "indented query",
			query:       indentedQuery,
			cache:       false,
			output: aql{
				Query: "FOR x IN documents RETURN { document: x, foo: bar }",
				Cache: false,
			},
		},
		{
			description: "use of parameters",
			query:       `FOR x IN documents FILTER x.attr1 == "%s" RETURN x`,
			params:      []interface{}{"foobar"},
			cache:       false,
			output: aql{
				Query: `FOR x IN documents FILTER x.attr1 == 'foobar' RETURN x`,
				Cache: false,
			},
		},
		{
			description: "quotes are replaced by single quotes, even from parameters",
			query:       `UPSERT { "id":"foo" } INSERT %s UPDATE {} IN bar`,
			params:      []interface{}{`{ "id":"foo" }`},
			cache:       false,
			output: aql{
				Query: `UPSERT { 'id':'foo' } INSERT { 'id':'foo' } UPDATE {} IN bar`,
				Cache: false,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			aql := requests.NewAQL(tc.query, tc.params...).Cache(tc.cache)
			if tc.batchSize != 0 {
				aql.BatchSize(tc.batchSize)
			}
			for name, value := range tc.bind {
				aql.Bind(name, value)
			}
			output, err := json.Marshal(tc.output)
			if err != nil {
				t.Error(err)
			}
			if string(output) != string(aql.Generate()) {
				t.Errorf("unexpected output. Expected %s, got %s", output, aql.Generate())
			}
		})
	}
}
