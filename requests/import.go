package requests

import "fmt"

// ImportCollection imports data to a collection
type ImportCollection struct {
	CollectionName string
	Data           []byte
}

func (c *ImportCollection) Path() string {
	return fmt.Sprintf("/_api/import/?type=auto&collection=%s", c.CollectionName)
}

func (c *ImportCollection) Method() string {
	return "POST"
}

func (c *ImportCollection) Generate() []byte {
	return c.Data
}
