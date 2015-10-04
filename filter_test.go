package arangolite

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetFilter runs tests on the arangolite GetFilter method.
func TestGetFilter(t *testing.T) {
	a := assert.New(t)
	r := require.New(t)

	filter, err := GetFilter(`foobar`)
	r.Error(err)
	a.Nil(filter)

	filter, err = GetFilter(`{}`)
	r.NoError(err)
	a.EqualValues(&Filter{}, filter)

	filter, err = GetFilter(`{"offset": 1, "limit": 2, "sort": ["age desc", "money"],
    "options": ["details"]}`)
	r.NoError(err)
	a.EqualValues(&Filter{Offset: 1, Limit: 2, Sort: []string{"age desc", "money"},
		Options: []string{"details"}}, filter)

	filter, err = GetFilter(`{"where": {"age": {"gte": 18}}}`)
	r.NoError(err)
	a.EqualValues(18, filter.Where["age"].(map[string]interface{})["gte"])
}

// TestGetAQLFilter runs tests on the arangolite GetAQLFilter method.
func TestGetAQLFilter(t *testing.T) {
	a := assert.New(t)
	r := require.New(t)

	filter, err := GetFilter(`{"offset": 1, "limit": 2, "sort": ["age desc", "money"],
    "where": {"age": {"gte": 18}}, "options": ["details"]}`)
	r.NoError(err)

	aqlFilter, err := GetAQLFilter(filter)
	r.NoError(err)
	a.EqualValues(
		`FOR var IN result LIMIT 1, 2 SORT var.age DESC, var.money ASC FILTER var.age >= 18 RETURN var`,
		aqlFilter)

	aqlFilter, err = GetAQLFilter(&Filter{})
	r.NoError(err)
	a.EqualValues(`FOR var IN result RETURN var`, aqlFilter)

	aqlFilter, err = GetAQLFilter(&Filter{Where: map[string]interface{}{"and": []interface{}{"foo", "bar"}}})
	r.Error(err)
	a.EqualValues(0, len(aqlFilter))
}
