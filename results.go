package arangolite

import "encoding/json"

type Result struct {
	Error        bool            `json:"error"`
	ErrorMessage string          `json:"errorMessage"`
	Content      json.RawMessage `json:"result"`
	Cached       bool            `json:"cached"`
	Count        int             `json:"count"`
	HasMore      bool            `json:"hasMore"`
	ID           string          `json:"id"`
}

// QueryResult represents the ArangoDB results returned by the REST API when an
// AQL query is executed.
type QueryResult struct {
	Result
	Content json.RawMessage `json:"result"`
	Cached  bool            `json:"cached"`
}

// TransactionResult represents the ArangoDB results returned by the REST API when an
// ArangoDB transaction is executed.
type TransactionResult struct {
	Result
	Content struct {
		TransactionContent json.RawMessage `json:"_documents"`
	} `json:"result"`
}

type AsyncResult struct {
	c       chan interface{}
	hasNext bool
}

func NewAsyncResult(c chan interface{}) *AsyncResult {
	return &AsyncResult{c: c, hasNext: true}
}

func (r *AsyncResult) HasNext() bool {
	return r.hasNext
}

func (ar *AsyncResult) Next() []byte {
	switch r := <-ar.c; r.(type) {
	case json.RawMessage:
		return r.(json.RawMessage)

	case error:
		ar.hasNext = false
		return nil

	default:
		if r == nil {
			ar.hasNext = false
			return nil
		}
	}

	return nil
}
