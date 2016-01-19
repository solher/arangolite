package arangolite

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// Transaction represents an ArangoDB transaction.
type Transaction struct {
	readCol, writeCol []string
	resultVars        []string
	queries           []Query
	returnVar         string
	bindVars          map[string]interface{}
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
	return t
}

// Bind sets the name and value of a bind parameter
// Binding parameters prevents AQL injection
// Example:
// transaction := arangolite.NewTransaction([]string{}, []string{}).
// 		AddQuery("var1", "FOR d IN nodes FILTER d._key == @key RETURN d._id").
// 		AddQuery("var2", "FOR n IN nodes FILTER n._id == {{.var1}}[0] RETURN n._key").Return("var2")
// transaction.Bind("key", 123)
//
func (t *Transaction) Bind(name string, value interface{}) *Transaction {
	if t.bindVars == nil {
		t.bindVars = make(map[string]interface{})
	}
	t.bindVars[name] = value
	return t
}

// Return sets the final "resultVar" that is returned at the end of the transaction.
func (t *Transaction) Return(resultVar string) *Transaction {
	t.returnVar = resultVar
	return t
}

func (t *Transaction) description() string {
	return "TRANSACTION"
}

func (t *Transaction) path() string {
	return "/_api/transaction"
}

func (t *Transaction) method() string {
	return "POST"
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

	jsFunc := bytes.NewBufferString("function () { var db = require(`internal`).db; ")

	for name, value := range t.bindVars {
		jsFunc.WriteString("var ")
		jsFunc.WriteString(name)
		jsFunc.WriteString(" = '")
		jsFunc.WriteString(fmt.Sprint(value))
		jsFunc.WriteString("'; ")
	}

	hasParams := len(t.bindVars) > 0

	if hasParams {
		jsFunc.WriteString("var params = {")
		for name := range t.bindVars {
			jsFunc.WriteString(name)
			jsFunc.WriteString(": ")
			jsFunc.WriteString(name)
			jsFunc.WriteString(", ")
		}
		jsFunc.Truncate(jsFunc.Len() - 2)
		jsFunc.WriteString("}; ")
	}

	for i, query := range t.queries {
		writeQuery(jsFunc, query.aql, hasParams, t.resultVars[i])
	}

	if len(t.returnVar) > 0 {
		jsFunc.WriteString("return ")
		jsFunc.WriteString(t.returnVar)
		jsFunc.WriteString(";")
	}

	jsFunc.WriteRune('}')

	transactionFmt.Action = jsFunc.String()
	jsonTransaction, _ := json.Marshal(transactionFmt)

	return jsonTransaction
}

// writeQuery translate a given aql query to bytes
// buff the buffer containing the resulting bytes
// aql the AQL query
// resultVarName the name of the variable that will accept the query result, if any - may be empty
func writeQuery(buff *bytes.Buffer, aql string, hasParams bool, resultVarName string) {
	if len(resultVarName) > 0 {
		buff.WriteString("var ")
		buff.WriteString(resultVarName)
		buff.WriteString(" = ")
	}

	buff.WriteString("db._query(aqlQuery`")
	buff.WriteString(aql)
	if hasParams {
		buff.WriteString("`, params).toArray(); ")
	} else {
		buff.WriteString("`).toArray(); ")
	}
}

func toES6Template(query string) string {
	query = strings.Replace(query, "{{.", "${", -1)
	return strings.Replace(query, "}}", "}", -1)
}
