# Arangolite [![Build Status](https://travis-ci.org/solher/arangolite.svg)](https://travis-ci.org/solher/arangolite) [![Coverage Status](https://coveralls.io/repos/solher/arangolite/badge.svg?branch=master&service=github)](https://coveralls.io/github/solher/arangolite?branch=master) [![Code Climate](https://codeclimate.com/github/solher/arangolite/badges/gpa.svg)](https://codeclimate.com/github/solher/arangolite)

Arangolite is a lightweight ArangoDB driver for Go.

It focuses entirely on pure AQL querying. See [AranGO](https://github.com/diegogub/aranGO) for a more ORM-like experience.

Arangolite also features a [LoopBack](http://loopback.io/) heavily inspired filtering system.

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

	key := "47473545749"

	r, err := db.RunAQL(`
    FOR n
    IN nodes
    FILTER n._key == %s
    RETURN n
  `, key)

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
//   }
// ]
```

## Document and Edge

```go
// Document represents a basic ArangoDB document
type Document struct {
	ID  string `json:"_id,omitempty"`
	Rev string `json:"_rev,omitempty"`
	Key string `json:"_key,omitempty"`
}

// Edge represents a basic ArangoDB edge
type Edge struct {
	Document
	From string `json:"_from,omitempty"`
	To   string `json:"_to,omitempty"`
}
```

## Filter
### Overview

In a similar way than in [LoopBack](http://loopback.io/), the filtering system is API client oriented.

Its goal is to provide an easy way of converting JSON filters passed through query strings into an actual AQL query:

```go
// Filter defines a way of filtering AQL queries.
type Filter struct {
  Offset  int                    `json:"offset"`
	Limit   int                    `json:"limit"`
	Sort    []string               `json:"sort"`
	Where   map[string]interface{} `json:"where"`
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
  "where": {
    "firstName": "Pierre",
    "or": [
      {"birthPlace": ["Paris", "Los Angeles"]},
      {"age": {"gte": 18}}
    ]
  },
  "options": ["details"]
}
```

AQL:
```
FOR var IN result
LIMIT 1, 2
SORT var.age DESC, var.money ASC
FILTER var.firstName == 'Pierre' && (var.birthPlace IN ['Paris', 'Los Angeles'] || var.age >= 18)
RETURN var
```

### Operators

- `and`: Logical AND operator.
- `or`: Logical OR operator.
- `gt`, `gte`: Numerical greater than (>); greater than or equal (>=).
- `lt`, `lte`: Numerical less than (<); less than or equal (<=).

### Usage

```go
func main() {
	db := arangolite.New(true)
	db.Connect("http://localhost:8000", "testDB", "user", "password")

	q := arangolite.NewQuery(`
      FOR n
      IN nodes
      RETURN n
    `)

	filter, err := arangolite.GetFilter(`{"limit": 2}`)
	if err != nil {
		panic(err)
	}

	if err := q.Filter(filter); err != nil {
		panic(err)
	}

	r, err := q.Run(db)
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

## Known Issues

Transactions are currently not supported as in ArangoDB, they MUST be written in Javascript.

## Roadmap

- Add database and collection management.

## License

MIT
