package arangolite

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Query represents an AQL query.
type Query struct {
	aql   string
	cache bool
}

// NewQuery returns a new Query object.
func NewQuery(aql string, params ...interface{}) *Query {
	aql = fmt.Sprintf(aql, params...)
	aql = processAQLQuery(aql)

	return &Query{aql: aql}
}

// Cache enables/disables the caching of the query.
// Unavailable prior to ArangoDB 2.7
func (q *Query) Cache(enable bool) *Query {
	q.cache = enable
	return q
}

// Run executes the Query into the database passed as argument.
func (q *Query) Run(db *DB) ([]byte, error) {
	if db == nil {
		return nil, errors.New("nil database")
	}

	if len(q.aql) == 0 {
		return nil, nil
	}

	jsonQuery := generateQuery(q)

	db.logBegin("QUERY", jsonQuery)

	start := time.Now()
	r, err := db.runQuery("/_api/cursor", jsonQuery)
	end := time.Now()

	if err != nil {
		return nil, err
	}

	result := &QueryResult{}
	_ = json.Unmarshal(r, result)

	if result.Error {
		db.logError(result.ErrorMessage, end.Sub(start))
		return nil, errors.New(result.ErrorMessage)
	}

	db.logResult(result.Content, end.Sub(start))

	return result.Content, nil
}

func generateQuery(q *Query) []byte {
	type QueryFmt struct {
		Query string `json:"query"`
		Cache bool   `json:"cache"`
	}

	jsonQuery, _ := json.Marshal(&QueryFmt{Query: q.aql, Cache: q.cache})

	return jsonQuery
}

func processAQLQuery(query string) string {
	query = strings.Replace(query, `"`, "'", -1)
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
