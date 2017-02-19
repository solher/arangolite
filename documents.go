package arangolite

// Document represents a basic ArangoDB document
// Fields are pointers to allow null values in ArangoDB
type Document struct {
	// The document handle. Format: ':collection/:key'
	ID string `json:"_id,omitempty"`
	// The document's revision token. Changes at each update.
	Rev string `json:"_rev,omitempty"`
	// The document's unique key.
	Key string `json:"_key,omitempty"`
}

// Edge represents a basic ArangoDB edge
// Fields are pointers to allow null values in ArangoDB
type Edge struct {
	Document
	// Reference to another document. Format: ':collection/:key'
	From string `json:"_from,omitempty"`
	// Reference to another document. Format: ':collection/:key'
	To string `json:"_to,omitempty"`
}
