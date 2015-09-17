package arangolite

type Transaction struct {
	readCol, writeCol, queries []string
}

func NewTransaction(readCol, writeCol []string) *Transaction {
	return &Transaction{readCol: readCol, writeCol: writeCol}
}

func (t *Transaction) AddAQL(query string, params ...interface{}) {
	t.queries = append(t.queries, processQuery(query, params...))
}
