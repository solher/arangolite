package arangolite

import "encoding/json"

// CreateCollection creates a collection in database.
type CreateCollection struct {
	JournalSize    int                    `json:"journalSize,omitempty"`
	KeyOptions     map[string]interface{} `json:"keyOptions,omitempty"`
	Name           string                 `json:"name"`
	WaitForSync    *bool                  `json:"waitForSync,omitempty"`
	DoCompact      *bool                  `json:"doCompact,omitempty"`
	IsVolatile     *bool                  `json:"isVolatile,omitempty"`
	ShardKeys      []string               `json:"shardKeys,omitempty"`
	NumberOfShards int                    `json:"numberOfShards,omitempty"`
	IsSystem       *bool                  `json:"isSystem,omitempty"`
	Type           int                    `json:"type,omitempty"`
	IndexBuckets   int                    `json:"indexBuckets,omitempty"`
}

func (q *CreateCollection) description() string {
	return "CREATE COLLECTION"
}

func (q *CreateCollection) path() string {
	return "/_api/collection"
}

func (q *CreateCollection) method() string {
	return "POST"
}

func (q *CreateCollection) generate() []byte {
	m, _ := json.Marshal(q)
	return m
}

// DropCollection deletes a collection in database.
type DropCollection struct {
	Name string `json:"name"`
}

func (q *DropCollection) description() string {
	return "DROP COLLECTION"
}

func (q *DropCollection) path() string {
	return "/_api/collection/" + q.Name
}

func (q *DropCollection) method() string {
	return "DELETE"
}

func (q *DropCollection) generate() []byte {
	return nil
}
