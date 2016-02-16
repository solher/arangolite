package filters

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFromJSON runs tests on the arangolite FromJSON method.
func TestFromJSON(t *testing.T) {
	a := assert.New(t)
	r := require.New(t)

	filter, err := FromJSON(`foobar`)
	r.Error(err)
	a.Nil(filter)

	filter, err = FromJSON(`{}`)
	r.NoError(err)
	a.EqualValues(&Filter{}, filter)

	filter, err = FromJSON(`{"offset": 1, "limit": 2, "sort": ["age desc", "money"],
    "options": {"details": true}}`)
	r.NoError(err)
	a.EqualValues(&Filter{Offset: 1, Limit: 2, Sort: []string{"age desc", "money"},
		Options: map[string]interface{}{"details": true}}, filter)

	filter, err = FromJSON(`{"where": [{"age": {"gte": 18}}]}`)
	r.NoError(err)
	a.EqualValues(18, filter.Where[0]["age"].(map[string]interface{})["gte"])
}

// TestToAQL runs tests on the arangolite ToAQL method.
func TestToAQL(t *testing.T) {
	a := assert.New(t)
	r := require.New(t)

	filter, err := FromJSON(`{"offset": 1, "limit": 2, "sort": ["age desc", "money"],
    "where": [{"age": {"gte": 18}}], "options": {"details": true}}`)
	r.NoError(err)

	aqlFilter, err := ToAQL("", filter)
	r.NoError(err)
	a.EqualValues(
		`FILTER var.age >= 18 SORT var.age DESC, var.money ASC LIMIT 1, 2`,
		aqlFilter)

	aqlFilter, err = ToAQL("var", &Filter{})
	r.NoError(err)
	a.EqualValues(``, aqlFilter)

	aqlFilter, err = ToAQL("var", &Filter{Where: []map[string]interface{}{{"and": []interface{}{"foo", "bar"}}}})
	r.Error(err)
	a.EqualValues(0, len(aqlFilter))
}
