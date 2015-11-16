package arangolite

import "encoding/json"

// DATABASE

// CreateDatabase creates a new database.
type CreateDatabase struct {
	Username string                   `json:"username,omitempty"`
	Name     string                   `json:"name"`
	Extra    json.RawMessage          `json:"extra,omitempty"`
	Passwd   string                   `json:"passwd,omitempty"`
	Active   *bool                    `json:"active,omitempty"`
	Users    []map[string]interface{} `json:"users,omitempty"`
}

func (r *CreateDatabase) description() string {
	return "CREATE DATABASE"
}

func (r *CreateDatabase) path() string {
	return "/_api/database"
}

func (r *CreateDatabase) method() string {
	return "POST"
}

func (r *CreateDatabase) generate() []byte {
	m, _ := json.Marshal(r)
	return m
}

// DropDatabase deletes a database.
type DropDatabase struct {
	Name string
}

func (r *DropDatabase) description() string {
	return "DROP DATABASE"
}

func (r *DropDatabase) path() string {
	return "/_api/database/" + r.Name
}

func (r *DropDatabase) method() string {
	return "DELETE"
}

func (r *DropDatabase) generate() []byte {
	return nil
}

// COLLECTION

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

func (r *CreateCollection) description() string {
	return "CREATE COLLECTION"
}

func (r *CreateCollection) path() string {
	return "/_api/collection"
}

func (r *CreateCollection) method() string {
	return "POST"
}

func (r *CreateCollection) generate() []byte {
	m, _ := json.Marshal(r)
	return m
}

// DropCollection deletes a collection in database.
type DropCollection struct {
	Name string
}

func (r *DropCollection) description() string {
	return "DROP COLLECTION"
}

func (r *DropCollection) path() string {
	return "/_api/collection/" + r.Name
}

func (r *DropCollection) method() string {
	return "DELETE"
}

func (r *DropCollection) generate() []byte {
	return nil
}

// TruncateCollection deletes a collection in database.
type TruncateCollection struct {
	Name string
}

func (r *TruncateCollection) description() string {
	return "TRUNCATE COLLECTION"
}

func (r *TruncateCollection) path() string {
	return "/_api/collection/" + r.Name + "/truncate"
}

func (r *TruncateCollection) method() string {
	return "PUT"
}

func (r *TruncateCollection) generate() []byte {
	return nil
}

// INDEX

// CreateHashIndex creates a hash index in database.
type CreateHashIndex struct {
	CollectionName string   `json:"-"`
	Fields         []string `json:"fields,omitempty"`
	Unique         *bool    `json:"unique,omitempty"`
	Type           string   `json:"type,omitempty"`
	Sparse         *bool    `json:"sparse,omitempty"`
}

func (r *CreateHashIndex) description() string {
	return "CREATE HASH INDEX"
}

func (r *CreateHashIndex) path() string {
	return "/_api/index?collection-name=" + r.CollectionName
}

func (r *CreateHashIndex) method() string {
	return "POST"
}

func (r *CreateHashIndex) generate() []byte {
	r.Type = "hash"
	m, _ := json.Marshal(r)
	return m
}

// CACHE

// SetCacheProperties sets the query cache properties.
type SetCacheProperties struct {
	Mode       string `json:"mode,omitempty"`
	MaxResults int    `json:"maxResults,omitempty"`
}

func (r *SetCacheProperties) description() string {
	return "SET CACHE PROPERTIES"
}

func (r *SetCacheProperties) path() string {
	return "/_api/query-cache/properties"
}

func (r *SetCacheProperties) method() string {
	return "PUT"
}

func (r *SetCacheProperties) generate() []byte {
	m, _ := json.Marshal(r)
	return m
}

// GetCacheProperties retrieves the current query cache properties.
type GetCacheProperties struct{}

func (r *GetCacheProperties) description() string {
	return "GET CACHE PROPERTIES"
}

func (r *GetCacheProperties) path() string {
	return "/_api/query-cache/properties"
}

func (r *GetCacheProperties) method() string {
	return "GET"
}

func (r *GetCacheProperties) generate() []byte {
	return nil
}
