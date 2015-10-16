package arangolite

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"regexp"
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

// BatchSize sets the batch size of the transaction
func (t *Transaction) BatchSize(size int) *Transaction {
	t.batchSize = size
	return t
}

// Run runs the transaction synchronously and returns the JSON array of all elements
// of every batch returned by the database.
func (t *Transaction) Run(db *DB) ([]byte, error) {
	async, err := t.RunAsync(db)
	if err != nil {
		return nil, err
	}

	return db.syncResult(async)
}

// RunAsync runs the transaction asynchronously and returns an async Result object.
func (t *Transaction) RunAsync(db *DB) (*Result, error) {
	if db == nil {
		return nil, errors.New("nil database")
	}

	c, err := db.runQuery("/_api/transaction", t)

	if err != nil {
		return nil, err
	}

	return &Result{c: c, hasNext: true}, nil
}

func (t *Transaction) generate() []byte {
	type TransactionFmt struct {
		Collections struct {
			Read  []string `json:"read"`
			Write []string `json:"write"`
		} `json:"collections"`
		Action    string `json:"action"`
		BatchSize int    `json:"batchSize,omitempty"`
	}

	transactionFmt := &TransactionFmt{BatchSize: t.batchSize}
	transactionFmt.Collections.Read = t.readCol
	transactionFmt.Collections.Write = t.writeCol

	jsFunc := bytes.NewBufferString("function () {var db = require(`internal`).db; ")

	for i, query := range t.queries {
		query.aql = replaceTemplate(query.aql)
		varName := t.resultVars[i]

		if len(varName) > 0 {
			jsFunc.WriteString("var ")
			jsFunc.WriteString(varName)
			jsFunc.WriteString(" = ")
		}

		jsFunc.WriteString("db._query(`")
		jsFunc.WriteString(query.aql)
		jsFunc.WriteString("`);")
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

func (t *Transaction) decode(body io.ReadCloser, r *result) {
	json.NewDecoder(body).Decode(r)
	body.Close()

	content := &struct {
		Documents json.RawMessage `json:"_documents"`
	}{}

	json.Unmarshal(r.Content, content)

	r.Content = content.Documents
}

func replaceTemplate(query string) string {
	reg, _ := regexp.Compile("{{\\.(.*)}}")
	templates := reg.FindAllString(query, -1)

	if templates == nil {
		return query
	}

	jsResult := bytes.NewBuffer(nil)

	for _, t := range templates {
		jsResult.WriteString("` + JSON.stringify(")
		jsResult.WriteString(t[3 : len(t)-2])
		jsResult.WriteString("._documents) + `")
		query = strings.Replace(query, t, jsResult.String(), -1)
	}

	return query
}
