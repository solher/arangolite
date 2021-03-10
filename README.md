# Arangolite [![Build Status](https://travis-ci.org/solher/arangolite.svg?branch=master)](https://travis-ci.org/solher/arangolite) [![Coverage Status](https://coveralls.io/repos/solher/arangolite/badge.svg?branch=master&service=github)](https://coveralls.io/github/solher/arangolite?branch=master) [![Code Climate](https://codeclimate.com/github/solher/arangolite/badges/gpa.svg)](https://codeclimate.com/github/solher/arangolite) [![GoDoc](https://godoc.org/github.com/solher/arangolite?status.svg)](https://godoc.org/github.com/solher/arangolite)

Arangolite is a lightweight ArangoDB driver for Go.

It focuses on pure AQL querying. See [AranGO](https://github.com/diegogub/aranGO) for a more ORM-like experience.

## IMPORTANT: Looking for maintainers

I don't have as much time as I used to have and I am not as a frequent user of ArangoDB as I used to be. This project would definitely benefit from some new maintainers. Any PR is greatly appreciated.

## Installation

To install Arangolite:

    go get -u github.com/solher/arangolite/v2

## Basic Usage

```go
package main

import (
  "context"
  "fmt"
  "log"

  "github.com/solher/arangolite/v2"
  "github.com/solher/arangolite/v2/requests"
)

type Node struct {
  arangolite.Document
}

func main() {
  ctx := context.Background()

  // We declare the database definition.
  db := arangolite.NewDatabase(
    arangolite.OptEndpoint("http://localhost:8529"),
    arangolite.OptBasicAuth("root", "rootPassword"),
    arangolite.OptDatabaseName("_system"),
  )

  // The Connect method does two things:
  // - Initializes the connection if needed (JWT authentication).
  // - Checks the database connectivity.
  if err := db.Connect(ctx); err != nil {
    log.Fatal(err)
  }

  // We create a new database.
  err := db.Run(ctx, nil, &requests.CreateDatabase{
    Name: "testDB",
    Users: []map[string]interface{}{
      {"username": "root", "passwd": "rootPassword"},
      {"username": "user", "passwd": "password"},
    },
  })
  if err != nil {
    log.Fatal(err)
  }

  // We sign in as the new created user on the new database.
  // We could eventually rerun a "db.Connect()" to confirm the connectivity.
  db.Options(
    arangolite.OptBasicAuth("user", "password"),
    arangolite.OptDatabaseName("testDB"),
  )

  // We create a new "nodes" collection.
  if err := db.Run(ctx, nil, &requests.CreateCollection{Name: "nodes"}); err != nil {
    log.Fatal(err)
  }

  // We declare a new AQL query with options and bind parameters.
  key := "48765564346"
  r := requests.NewAQL(`
    FOR n
    IN nodes
    FILTER n._key == @key
    RETURN n
  `, key).
    Bind("key", key).
    Cache(true).
    BatchSize(500) // The caching feature is unavailable prior to ArangoDB 2.7

  // The Run method returns all the query results of every pages
  // available in the cursor and unmarshal it into the given struct.
  // Cancelling the context cancels every running request. 
  nodes := []Node{}
  if err := db.Run(ctx, &nodes, r); err != nil {
    log.Fatal(err)
  }

  // The Send method gives more control to the user and doesn't follow an eventual cursor.
  // It returns a raw result object.
  result, err := db.Send(ctx, r)
  if err != nil {
    log.Fatal(err)
  }
  nodes = []Node{}
  result.UnmarshalResult(&nodes)

  for result.HasMore() {
    result, err = db.Send(ctx, &requests.FollowCursor{Cursor: result.Cursor()})
    if err != nil {
      log.Fatal(err)
    }
    tmp := []Node{}
    result.UnmarshalResult(&tmp)

    nodes = append(nodes, tmp...)
  }

  fmt.Println(nodes)
}
```

## Document and Edge

```go
// Document represents a basic ArangoDB document
type Document struct {
  // The document handle. Format: ':collection/:key'
  ID string `json:"_id,omitempty"`
  // The document's revision token. Changes at each update.
  Rev string `json:"_rev,omitempty"`
  // The document's unique key.
  Key string `json:"_key,omitempty"`
}

// Edge represents a basic ArangoDB edge
type Edge struct {
  Document
  // Reference to another document. Format: ':collection/:key'
  From string `json:"_from,omitempty"`
  // Reference to another document. Format: ':collection/:key'
  To string `json:"_to,omitempty"`
}
```

## Transactions
### Overview

Arangolite provides an abstraction layer to the Javascript ArangoDB transactions.

The only limitation is that no Javascript processing can be manually added inside the transaction. The queries can be connected using the Go templating conventions.

### Usage

```go
t := requests.NewTransaction([]string{"nodes"}, nil).
  AddAQL("nodes", `
    FOR n
    IN nodes
    RETURN n
`).
  AddAQL("ids", `
    FOR n
    IN {{.nodes}}
    RETURN n._id
`).
  Return("ids")

ids := []string{}
if err := db.Run(ctx, ids, t); err != nil {
  log.Fatal(err)
}
```

## Graphs
### Overview

AQL may be used for querying graph data. But to manage graphs, Arangolite offers a few specific requests:

- `CreateGraph` to create a graph.
- `ListGraphs` to list available graphs.
- `GetGraph` to get an existing graph.
- `DropGraph` to delete a graph.

### Usage

```go
// Check graph existence.
if err := db.Run(ctx, nil, &requests.GetGraph{Name: "graphName"}); err != nil {
  switch {
  case arangolite.IsErrNotFound(err):
    // If graph does not exist, create a new one.
    edgeDefinitions := []requests.EdgeDefinition{
      {
        Collection: "edgeCollectionName",
        From:       []string{"firstCollectionName"},
        To:         []string{"secondCollectionName"},
      },
    }
    db.Run(ctx, nil, &requests.CreateGraph{Name: "graphName", EdgeDefinitions: edgeDefinitions})
  default:
    log.Fatal(err)
  }
}

// List existing graphs.
list := &requests.GraphList{}
db.Run(ctx, list, &requests.ListGraphs{})
fmt.Printf("Graph list: %v\n", list)

// Destroy the graph we just created, and the related collections.
db.Run(ctx, nil, &requests.DropGraph{Name: "graphName", DropCollections: true})
```

## Error Handling

Errors can be handled using the provided basic testers:

```go
// IsErrInvalidRequest returns true when the database returns a 400.
func IsErrInvalidRequest(err error) bool {
  return HasStatusCode(err, 400)
}

// IsErrUnauthorized returns true when the database returns a 401.
func IsErrUnauthorized(err error) bool {
  return HasStatusCode(err, 401)
}

// IsErrForbidden returns true when the database returns a 403.
func IsErrForbidden(err error) bool {
  return HasStatusCode(err, 403)
}

// IsErrUnique returns true when the error num is a 1210 - ERROR_ARANGO_UNIQUE_CONSTRAINT_VIOLATED.
func IsErrUnique(err error) bool {
  return HasErrorNum(err, 1210)
}

// IsErrNotFound returns true when the database returns a 404 or when the error num is:
// 1202 - ERROR_ARANGO_DOCUMENT_NOT_FOUND
// 1203 - ERROR_ARANGO_COLLECTION_NOT_FOUND
func IsErrNotFound(err error) bool {
  return HasStatusCode(err, 404) || HasErrorNum(err, 1202, 1203)
}

```

Or manually via the `HasStatusCode` and `HasErrorNum` methods.

## Contributing

Currently, very few methods of the ArangoDB HTTP API are implemented in Arangolite.
Fortunately, it is really easy to add your own by implementing the `Runnable` interface.
You can then use the regular `Run` and `Send` methods.

```go
// Runnable defines requests runnable by the Run and Send methods.
// A Runnable library is located in the 'requests' package.
type Runnable interface {
  // The body of the request.
  Generate() []byte
  // The path where to send the request.
  Path() string
  // The HTTP method to use.
  Method() string
}
```

**Please pull request when you implement some new features so everybody can use it.**

## License

MIT
