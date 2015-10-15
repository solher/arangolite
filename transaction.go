package arangolite

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Transaction represents an ArangoDB transaction.
type Transaction struct {
	readCol, writeCol []string
	resultVars        []string
	queries           []Query
	returnVar         string
}

// NewTransaction returns a new Transaction object.
func NewTransaction(readCol, writeCol []string) *Transaction {
	if readCol == nil {
		readCol = []string{}
	}

	if writeCol == nil {
		writeCol = []string{}
	}

	return &Transaction{readCol: readCol, writeCol: writeCol}
}

// AddQuery adds a new AQL query to the transaction. The result will be set in
// a temp variable named after the value of "resultVar".
// To use it from elsewhere in the transaction, use the Go templating convention.
//
// e.g. NewTransaction([]string{}, []string{}).
//      AddQuery("var1", "FOR d IN documents RETURN d").
//      AddQuery("var2", "FOR d IN {{.var1}} RETURN d._id").Run(db)
//
func (t *Transaction) AddQuery(resultVar, aql string, params ...interface{}) *Transaction {
	t.resultVars = append(t.resultVars, resultVar)
	t.queries = append(t.queries, *NewQuery(aql, params...))

	if len(resultVar) != 0 {
		t.returnVar = resultVar
	}

	return t
}

// Return sets the final "resultVar" that is returned at the end of the transaction.
func (t *Transaction) Return(resultVar string) *Transaction {
	t.returnVar = resultVar
	return t
}

// Run executes the Transaction into the database passed as argument.
func (t *Transaction) Run(db *DB) ([]byte, error) {
	if db == nil {
		return nil, errors.New("nil database")
	}

	// db.logBegin("TRANSACTION", "/_api/transaction", jsonTransaction)

	start := time.Now()
	_, err := db.runQuery("/_api/transaction", t)
	end := time.Now()

	if err != nil {
		return nil, err
	}

	r := []byte{}

	result := &TransactionResult{}
	_ = json.Unmarshal(r, result)

	if result.Error {
		db.logError(result.ErrorMessage, end.Sub(start))
		return nil, errors.New(result.ErrorMessage)
	}

	db.logResult(result.Content.TransactionContent, false, end.Sub(start))

	return result.Content.TransactionContent, nil
}

func (t *Transaction) generate() []byte {
	type TransactionFmt struct {
		Collections struct {
			Read  []string `json:"read"`
			Write []string `json:"write"`
		} `json:"collections"`
		Action string `json:"action"`
	}

	transactionFmt := &TransactionFmt{}
	transactionFmt.Collections.Read = t.readCol
	transactionFmt.Collections.Write = t.writeCol

	jsFunc := "function () {var db = require(`internal`).db; "

	for i, query := range t.queries {
		query.aql = replaceTemplate(query.aql)
		varName := t.resultVars[i]

		if len(varName) == 0 {
			jsFunc = fmt.Sprintf("%sdb._query(`%s`); ", jsFunc, query.aql)
		} else {
			jsFunc = fmt.Sprintf("%svar %s = db._query(`%s`); ", jsFunc, varName, query.aql)
		}
	}

	transactionFmt.Action = jsFunc + "return " + t.returnVar + ";}"

	jsonTransaction, _ := json.Marshal(transactionFmt)

	return jsonTransaction
}

func (t *Transaction) getBatchSize() int {
	return 1000
}

func replaceTemplate(query string) string {
	var jsResult string

	reg, _ := regexp.Compile("{{\\.(.*)}}")
	templates := reg.FindAllString(query, -1)

	if templates == nil {
		return query
	}

	for _, t := range templates {
		jsResult = "` + JSON.stringify(" + t[3:len(t)-2] + "._documents) + `"
		query = strings.Replace(query, t, jsResult, -1)
	}

	return query
}
