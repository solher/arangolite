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

// Result defines a query result, allowing the user to retrieve asynchronously
// every batch returned by the database.
type Result struct {
	c       chan interface{}
	hasNext bool
}

// HasNext indicates if another batch is available to get.
func (r *Result) HasNext() bool {
	return r.hasNext
}

// Next returns the JSON formatted next batch.
func (ar *Result) Next() []byte {
	switch r := <-ar.c; r.(type) {
	case json.RawMessage:
		return r.(json.RawMessage)
	}

	ar.hasNext = false
	return nil
}
