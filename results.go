package arangolite

import "encoding/json"

type result struct {
	Error        bool            `json:"error"`
	ErrorMessage string          `json:"errorMessage"`
	Content      json.RawMessage `json:"result"`
	Cached       bool            `json:"cached"`
	HasMore      bool            `json:"hasMore"`
	ID           string          `json:"id"`
}

// TransactionResult represents the ArangoDB results returned by the REST API when an
// ArangoDB transaction is executed.
type TransactionResult struct {
	result
	Content struct {
		TransactionContent json.RawMessage `json:"_documents"`
	} `json:"result"`
}

type Result struct {
	c       chan interface{}
	hasNext bool
}

func (r *Result) HasNext() bool {
	return r.hasNext
}

func (ar *Result) Next() []byte {
	switch r := <-ar.c; r.(type) {
	case json.RawMessage:
		return r.(json.RawMessage)
	}

	ar.hasNext = false
	return nil
}
