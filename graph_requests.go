package arangolite

import (
	"encoding/json"
)

// GRAPH management

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
	EdgeDefinitions []EdgeDefinition `json:"edgeDefinitions",omitempty`
	//An array of additional vertex collections.
	OrphanCollections []string `json:"orphanCollections",omitempty`
	ID                string   `json:"_id",omitempty`
	Rev               string   `json:"_rev",omitempty`
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
	EdgeDefinitions   []EdgeDefinition `json:"edgeDefinitions",omitempty`
	OrphanCollections []string         `json:"orphanCollections",omitempty`
}

func (c *CreateGraph) description() string {
	return "CREATE GRAPH"
}

func (c *CreateGraph) path() string {
	return "/_api/gharial"
}

func (c *CreateGraph) method() string {
	return "POST"
}

func (c *CreateGraph) generate() []byte {
	m, _ := json.Marshal(c)
	return m
}

// GetGraph gets a graph from the graph module.
type GetGraph struct {
	Name string
}

func (g *GetGraph) description() string {
	return "GET GRAPH"
}

func (g *GetGraph) path() string {
	return "/_api/gharial/" + g.Name
}

func (g *GetGraph) method() string {
	return "GET"
}

func (g *GetGraph) generate() []byte {
	return nil
}

// ListGraph lists all graphs known by the graph module.
type ListGraphs struct{}

func (l *ListGraphs) description() string {
	return "LIST GRAPHS"
}

func (l *ListGraphs) path() string {
	return "/_api/gharial/"
}

func (l *ListGraphs) method() string {
	return "GET"
}

func (l *ListGraphs) generate() []byte {
	return nil
}

// DropGraph deletes an existing graph.
type DropGraph struct {
	Name            string
	DropCollections bool
}

func (d *DropGraph) description() string {
	return "DROP GRAPH"
}

func (d *DropGraph) path() string {
	var queryParameters string

	if d.DropCollections {
		queryParameters = "?dropCollections=true"
	}

	return "/_api/gharial/" + d.Name + queryParameters
}

func (d *DropGraph) method() string {
	return "DELETE"
}

func (d *DropGraph) generate() []byte {
	return nil
}
