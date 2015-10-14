# Arangolite [![Build Status](https://travis-ci.org/solher/arangolite.svg)](https://travis-ci.org/solher/arangolite) [![Coverage Status](https://coveralls.io/repos/solher/arangolite/badge.svg?branch=master&service=github)](https://coveralls.io/github/solher/arangolite?branch=master) [![Code Climate](https://codeclimate.com/github/solher/arangolite/badges/gpa.svg)](https://codeclimate.com/github/solher/arangolite)

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
  db := arangolite.New(true)
  db.Connect("http://localhost:8000", "testDB", "user", "password")

  key := "48765564346"

  r, err := arangolite.NewQuery(`
    FOR n
    IN nodes
    FILTER n._key == %s
    RETURN n
  `, key).Cache(true).Run(db) // The caching feature is unavailable prior to ArangoDB 2.7

  if err != nil {
    panic(err)
  }

  nodes := []Node{}

  if err = json.Unmarshal(r, &nodes); err != nil {
    panic(err)
  }

  fmt.Printf("%v", nodes)
}

// OUTPUT:
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
  db := arangolite.New(true)
  db.Connect("http://localhost:8000", "testDB", "user", "password")

  r, err := arangolite.NewTransaction([]string{"nodes"}, nil).
    AddQuery("nodes", `
    FOR n
    IN nodes
    RETURN n
  `).AddQuery("ids", `
    FOR n
    IN {{.nodes}}
    RETURN n._id
  `).Return("ids").Run(db)

  if err != nil {
    panic(err)
  }

  ids := []string{}

  if err = json.Unmarshal(r, &ids); err != nil {
    panic(err)
  }

  fmt.Printf("%v", ids)
}
```

## Filters
### Overview

In a similar way than in [LoopBack](http://loopback.io/), the filtering system is API client oriented.

Its goal is to provide an easy way of converting JSON filters passed through query strings into an actual AQL query:

```go
// Filter defines a way of filtering AQL queries.
type Filter struct {
  Offset  int                    `json:"offset"`
  Limit   int                    `json:"limit"`
  Sort    []string               `json:"sort"`
  Where   []map[string]interface{} `json:"where"`
  Options []string               `json:"options"`
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

  r, err := arangolite.NewQuery(`
    FOR n
    IN nodes
    %s
    RETURN n
  `, aqlFilter).Run(db)

  if err != nil {
    panic(err)
  }

  nodes := []Node{}

  if err = json.Unmarshal(r, &nodes); err != nil {
    panic(err)
  }

  fmt.Printf("%v", nodes)
}

// OUTPUT:
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

## Roadmap

- Add database and collection management.

## License

MIT
