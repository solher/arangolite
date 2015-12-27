# Arangolite [![Build Status](https://travis-ci.org/solher/arangolite.svg?branch=master)](https://travis-ci.org/solher/arangolite) [![Coverage Status](https://coveralls.io/repos/solher/arangolite/badge.svg?branch=master&service=github)](https://coveralls.io/github/solher/arangolite?branch=master) [![Code Climate](https://codeclimate.com/github/solher/arangolite/badges/gpa.svg)](https://codeclimate.com/github/solher/arangolite)

Arangolite is a lightweight ArangoDB driver for Go.

It focuses entirely on pure AQL querying. See [AranGO](https://github.com/diegogub/aranGO) for a more ORM-like experience.

Arangolite also features a [LoopBack](http://loopback.io/) heavily inspired filtering system, in a separated package so you don't need to import it if you don't use it.

## Installation

To install Arangolite:

    go get github.com/solher/arangolite

## Basic Usage

```go
package main

import (
  "encoding/json"
  "fmt"

  "github.com/solher/arangolite"
)

type Node struct {
  arangolite.Document
}

func main() {
  db := arangolite.New().
    LoggerOptions(false, false, false).
    Connect("http://localhost:8000", "_system", "root", "rootPassword")

  _, _ := db.Run(&arangolite.CreateDatabase{
		Name: "testDB",
		Users: []map[string]interface{}{
			{"username": "root", "passwd": "rootPassword"},
			{"username": "user", "passwd": "password"},
		},
	})

  db.SwitchDatabase("testDB").SwitchUser("user", "password")

  _, _ := db.Run(&arangolite.CreateCollection{Name: "nodes"})

  key := "48765564346"

  q := arangolite.NewQuery(`
    FOR n
    IN nodes
    FILTER n._key == %s
    RETURN n
  `, key).Cache(true).BatchSize(500) // The caching feature is unavailable prior to ArangoDB 2.7

  // The Run method returns all the query results of every batches
  // available in the cursor as a slice of byte.
  r, _ := db.Run(q)

  nodes := []Node{}
  json.Unmarshal(r, &nodes)

  // The RunAsync method returns a Result struct allowing to handle batches as they
  // are retrieved from the database.
  async, _ := db.RunAsync(q)

  nodes = []Node{}
  decoder := json.NewDecoder(async.Buffer())

  for async.HasMore() {
    batch := []Node{}
    decoder.Decode(&batch)
    nodes = append(nodes, batch...)
  }

  fmt.Printf("%v", nodes)
}

// OUTPUT EXAMPLE:
// [
//   {
//     "_id": "nodes/48765564346",
//     "_rev": "48765564346",
//     "_key": "48765564346"
//   }
// ]
```

## Document and Edge

```go
// Document represents a basic ArangoDB document
// Fields are pointers to allow null values in ArangoDB
type Document struct {
  ID  *string `json:"_id,omitempty"`
  Rev *string `json:"_rev,omitempty"`
  Key *string `json:"_key,omitempty"`
}

// Edge represents a basic ArangoDB edge
// Fields are pointers to allow null values in ArangoDB
type Edge struct {
  Document
  From *string `json:"_from,omitempty"`
  To   *string `json:"_to,omitempty"`
}
```

## Transactions
### Overview

Arangolite provides an abstraction layer to the Javascript ArangoDB transactions.

The only limitation is that no Javascript processing can be manually added inside the transaction. The queries can only be connected in a raw way, using the Go templating conventions.

### Usage

```go
func main() {
  db := arangolite.New()
  db.Connect("http://localhost:8000", "testDB", "user", "password")

  t := arangolite.NewTransaction([]string{"nodes"}, nil).
  AddQuery("nodes", `
    FOR n
    IN nodes
    RETURN n
  `).AddQuery("ids", `
    FOR n
    IN {{.nodes}}
    RETURN n._id
  `).Return("ids")

  r, _ := db.Run(t)

  ids := []string{}
  json.Unmarshal(r, &ids)

  fmt.Printf("%v", ids)
}
```

## Graphs
### Overview
AQL may be used for querying graph data. But to manage graphs, aroangolite offers a few specific requests:
* CreateGraph to create a graph
* ListGraphs to list available graphs
* GetGraph to get an existing graph
* DropGraph to delete a graph

### Usage
```go
  db.Run(&arangolite.CreateCollection{Name: "CollectionName"})
  db.Run(&arangolite.CreateCollection{Name: "RelationshipCollectionName", Type: 3})

  // check graph existence
  _, err := db.Run(&arangolite.GetGraph{Name: "GraphName"})

  // if graph does not exist, create a new one
  if err != nil {
    from := make([]string, 1)
    from[0] = "FirstCollectionName"
    to := make([]string, 1)
    to[0] = "SecondCollectionName"

    edgeDefinition := arangolite.EdgeDefinition{Collection: "EdgeCollectionName", From: from, To: to}
    edgeDefinitions := make([]arangolite.EdgeDefinition, 1)
    edgeDefinitions[0] = edgeDefinition
    db.Run(&arangolite.CreateGraph{Name: "GraphName", EdgeDefinitions: edgeDefinitions})
  }

  // grab the graph
  graphBytes, _ := config.DB().Run(&arangolite.GetGraph{Name: "GraphName"})
  graph := &arangolite.GraphData{}
  json.Unmarshal(graphBytes, &graph)
  fmt.Printf("Graph: %+v", graph)

  // list existing graphs
  listBytes, _ :=  db.Run(&arangolite.ListGraphs{})
  list := &arangolite.GraphList{}
  json.Unmarshal(listBytes, &list)
  fmt.Printf("Graph list: %+v", list)

  // destroy the graph we just created, and the related collections
  db.Run(&arangolite.DropGraph{Name: "GraphName", DropCollections: true})

```

## Filters
### Overview

In a similar way than in [LoopBack](http://loopback.io/), the filtering system is API client oriented.

Its goal is to provide an easy way of converting JSON filters passed through query strings into an actual AQL query:

```go
// Filter defines a way of filtering AQL queries.
type Filter struct {
  Offset  int                      `json:"offset"`
  Limit   int                      `json:"limit"`
  Sort    []string                 `json:"sort"`
  Where   []map[string]interface{} `json:"where"`
  Options []string                 `json:"options"`
}
```

### Options Field

The `Options` field implementation is left to the developer.
It is not translated into AQL during the filtering.

Its main goal is to allow a filtering similar to the `Include` one in traditional ORMs, as a relation can be a join or a edge in ArangoDB.

Of course, the `Options` field can also be used as a more generic option selector (*e.g.*, `Options: "Basic"` to only return the basic info about a resource).

### Translation example

JSON:
```json
{
  "offset": 1,
  "limit": 2,
  "sort": ["age desc", "money"],
  "where": [
    {"firstName": "Pierre"},
    {
      "or": [
        {"birthPlace": ["Paris", "Los Angeles"]},
        {"age": {"gte": 18}}
      ]
    }
  ]
  },
  "options": ["details"]
}
```

AQL:
```
LIMIT 1, 2
SORT var.age DESC, var.money ASC
FILTER var.firstName == 'Pierre' && (var.birthPlace IN ['Paris', 'Los Angeles'] || var.age >= 18)
```

### Operators

- `and`: Logical AND operator.
- `or`: Logical OR operator.
- `not`: Logical NOT operator.
- `gt`, `gte`: Numerical greater than (>); greater than or equal (>=).
- `lt`, `lte`: Numerical less than (<); less than or equal (<=).
- `eq`, `neq`: Equal (==); non equal (!=).

### Usage

```go
func main() {
  db := arangolite.New(true)
  db.Connect("http://localhost:8000", "testDB", "user", "password")

  filter, err := filters.FromJSON(`{"limit": 2}`)
  if err != nil {
    panic(err)
  }

  aqlFilter, err := filters.ToAQL("n", filter)
  if err != nil {
    panic(err)
  }

  r, _ := db.Run(arangolite.NewQuery(`
    FOR n
    IN nodes
    %s
    RETURN n
  `, aqlFilter))

  nodes := []Node{}
  json.Unmarshal(r, &nodes)

  fmt.Printf("%v", nodes)
}

// OUTPUT EXAMPLE:
// [
//   {
//     "_id": "nodes/47473545749",
//     "_rev": "47473545749",
//     "_key": "47473545749"
//   },
//   {
//     "_id": "nodes/47472824853",
//     "_rev": "47472824853",
//     "_key": "47472824853"
//   }
// ]
```

## Contributing

Currently, very few methods of the ArangoDB HTTP API are implemented in Arangolite.
Fortunately, it is really easy to add your own. There are two ways:

- Using the `Send` method, passing a struct representing the request you want to send.

```go
func (db *DB) Send(description, method, path string, req interface{}) ([]byte, error) {}
```

- Implementing the `Runnable` interface. You can then use the regular `Run` method.

```go
// Runnable defines requests runnable by the Run and RunAsync methods.
// Queries, transactions and everything in the requests.go file are Runnable.
type Runnable interface {
	description() string // Description shown in the logger
	generate() []byte // The body of the request
	path() string // The path where to send the request
	method() string // The HTTP method to use
}
```

**Please pull request in the requests.go file when you implement some new features so everybody can use it.**

## License

MIT
