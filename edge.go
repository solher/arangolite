package arangolite

// Edge represents a basic ArangoDB edge
// Fields are pointers to allow null values in ArangoDB
type Edge struct {
	Document
	From *string `json:"_from,omitempty"`
	To   *string `json:"_to,omitempty"`
}
