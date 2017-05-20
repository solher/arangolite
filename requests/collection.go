package requests

import (
	"encoding/json"
	"fmt"
)

// CreateCollection creates a collection in database.
type CreateCollection struct {
	JournalSize    int                    `json:"journalSize,omitempty"`
	KeyOptions     map[string]interface{} `json:"keyOptions,omitempty"`
	Name           string                 `json:"name"`
	WaitForSync    bool                   `json:"waitForSync,omitempty"`
	DoCompact      bool                   `json:"doCompact,omitempty"`
	IsVolatile     bool                   `json:"isVolatile,omitempty"`
	ShardKeys      []string               `json:"shardKeys,omitempty"`
	NumberOfShards int                    `json:"numberOfShards,omitempty"`
	IsSystem       bool                   `json:"isSystem,omitempty"`
	Type           int                    `json:"type,omitempty"`
	IndexBuckets   int                    `json:"indexBuckets,omitempty"`
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

func (r *DropCollection) Path() string {
	return fmt.Sprintf("/_api/collection/%s", r.Name)
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

func (r *TruncateCollection) Path() string {
	return fmt.Sprintf("/_api/collection/%s/truncate", r.Name)
}

func (r *TruncateCollection) Method() string {
	return "PUT"
}

func (r *TruncateCollection) Generate() []byte {
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

func (c *ListCollections) Path() string {
	return fmt.Sprintf("/_api/collection?excludeSystem=%v", !c.includeSystem)
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

func (c *GetCollectionInfo) Path() string {
	return fmt.Sprintf("/_api/collection/%s?excludeSystem=%v", c.CollectionName, !c.IncludeSystem)
}

func (c *GetCollectionInfo) Method() string {
	return "GET"
}

func (c *GetCollectionInfo) Generate() []byte {
	return nil
}
