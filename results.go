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

func NewResult(c chan interface{}) *Result {
	return &Result{c: c, buffer: bytes.NewBuffer([]byte{'['})}
}

// HasMore indicates if another batch is available to get.
func (r *Result) HasMore() bool {
	if r.c == nil {
		return false
	}

	obj := <-r.c
	switch msg := obj.(type) {
	case json.RawMessage:
		r.buffer.Write(msg[1 : len(msg)-1])
		r.buffer.WriteRune(',')
		return true
	}

	if r.buffer.Len() > 1 {
		r.buffer.Truncate(r.buffer.Len() - 1)
	}

	r.buffer.WriteRune(']')

	return false
}

// Next returns the JSON formatted next batch.
// func (ar *Result) Next() []byte {
// 	switch r := <-ar.c; r.(type) {
// 	case json.RawMessage:
// 		return r.(json.RawMessage)
// 	}
//
// 	ar.hasNext = false
// 	return nil
// }

func (ar *Result) Buffer() *bytes.Buffer {
	return ar.buffer
}
