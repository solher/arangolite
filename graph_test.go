package arangolite

import (
	"testing"

	"github.com/h2non/gentleman-mock"
)

// TestRequests covers the graph requests
func TestGraphRequests(t *testing.T) {
	db := New().LoggerOptions(false, false, false).Connect("http://graph:8000", "dbName", "foo", "bar")
	db.conn.Use(mock.Plugin)
	defer mock.Disable()
	mock.New("http://graph:8000").Persist().Reply(200).BodyString("[]")

	db.Run(&CreateGraph{})
	db.Run(&ListGraphs{})
	db.Run(&GetGraph{})
	db.Run(&DropGraph{})
}
