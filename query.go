package arangolite

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
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
	q.aql = `{"query": "` + q.aql + `"}`
	db.logger.Printf("%s QUERY %s\n    %s", blue, reset, indentJSON(q.aql))

	start := time.Now()
	r, err := db.conn.Post(db.url+"/_db/"+db.database+"/_api/cursor", "application/json", bytes.NewBufferString(q.aql))
	end := time.Now()

	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	result := &QueryResult{}

	if err := json.NewDecoder(r.Body).Decode(result); err != nil {
		return nil, err
	}

	resultLog := fmt.Sprintf("%s RESULT %s | Execution: %v\n    ",
		blue, reset, end.Sub(start))

	if result.Error {
		db.logger.Printf("%sERROR: %s", resultLog, result.ErrorMessage)
		return nil, errors.New(result.ErrorMessage)
	}

	db.logger.Printf(resultLog + indentJSON(string(result.Content)))

	return result.Content, nil
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
