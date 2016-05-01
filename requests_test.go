package arangolite

import (
	"testing"

	"github.com/h2non/gentleman-mock"
)

// TestRequests covers the requests
func TestRequests(t *testing.T) {
	db := New().LoggerOptions(false, false, false).Connect("http://requests:8000", "dbName", "foo", "bar")
	db.conn.Use(mock.Plugin)
	defer mock.Disable()
	mock.New("http://requests:8000").Persist().Reply(200).BodyString("[]")

	// DATABASE
	db.Run(&CreateDatabase{})
	db.Run(&DropDatabase{})

	// COLLECTION
	db.Run(&CreateCollection{})
	db.Run(&DropCollection{})
	db.Run(&TruncateCollection{})
	db.Run(&ListCollections{})
	db.Run(&GetCollectionInfo{})
	db.Run(&ImportCollection{})

	// INDEX
	db.Run(&CreateHashIndex{})

	// CACHE
	db.Run(&SetCacheProperties{})
	db.Run(&GetCacheProperties{})

}
