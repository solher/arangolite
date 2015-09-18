package arangolite

type Document struct {
	ID  string `json:"_id,omitempty"`
	Rev string `json:"_rev,omitempty"`
	Key string `json:"_key,omitempty"`
}
