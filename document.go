package arangolite

// Document represents a basic ArangoDB document
// Fields are pointers to allow null values in ArangoDB
type Document struct {
	ID  *string `json:"_id,omitempty"`
	Rev *string `json:"_rev,omitempty"`
	Key *string `json:"_key,omitempty"`
}
