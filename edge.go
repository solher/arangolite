package arangolite

// Edge represents a basic ArangoDB edge
// Fields are pointers to allow null values in ArangoDB
type Edge struct {
	Document
	// Reference to another document. Format: ':collection/:key'
	// Required: true
	From *string `json:"_from,omitempty"`
	// Reference to another document. Format: ':collection/:key'
	// Required: true
	To *string `json:"_to,omitempty"`
}
