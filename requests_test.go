package arangolite

import "testing"

// TestRequests covers the requests
func TestRequests(t *testing.T) {
	db := New().LoggerOptions(false, false, false).Connect("http://arangodb:8000", "dbName", "foo", "bar")

	// DATABASE
	db.Run(&CreateDatabase{})
	db.Run(&DropDatabase{})

	// COLLECTION
	db.Run(&CreateCollection{})
	db.Run(&DropCollection{})
	db.Run(&TruncateCollection{})

	// INDEX
	db.Run(&CreateHashIndex{})

	// CACHE
	db.Run(&SetCacheProperties{})
	db.Run(&GetCacheProperties{})
}
