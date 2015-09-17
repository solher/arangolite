package arangolite

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type DB struct {
	url, database, user, password string
}

func New() *DB {
	return &DB{}
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

	r, err := http.Post(db.url+"/_db/"+db.database+"/_api/cursor", "application/json", bytes.NewBufferString(query))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	result := &QueryResult{}

	if err := json.NewDecoder(r.Body).Decode(result); err != nil {
		return nil, err
	}

	if result.Error {
		return nil, errors.New(result.ErrorMessage)
	}

	return result.Content, nil
}

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

func processQuery(query string, params ...interface{}) string {
	query = strings.Replace(query, `"`, "'", -1)
	query = strings.Replace(query, "\n", " ", -1)
	query = strings.Replace(query, "\t", "", -1)
	query = strings.TrimSpace(query)
	query = fmt.Sprintf(query, params...)

	return query
}
