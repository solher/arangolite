package arangolite

import "testing"

// TestRequests covers the graph requests
func TestGraphRequests(t *testing.T) {
	db := New().LoggerOptions(false, false, false).Connect("http://arangodb:8000", "dbName", "foo", "bar")

	db.Run(&CreateGraph{})
	db.Run(&ListGraphs{})
	db.Run(&GetGraph{})
	db.Run(&DropGraph{})

}
