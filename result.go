package arangolite

import "encoding/json"

type errorResult struct {
	Error        bool   `json:"error"`
	ErrorMessage string `json:"errorMessage"`
}

// QueryResult represents the ArangoDB results returned by the REST API when an
// AQL query is executed.
type QueryResult struct {
	errorResult
	Content json.RawMessage `json:"result"`
}

// TransactionResult represents the ArangoDB results returned by the REST API when an
// ArangoDB transaction is executed.
type TransactionResult struct {
	errorResult
	Content struct {
		TransactionContent json.RawMessage `json:"_documents"`
	} `json:"result"`
}
