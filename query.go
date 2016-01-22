package arangolite

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Query represents an AQL query.
type Query struct {
	aql       string
	bindVars  map[string]interface{}
	cache     *bool
	batchSize int
}

// NewQuery returns a new Query object.
func NewQuery(aql string, params ...interface{}) *Query {
	aql = processAQLQuery(aql) // Process to remove eventual tabs/spaces used when indenting the query
	aql = fmt.Sprintf(aql, params...)
	aql = strings.Replace(aql, `"`, "'", -1) // Replace by single quotes so there is no conflict when serialised in JSON

	return &Query{aql: aql}
}

// Cache enables/disables the caching of the query.
// Unavailable prior to ArangoDB 2.7
func (q *Query) Cache(enable bool) *Query {
	q.cache = &enable
	return q
}

// BatchSize sets the batch size of the query
func (q *Query) BatchSize(size int) *Query {
	q.batchSize = size
	return q
}

// Bind sets the name and value of a bind parameter
// Binding parameters prevents AQL injection
func (q *Query) Bind(name string, value interface{}) *Query {
	if q.bindVars == nil {
		q.bindVars = make(map[string]interface{})
	}
	q.bindVars[name] = value
	return q
}

func (q *Query) Description() string {
	return "QUERY"
}

func (q *Query) Path() string {
	return "/_api/cursor"
}

func (q *Query) Method() string {
	return "POST"
}

func (q *Query) Generate() []byte {
	type QueryFmt struct {
		Query     string                 `json:"query"`
		BindVars  map[string]interface{} `json:"bindVars,omitempty"`
		Cache     *bool                  `json:"cache,omitempty"`
		BatchSize int                    `json:"batchSize,omitempty"`
	}

	jsonQuery, _ := json.Marshal(&QueryFmt{Query: q.aql, BindVars: q.bindVars, Cache: q.cache, BatchSize: q.batchSize})

	return jsonQuery
}

func processAQLQuery(query string) string {
	query = strings.Replace(query, "\n", " ", -1)
	query = strings.Replace(query, "\t", "", -1)

	split := strings.Split(query, " ")
	split2 := []string{}

	for _, s := range split {
		if len(s) == 0 {
			continue
		}
		split2 = append(split2, s)
	}

	query = strings.Join(split2, " ")

	return query
}
