package arangolite

type Edge struct {
	Document
	From string `json:"_from,omitempty"`
	To   string `json:"_to,omitempty"`
}
