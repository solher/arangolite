package requests

import (
	"encoding/json"
	"fmt"
)

// EdgeDefinition is a definition of the graph edges
type EdgeDefinition struct {
	Collection string   `json:"collection"`
	From       []string `json:"from"`
	To         []string `json:"to"`
}

// Graph represents a graph definition.
type Graph struct {
	Name string `json:"name"`
	//An array of definitions for the edges
	EdgeDefinitions []EdgeDefinition `json:"edgeDefinitions,omitempty"`
	//An array of additional vertex collections.
	OrphanCollections []string `json:"orphanCollections,omitempty"`
	ID                string   `json:"_id,omitempty"`
	Rev               string   `json:"_rev,omitempty"`
}

// GraphData is a container for data returned by a GET GRAPH request
type GraphData struct {
	Graph Graph `json:"graph"`
}

// GraphList is a container for data returned by a LIST GRAPHS request
type GraphList struct {
	Graphs []Graph `json:"graphs"`
}

// CreateGraph creates a collection in database.
type CreateGraph struct {
	Name              string           `json:"name"`
	EdgeDefinitions   []EdgeDefinition `json:"edgeDefinitions,omitempty"`
	OrphanCollections []string         `json:"orphanCollections,omitempty"`
}

func (c *CreateGraph) Path() string {
	return "/_api/gharial"
}

func (c *CreateGraph) Method() string {
	return "POST"
}

func (c *CreateGraph) Generate() []byte {
	m, _ := json.Marshal(c)
	return m
}

// GetGraph gets a graph from the graph module.
type GetGraph struct {
	Name string
}

func (g *GetGraph) Path() string {
	return fmt.Sprintf("/_api/gharial/%s", g.Name)
}

func (g *GetGraph) Method() string {
	return "GET"
}

func (g *GetGraph) Generate() []byte {
	return nil
}

// ListGraph lists all graphs known by the graph module.
type ListGraphs struct{}

func (l *ListGraphs) Path() string {
	return "/_api/gharial"
}

func (l *ListGraphs) Method() string {
	return "GET"
}

func (l *ListGraphs) Generate() []byte {
	return nil
}

// DropGraph deletes an existing graph.
type DropGraph struct {
	Name            string
	DropCollections bool
}

func (d *DropGraph) Path() string {
	return fmt.Sprintf("/_api/gharial/%s?dropCollections=%v", d.Name, d.DropCollections)
}

func (d *DropGraph) Method() string {
	return "DELETE"
}

func (d *DropGraph) Generate() []byte {
	return nil
}
