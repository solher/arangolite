package arangolite

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var offsetFilter = &Filter{
	Offset: 1,
}

var limitFilter = &Filter{
	Limit: 2,
}

var offsetLimitFilter = &Filter{
	Offset: 3,
	Limit:  4,
}

var sortFilter = &Filter{
	Sort: []string{"firstName ASC", "lastName dESc", "age"},
}

var basicWhereFilter = &Filter{
	Where: map[string]interface{}{
		"password":   "qwertyuiop",
		"age":        22,
		"money":      3000.55,
		"awesome":    true,
		"graduated":  []int{2010, 2015},
		"avg":        []float64{15.5, 13.24},
		"birthPlace": []interface{}{"Chalon", "Macon"},
	},
}

var orWhereFilter = &Filter{
	Where: map[string]interface{}{
		"oR": []interface{}{
			map[string]interface{}{"lastName": map[string]interface{}{"eq": "Fabien"}},
			map[string]interface{}{"age": map[string]interface{}{"gt": 23}},
			map[string]interface{}{"age": map[string]interface{}{"lt": 26}},
		},
	},
}

var andWhereFilter = &Filter{
	Where: map[string]interface{}{
		"and": []interface{}{
			map[string]interface{}{"firstName": map[string]interface{}{"neq": "Toto"}},
			map[string]interface{}{"money": 200.5},
		},
	},
}

var notWhereFilter = &Filter{
	Where: map[string]interface{}{
		"not": map[string]interface{}{"firstName": "Fabien"},
		"nOt": map[string]interface{}{
			"or": []interface{}{
				map[string]interface{}{"lastName": "Herfray"},
				map[string]interface{}{"money": map[string]interface{}{"gte": 0}},
				map[string]interface{}{"money": map[string]interface{}{"lte": 1000.5}}},
		},
	},
}

var pluckFilter = &Filter{
	Pluck: "_id",
}

// TestProcessFilter runs tests on the arangolite processFilter method.
func TestProcessFilter(t *testing.T) {
	a := assert.New(t)
	r := require.New(t)

	// Offset and limit filters
	p, err := processFilter(offsetFilter)
	r.NoError(err)
	a.Equal("1", p.OffsetLimit)

	p, err = processFilter(limitFilter)
	r.NoError(err)
	a.Equal("2", p.OffsetLimit)

	p, err = processFilter(offsetLimitFilter)
	r.NoError(err)
	a.Equal("3, 4", p.OffsetLimit)

	p, err = processFilter(&Filter{Offset: -1})
	a.Error(err)
	a.Nil(p)

	p, err = processFilter(&Filter{Limit: -1})
	a.Error(err)
	a.Nil(p)

	// Sort filter
	p, err = processFilter(sortFilter)
	r.NoError(err)
	a.Equal("firstName ASC, lastName DESC, age ASC", p.Sort)

	p, err = processFilter(&Filter{Sort: []string{}})
	r.NoError(err)
	a.Equal("", p.Sort)

	// Where filter
	p, err = processFilter(basicWhereFilter)
	r.NoError(err)
	split := strings.Split(p.Where, " && ")
	a.Equal(7, len(split))
	expected := []string{
		`awesome == true`,
		`graduated IN [2010, 2015]`,
		`avg IN [15.5, 13.24]`,
		`birthPlace IN ['Chalon', 'Macon']`,
		`password == 'qwertyuiop'`,
		`age == 22`,
		`money == 3000.55`,
	}
	for _, s := range split {
		a.Contains(expected, s)
	}

	p, err = processFilter(orWhereFilter)
	r.NoError(err)
	a.Equal(`(lastName == 'Fabien' || age > 23 || age < 26)`, p.Where)

	p, err = processFilter(andWhereFilter)
	r.NoError(err)
	a.Equal(`(firstName != 'Toto' && money == 200.5)`, p.Where)

	p, err = processFilter(notWhereFilter)
	r.NoError(err)
	a.Equal(`!(firstName == 'Fabien') && !((lastName == 'Herfray' || money >= 0 || money <= 1000.5))`, p.Where)

	// Pluck filter
	p, err = processFilter(pluckFilter)
	r.NoError(err)
	a.Equal("_id", p.Pluck)

	p, err = processFilter(&Filter{Pluck: "foo, bar"})
	a.Error(err)
	a.Nil(p)
}
