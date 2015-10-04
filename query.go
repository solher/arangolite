package arangolite

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type Query struct {
	aql                   string
	filter                *processedFilter
	processTime, execTime time.Duration
}

func NewQuery(aql string, params ...interface{}) *Query {
	start := time.Now()

	aql = fmt.Sprintf(aql, params...)
	aql = processAQLQuery(aql)

	end := time.Now()

	return &Query{aql: aql, processTime: end.Sub(start)}
}

func (q *Query) Filter(filter *Filter) error {
	if filter == nil || len(q.aql) == 0 {
		return nil
	}

	start := time.Now()

	if checkWriteOperation(q.aql) {
		return errors.New("cannot filter on a writting operation")
	}

	aqlFilter, err := GetAQLFilter(filter)
	if err != nil {
		return err
	}

	regex, _ := regexp.Compile(`\s(?i)LET\s`)
	matches := regex.FindStringSubmatchIndex(q.aql)

	if matches == nil {
		q.aql = fmt.Sprintf("LET result = (%s) %s", q.aql, aqlFilter)
	} else {
		lastIndex := matches[len(matches)-1]

		counter := 0
		searching := false
		for i, r := range q.aql[lastIndex:] {
			switch r {
			case '(':
				counter = counter + 1
				searching = true
			case ')':
				counter = counter - 1
			}

			if searching && counter == 0 {
				lastIndex = lastIndex + i + 2
				break
			}
		}

		q.aql = fmt.Sprintf("%sLET result = (%s) %s", q.aql[:lastIndex-1], q.aql[lastIndex:], aqlFilter)
	}

	end := time.Now()
	q.processTime = q.processTime + end.Sub(start)

	return nil
}

func (q *Query) Run(db *DB) ([]byte, error) {
	q.aql = `{"query": "` + q.aql + `"}`
	db.logger.Printf("%s QUERY %s\n    %s", blue, reset, indentJSON(q.aql))

	start := time.Now()
	r, err := http.Post(db.url+"/_db/"+db.database+"/_api/cursor", "application/json", bytes.NewBufferString(q.aql))
	end := time.Now()
	q.execTime = end.Sub(start)

	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	result := &QueryResult{}

	if err := json.NewDecoder(r.Body).Decode(result); err != nil {
		return nil, err
	}

	resultLog := fmt.Sprintf("%s RESULT %s | Processing: %v | Execution: %v | Total: %v\n    ",
		blue, reset, q.processTime, q.execTime, q.processTime+q.execTime)

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

func checkWriteOperation(str string) bool {
	// start := time.Now()
	//
	// aqlOperators := []string{"REMOVE", "UPDATE", "REPLACE", "INSERT", "UPSERT"}
	//
	// regex := ""
	// for _, op := range aqlOperators {
	// 	regex = fmt.Sprintf("%s([^\\w]|\\A)(?i)%s([^\\w]|\\z)|", regex, op)
	// }
	//
	// regex = fmt.Sprintf("(%s)", regex[:len(regex)-1])
	// cRegex, _ := regexp.Compile(regex)
	//
	// matched := cRegex.FindStringIndex(str)
	//
	// if matched != nil {
	// 	return true
	// }
	//
	// end := time.Now()
	// utils.Dump(end.Sub(start))

	return false
}
