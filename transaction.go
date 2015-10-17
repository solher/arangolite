package arangolite

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
)

// Transaction represents an ArangoDB transaction.
type Transaction struct {
	readCol, writeCol []string
	resultVars        []string
	queries           []Query
	returnVar         string
	batchSize         int
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
	t.queries = append(t.queries, *NewQuery(toES6Template(aql), params...))

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

// Run runs the transaction synchronously and returns the JSON array of all elements
// of every batch returned by the database.
func (t *Transaction) Run(db *DB) ([]byte, error) {
	if db == nil {
		return nil, errors.New("nil database")
	}

	c, err := db.runQuery("/_api/transaction", t)

	if err != nil {
		return nil, err
	}

	return db.syncResult(&Result{c: c, hasNext: true})
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

	jsFunc := bytes.NewBufferString("function () {var db = require(`internal`).db; ")

	for i, query := range t.queries {
		varName := t.resultVars[i]

		if len(varName) > 0 {
			jsFunc.WriteString("var ")
			jsFunc.WriteString(varName)
			jsFunc.WriteString(" = ")
		}

		jsFunc.WriteString("db._query(aqlQuery`")
		jsFunc.WriteString(query.aql)
		jsFunc.WriteString("`).toArray();")
	}

	jsFunc.WriteString(" return ")
	jsFunc.WriteString(t.returnVar)
	jsFunc.WriteString(";}")

	transactionFmt.Action = jsFunc.String()
	jsonTransaction, _ := json.Marshal(transactionFmt)

	return jsonTransaction
}

func (t *Transaction) description() string {
	return "TRANSACTION"
}

func toES6Template(query string) string {
	query = strings.Replace(query, "{{.", "${", -1)
	return strings.Replace(query, "}}", "}", -1)
}
