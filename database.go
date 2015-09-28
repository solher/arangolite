package arangolite

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
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

	return &DB{logger: log.New(out, fmt.Sprintf("[Arangolite] "), 0)}
}

func (db *DB) Connect(url, database, user, password string) {
	db.url = url
	db.database = database
	db.user = user
	db.password = password
}

func (db *DB) RunAQL(query string, params ...interface{}) ([]byte, error) {
	query = processQuery(query, params...)
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

// func (db *DB) RunAQLTransaction(t *Transaction) ([]byte, error) {
// 	readCol, err := json.Marshal(t.readCol)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	writeCol, err := json.Marshal(t.writeCol)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	var jsQueries string
//
// 	for _, query := range t.queries {
// 		jsQueries = fmt.Sprintf(`%sdb._query("%s");`+"\n", jsQueries, query)
// 	}
//
// 	query := fmt.Sprintf(`
// 		{
// 		  collections: {
// 			 	read: %s
// 		    write: %s
// 		  },
// 		  action: function () {
// 		    var db = require("org/arangodb").db;
// 		    %s
// 		  }
// 		}
// 	`, readCol, writeCol, jsQueries)
//
// 	utils.Dump(query)
//
// 	r, err := http.Post(db.url+"/_db/"+db.database+"/_api/transaction", "application/json", bytes.NewBufferString(query))
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer r.Body.Close()
//
// 	result := &QueryResult{}
//
// 	if err := json.NewDecoder(r.Body).Decode(result); err != nil {
// 		return nil, err
// 	}
//
// 	if result.Error {
// 		return nil, errors.New(result.ErrorMessage)
// 	}
//
// 	return result.Content, nil
// }

func indentJSON(in string) string {
	b := &bytes.Buffer{}
	_ = json.Indent(b, []byte(in), "    ", "  ")

	return b.String()
}

func processQuery(query string, params ...interface{}) string {
	query = strings.Replace(query, `"`, "'", -1)
	query = strings.Replace(query, "\n", " ", -1)
	query = strings.Replace(query, "\t", "", -1)
	query = strings.TrimSpace(query)
	query = fmt.Sprintf(query, params...)

	return query
}
