package requests

import (
	"encoding/json"
	"fmt"
	"strings"
)

// AQL represents an AQL query.
type AQL struct {
	query     string
	bindVars  map[string]interface{}
	cache     *bool
	batchSize int
}

// NewAQL returns a new AQL object.
func NewAQL(query string, params ...interface{}) *AQL {
	query = processAQL(query) // Process to remove eventual tabs/spaces used when indenting the query
	query = fmt.Sprintf(query, params...)
	query = strings.Replace(query, `"`, "'", -1) // Replace by single quotes so there is no conflict when serialised in JSON

	return &AQL{query: query}
}

// Cache enables/disables the caching of the query.
// Unavailable prior to ArangoDB 2.7
func (a *AQL) Cache(enable bool) *AQL {
	a.cache = &enable
	return a
}

// BatchSize sets the batch size of the query
func (a *AQL) BatchSize(size int) *AQL {
	a.batchSize = size
	return a
}

// Bind sets the name and value of a bind parameter
// Binding parameters prevents AQL injection
func (a *AQL) Bind(name string, value interface{}) *AQL {
	if a.bindVars == nil {
		a.bindVars = make(map[string]interface{})
	}
	a.bindVars[name] = value
	return a
}

func (a *AQL) Path() string {
	return "/_api/cursor"
}

func (a *AQL) Method() string {
	return "POST"
}

func (a *AQL) Generate() []byte {
	type AQLFmt struct {
		Query     string                 `json:"query"`
		BindVars  map[string]interface{} `json:"bindVars,omitempty"`
		Cache     *bool                  `json:"cache,omitempty"`
		BatchSize int                    `json:"batchSize,omitempty"`
	}

	jsonAQL, _ := json.Marshal(&AQLFmt{Query: a.query, BindVars: a.bindVars, Cache: a.cache, BatchSize: a.batchSize})

	return jsonAQL
}

func processAQL(query string) string {
	query = strings.Replace(query, "\n", " ", -1)
	query = strings.Replace(query, "\t", "", -1)

	split := strings.Split(query, " ")
	split2 := []string{}

	for _, s := range split {
		if len(s) == 0 {
			continue
		}
		split2 = append(split2, s)
	}

	query = strings.Join(split2, " ")

	return query
}
