package requests

import "encoding/json"
import "fmt"

// CreateHashIndex creates a hash index in database.
type CreateHashIndex struct {
	CollectionName string   `json:"-"`
	Fields         []string `json:"fields,omitempty"`
	Unique         bool     `json:"unique,omitempty"`
	Type           string   `json:"type,omitempty"`
	Sparse         bool     `json:"sparse,omitempty"`
}

func (r *CreateHashIndex) Path() string {
	return fmt.Sprintf("/_api/index?collection=%s", r.CollectionName)
}

func (r *CreateHashIndex) Method() string {
	return "POST"
}

func (r *CreateHashIndex) Generate() []byte {
	r.Type = "hash"
	m, _ := json.Marshal(r)
	return m
}
