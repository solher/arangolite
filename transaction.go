package arangolite

type Transaction struct {
	Queries []string
}

func NewTransaction() *Transaction {
	return &Transaction{}
}
