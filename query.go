package arangolite

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Query represents an AQL query.
type Query struct {
	aql string
}

// NewQuery returns a new Query object.
func NewQuery(aql string, params ...interface{}) *Query {
	aql = fmt.Sprintf(aql, params...)
	aql = processAQLQuery(aql)

	return &Query{aql: aql}
}

// Run executes the Query into the database passed as argument.
func (q *Query) Run(db *DB) ([]byte, error) {
	if db == nil {
		return nil, errors.New("nil database")
	}

	if len(q.aql) == 0 {
		return nil, nil
	}

	type QueryFmt struct {
		Query string `json:"query"`
	}

	jsonQuery, _ := json.Marshal(&QueryFmt{Query: q.aql})

	return db.runQuery("/_api/cursor", jsonQuery)
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
