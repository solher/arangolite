package requests

import "encoding/json"

// SetCacheProperties sets the query cache properties.
type SetCacheProperties struct {
	Mode       string `json:"mode,omitempty"`
	MaxResults int    `json:"maxResults,omitempty"`
}

func (r *SetCacheProperties) Path() string {
	return "/_api/query-cache/properties"
}

func (r *SetCacheProperties) Method() string {
	return "PUT"
}

func (r *SetCacheProperties) Generate() []byte {
	m, _ := json.Marshal(r)
	return m
}

// GetCacheProperties retrieves the current query cache properties.
type GetCacheProperties struct{}

func (r *GetCacheProperties) Path() string {
	return "/_api/query-cache/properties"
}

func (r *GetCacheProperties) Method() string {
	return "GET"
}

func (r *GetCacheProperties) Generate() []byte {
	return nil
}
