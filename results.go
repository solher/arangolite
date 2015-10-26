package arangolite

import (
	"bytes"
	"encoding/json"
)

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
	c      chan interface{}
	buffer *bytes.Buffer
}

// NewResult returns a new Result object.
func NewResult(c chan interface{}) *Result {
	return &Result{c: c, buffer: bytes.NewBuffer(nil)}
}

// HasMore indicates if another batch is available to get.
// If true, the content can be read from the buffer.
func (r *Result) HasMore() bool {
	if r.c == nil {
		return false
	}

	r.buffer.Reset()

	obj := <-r.c
	switch msg := obj.(type) {
	case []byte:
		r.buffer.Write(msg)
		return true
	case json.RawMessage:
		r.buffer.Write(msg)
		return true
	}

	return false
}

// Buffer returns the Result buffer.
func (r *Result) Buffer() *bytes.Buffer {
	return r.buffer
}
