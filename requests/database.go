package requests

import (
	"encoding/json"
	"fmt"
)

// CurrentDatabase retrieves information on the current database.
type CurrentDatabase struct{}

func (r *CurrentDatabase) Path() string {
	return "/_api/database/current"
}

func (r *CurrentDatabase) Method() string {
	return "GET"
}

func (r *CurrentDatabase) Generate() []byte {
	return nil
}

// CreateDatabase creates a new database.
type CreateDatabase struct {
	Username string                   `json:"username,omitempty"`
	Name     string                   `json:"name"`
	Extra    json.RawMessage          `json:"extra,omitempty"`
	Passwd   string                   `json:"passwd,omitempty"`
	Active   bool                     `json:"active,omitempty"`
	Users    []map[string]interface{} `json:"users,omitempty"`
}

func (r *CreateDatabase) Path() string {
	return "/_api/database"
}

func (r *CreateDatabase) Method() string {
	return "POST"
}

func (r *CreateDatabase) Generate() []byte {
	m, _ := json.Marshal(r)
	return m
}

// DropDatabase deletes a database.
type DropDatabase struct {
	Name string
}

func (r *DropDatabase) Path() string {
	return fmt.Sprintf("/_api/database/%s", r.Name)
}

func (r *DropDatabase) Method() string {
	return "DELETE"
}

func (r *DropDatabase) Generate() []byte {
	return nil
}
