package arangolite

import (
	"encoding/json"
	"strconv"
)

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

func (r *CreateDatabase) Description() string {
	return "CREATE DATABASE"
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

func (r *DropDatabase) Description() string {
	return "DROP DATABASE"
}

func (r *DropDatabase) Path() string {
	return "/_api/database/" + r.Name
}

func (r *DropDatabase) Method() string {
	return "DELETE"
}

func (r *DropDatabase) Generate() []byte {
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

func (r *CreateCollection) Description() string {
	return "CREATE COLLECTION"
}

func (r *CreateCollection) Path() string {
	return "/_api/collection"
}

func (r *CreateCollection) Method() string {
	return "POST"
}

func (r *CreateCollection) Generate() []byte {
	m, _ := json.Marshal(r)
	return m
}

// DropCollection deletes a collection in database.
type DropCollection struct {
	Name string
}

func (r *DropCollection) Description() string {
	return "DROP COLLECTION"
}

func (r *DropCollection) Path() string {
	return "/_api/collection/" + r.Name
}

func (r *DropCollection) Method() string {
	return "DELETE"
}

func (r *DropCollection) Generate() []byte {
	return nil
}

// TruncateCollection deletes a collection in database.
type TruncateCollection struct {
	Name string
}

func (r *TruncateCollection) Description() string {
	return "TRUNCATE COLLECTION"
}

func (r *TruncateCollection) Path() string {
	return "/_api/collection/" + r.Name + "/truncate"
}

func (r *TruncateCollection) Method() string {
	return "PUT"
}

func (r *TruncateCollection) Generate() []byte {
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

func (r *CreateHashIndex) Description() string {
	return "CREATE HASH INDEX"
}

func (r *CreateHashIndex) Path() string {
	return "/_api/index?collection=" + r.CollectionName
}

func (r *CreateHashIndex) Method() string {
	return "POST"
}

func (r *CreateHashIndex) Generate() []byte {
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

func (r *SetCacheProperties) Description() string {
	return "SET CACHE PROPERTIES"
}

func (r *SetCacheProperties) Path() string {
	return "/_api/query-cache/properties"
}

func (r *SetCacheProperties) Method() string {
	return "PUT"
}

func (r *SetCacheProperties) Generate() []byte {
	m, _ := json.Marshal(r)
	return m
}

// GetCacheProperties retrieves the current query cache properties.
type GetCacheProperties struct{}

func (r *GetCacheProperties) Description() string {
	return "GET CACHE PROPERTIES"
}

func (r *GetCacheProperties) Path() string {
	return "/_api/query-cache/properties"
}

func (r *GetCacheProperties) Method() string {
	return "GET"
}

func (r *GetCacheProperties) Generate() []byte {
	return nil
}

type CollectionInfo struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	IsSystem bool   `json:"isSystem"`
	Status   int    `json:"status"`
	Type     int    `json:"type"`
}

type CollectionInfoList struct {
	Collections []CollectionInfo `json:"collections"`
	Error       bool             `json:"error"`
	Code        int              `json:"code"`
}

// ListCollections lists all collections from the current DB
type ListCollections struct {
	includeSystem bool
}

func (c *ListCollections) Description() string {
	return "LIST COLLECTIONS"
}

func (c *ListCollections) Path() string {
	return "/_api/collection?excludeSystem=" + strconv.FormatBool(!c.includeSystem)
}

func (c *ListCollections) Method() string {
	return "GET"
}

func (c *ListCollections) Generate() []byte {
	return nil
}

// CollectionInfo gets information about the collection
type GetCollectionInfo struct {
	CollectionName string
	IncludeSystem  bool
}

func (c *GetCollectionInfo) Description() string {
	return "COLLECTION INFO"
}

func (c *GetCollectionInfo) Path() string {
	return "/_api/collection/" + c.CollectionName + "?excludeSystem=" + strconv.FormatBool(!c.IncludeSystem)
}

func (c *GetCollectionInfo) Method() string {
	return "GET"
}

func (c *GetCollectionInfo) Generate() []byte {
	return nil
}

// ImportCollection imports data to a collection
type ImportCollection struct {
	CollectionName string
	Data           []byte
	Update         bool
}

func (c *ImportCollection) Description() string {
	return "IMPORT COLLECTION"
}

func (c *ImportCollection) Path() string {
	if c.Update {
		return "/_api/import/?type=auto&update=true&collection=" + c.CollectionName
	}

	return "/_api/import/?type=auto&collection=" + c.CollectionName
}

func (c *ImportCollection) Method() string {
	return "POST"
}

func (c *ImportCollection) Generate() []byte {
	return c.Data
}
