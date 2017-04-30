package requests_test

import (
	"testing"

	"strings"

	"github.com/solher/arangolite/requests"
)

type aqlParams struct {
	resultVar, query string
}

// TestTransaction runs tests on the Transaction request.
func TestTransaction(t *testing.T) {
	var testCases = []struct {
		// Case description
		description string
		// Arguments
		readCol, writeCol []string
		aqls              []aqlParams
		bind              map[string]interface{}
		returnVar         string
		// Expected results
		output string
	}{
		{
			description: "empty transaction",
			readCol:     []string{"foo", "bar"},
			writeCol:    []string{"bar", "foo"},
			output:      `{"collections":{"read":["foo","bar"],"write":["bar","foo"]},"action":"function () { var db = require('internal').db; }"}`,
		},
		{
			description: "simple query",
			readCol:     []string{},
			writeCol:    []string{},
			aqls: []aqlParams{
				{resultVar: "documents", query: "FOR x IN documents RETURN x"},
			},
			returnVar: "documents",
			output:    "{\"collections\":{\"read\":[],\"write\":[]},\"action\":\"function () { var db = require('internal').db; var documents = db._query(aqlQuery`FOR x IN documents RETURN x`).toArray(); return documents; }\"}",
		},
		{
			description: "simple query, no return",
			readCol:     []string{},
			writeCol:    []string{},
			aqls: []aqlParams{
				{resultVar: "", query: "FOR x IN documents RETURN x"},
			},
			output: "{\"collections\":{\"read\":[],\"write\":[]},\"action\":\"function () { var db = require('internal').db; db._query(aqlQuery`FOR x IN documents RETURN x`).toArray(); }\"}",
		},
		{
			description: "multiple queries",
			readCol:     []string{},
			writeCol:    []string{},
			aqls: []aqlParams{
				{resultVar: "documents", query: "FOR x IN documents RETURN x"},
				{resultVar: "result", query: "FOR x IN ${documents} RETURN x"},
			},
			returnVar: "result",
			output:    "{\"collections\":{\"read\":[],\"write\":[]},\"action\":\"function () { var db = require('internal').db; var documents = db._query(aqlQuery`FOR x IN documents RETURN x`).toArray(); var result = db._query(aqlQuery`FOR x IN ${documents} RETURN x`).toArray(); return result;}\"}",
		},
		{
			description: "bind variables",
			readCol:     []string{},
			writeCol:    []string{},
			bind:        map[string]interface{}{"city": "Los Angeles"},
			aqls: []aqlParams{
				{resultVar: "documents", query: `RETURN {name: @city}`},
			},
			returnVar: "documents",
			output:    "{\"collections\":{\"read\":[],\"write\":[]},\"action\":\"function () { var db = require('internal').db; var city = 'Los Angeles'; var documents = db._query(aqlQuery`RETURN {name: ${city}}`).toArray(); return documents; }\"}",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			tr := requests.NewTransaction(tc.readCol, tc.writeCol)
			for _, aql := range tc.aqls {
				tr.AddAQL(aql.resultVar, aql.query)
			}
			for name, value := range tc.bind {
				tr.Bind(name, value)
			}
			if tc.returnVar != "" {
				tr.Return(tc.returnVar)
			}

			if strings.Replace(string(tc.output), " ", "", -1) != strings.Replace(string(tr.Generate()), " ", "", -1) {
				t.Errorf("unexpected output. Expected %s, got %s", tc.output, tr.Generate())
			}
		})
	}
}
