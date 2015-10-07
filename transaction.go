package arangolite

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
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
	t.returnVar = resultVar
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

	jsFunc := `function () {var db = require('internal').db; `

	for i, query := range t.queries {
		query.aql = replaceTemplate(query.aql)
		jsFunc = fmt.Sprintf(`%svar %s = db._query('%s'); `, jsFunc, t.resultVars[i], query.aql)
	}

	transactionFmt.Action = jsFunc + `return ` + t.returnVar + `;`

	jsonQuery, _ := json.Marshal(transactionFmt)

	return db.runQuery("/_api/transaction", jsonQuery)
}

func replaceTemplate(query string) string {
	var jsResult string

	reg, _ := regexp.Compile("{{\\.(.*)}}")
	templates := reg.FindAllString(query, -1)

	if templates == nil {
		return query
	}

	for _, t := range templates {
		jsResult = `' + JSON.stringify(` + t[3:len(t)-2] + `._documents) + '`
		query = strings.Replace(query, t, jsResult, -1)
	}

	return query
}
