package arangolite

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

type DB struct {
	url, database, user, password string
	logger                        *log.Logger
}

func New(logEnabled bool) *DB {
	var out *os.File

	if logEnabled {
		out = os.Stdout
	}

	return &DB{logger: log.New(out, fmt.Sprintf("\n[Arangolite] "), 0)}
}

func (db *DB) Connect(url, database, user, password string) {
	db.url = url
	db.database = database
	db.user = user
	db.password = password
}

func (db *DB) RunAQL(query string, params ...interface{}) ([]byte, error) {
	return db.RunFilteredAQL(nil, query, params...)
}

func (db *DB) RunFilteredAQL(filter *Filter, query string, params ...interface{}) ([]byte, error) {
	if len(query) == 0 {
		return nil, errors.New("the query cannot be empty")
	}

	query, err := buildAQLQuery(filter, query, params...)
	if err != nil {
		return nil, err
	}

	if filter != nil {
		if checkWriteOperation(query) {
			return nil, errors.New("cannot filter on a writting operation")
		}
	}

	query = `{"query": "` + query + `"}`

	db.logger.Printf("%s QUERY %s\n    %s", blue, reset, indentJSON(query))

	// start timer
	start := time.Now()

	r, err := http.Post(db.url+"/_db/"+db.database+"/_api/cursor", "application/json", bytes.NewBufferString(query))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	// stop timer
	end := time.Now()
	latency := end.Sub(start)

	result := &QueryResult{}

	if err := json.NewDecoder(r.Body).Decode(result); err != nil {
		return nil, err
	}

	if result.Error {
		db.logger.Printf("%s RESULT %s | %v\n    ERROR: %s", blue, reset, latency, result.ErrorMessage)
		return nil, errors.New(result.ErrorMessage)
	}

	db.logger.Printf("%s RESULT %s | %v\n    %s", blue, reset, latency, indentJSON(string(result.Content)))

	return result.Content, nil
}

var (
	green   = string([]byte{27, 91, 57, 55, 59, 52, 50, 109})
	white   = string([]byte{27, 91, 57, 48, 59, 52, 55, 109})
	yellow  = string([]byte{27, 91, 57, 55, 59, 52, 51, 109})
	red     = string([]byte{27, 91, 57, 55, 59, 52, 49, 109})
	blue    = string([]byte{27, 91, 57, 55, 59, 52, 52, 109})
	magenta = string([]byte{27, 91, 57, 55, 59, 52, 53, 109})
	cyan    = string([]byte{27, 91, 57, 55, 59, 52, 54, 109})
	reset   = string([]byte{27, 91, 48, 109})
)

func indentJSON(in string) string {
	b := &bytes.Buffer{}
	_ = json.Indent(b, []byte(in), "    ", "  ")

	return b.String()
}

func checkWriteOperation(str string) bool {
	aqlOperators := []string{"REMOVE", "UPDATE", "REPLACE", "INSERT", "UPSERT"}

	regex := ""
	for _, op := range aqlOperators {
		regex = fmt.Sprintf("%s([^\\w]|\\A)(?i)%s([^\\w]|\\z)|", regex, op)
	}

	regex = fmt.Sprintf("(%s)", regex[:len(regex)-1])
	cRegex, _ := regexp.Compile(regex)

	matched := cRegex.FindStringIndex(str)

	if matched != nil {
		return true
	}

	return false
}

func buildAQLQuery(filter *Filter, query string, params ...interface{}) (string, error) {
	query = fmt.Sprintf(query, params...)
	query = processAQLQuery(query)

	if filter == nil {
		return query, nil
	}

	aqlFilter, err := GetAQLFilter(filter)
	if err != nil {
		return "", err
	}

	regex, _ := regexp.Compile(`\s(?i)LET\s`)
	matches := regex.FindStringSubmatchIndex(query)

	if matches == nil {
		query = fmt.Sprintf("LET result = (%s) %s", query, aqlFilter)
	} else {
		lastIndex := matches[len(matches)-1]

		counter := 0
		searching := false
		for i, r := range query[lastIndex:] {
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

		query = fmt.Sprintf("%sLET result = (%s) %s", query[:lastIndex-1], query[lastIndex:], aqlFilter)
	}

	return query, nil
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
