package arangolite

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Query represents an AQL query.
type Query struct {
	aql       string
	cache     bool
	batchSize int
}

// NewQuery returns a new Query object.
func NewQuery(aql string, params ...interface{}) *Query {
	aql = fmt.Sprintf(aql, params...)
	aql = processAQLQuery(aql)

	return &Query{aql: aql, batchSize: 1000}
}

// Cache enables/disables the caching of the query.
// Unavailable prior to ArangoDB 2.7
func (q *Query) Cache(enable bool) *Query {
	q.cache = enable
	return q
}

func (q *Query) BatchSize(size int) *Query {
	q.batchSize = size
	return q
}

func (q *Query) Run(db *DB) ([]byte, error) {
	async, err := q.RunAsync(db)

	if err != nil {
		return nil, err
	}

	allElem := []interface{}{}

	for async.HasNext() {
		r := async.Next()

		batchElem := []interface{}{}
		_ = json.Unmarshal(r, &batchElem)

		allElem = append(allElem, batchElem...)
	}

	return json.Marshal(allElem)
}

func (q *Query) RunAsync(db *DB) (*AsyncResult, error) {
	if db == nil {
		return nil, errors.New("nil database")
	}

	if len(q.aql) == 0 {
		return &AsyncResult{hasNext: false}, nil
	}

	// db.logBegin("QUERY", "/_api/cursor", jsonQuery)

	// start := time.Now()
	c, err := db.runQuery("/_api/cursor", q)
	// end := time.Now()

	if err != nil {
		return nil, err
	}

	// db.logResult(result.Content, result.Cached, end.Sub(start))

	return &AsyncResult{c: c, hasNext: true}, nil
}

func (q *Query) generate() []byte {
	type QueryFmt struct {
		Query     string `json:"query"`
		Cache     bool   `json:"cache"`
		Count     bool   `json:"count"`
		BatchSize int    `json:"batchSize,omitempty"`
	}

	jsonQuery, _ := json.Marshal(&QueryFmt{Query: q.aql, Cache: q.cache, Count: true, BatchSize: q.batchSize})

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
