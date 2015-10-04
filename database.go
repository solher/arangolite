package arangolite

import (
	"fmt"
	"log"
	"os"
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

func (db *DB) RunAQL(aql string, params ...interface{}) ([]byte, error) {
	q := NewQuery(aql, params...)
	return q.Run(db)
}
