package arangolite

import "encoding/json"

// QueryResult represents the ArangoDB results returned by the REST API when an
// AQL query is executed.
type QueryResult struct {
	HasMore      bool            `json:"hasMore"`
	Error        bool            `json:"error"`
	ErrorNum     int             `json:"errorNum"`
	ErrorMessage string          `json:"errorMessage"`
	Content      json.RawMessage `json:"result"`
	Code         int             `json:"code"`
	Count        int             `json:"count"`
}
