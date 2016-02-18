package filters

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

// integers are converted to float64 because that is what the json unmarshaller do
var basicWhereFilter = &Filter{
	Where: []map[string]interface{}{{
		"password":   "qwertyuiop",
		"age":        float64(22),
		"money":      3000.55,
		"awesome":    true,
		"notAwesome": false,
		"graduated":  []interface{}{float64(2010), float64(2015)},
		"avg":        []interface{}{15.5, 13.24},
		"birthPlace": []interface{}{"Chalon", "Macon"},
		"bools":      []interface{}{true, false},
	}},
}

var orWhereFilter = &Filter{
	Where: []map[string]interface{}{{
		"oR": []interface{}{
			map[string]interface{}{"lastName": map[string]interface{}{"eq": "Fabien"}},
			map[string]interface{}{"age": map[string]interface{}{"gt": float64(23)}},
			map[string]interface{}{"age": map[string]interface{}{"lt": float64(26)}},
		}},
	},
}

var andWhereFilter = &Filter{
	Where: []map[string]interface{}{{
		"and": []interface{}{
			map[string]interface{}{"firstName": map[string]interface{}{"neq": "Toto"}},
			map[string]interface{}{"money": 200.5},
		}},
	},
}

var notWhereFilter = &Filter{
	Where: []map[string]interface{}{
		{"not": map[string]interface{}{"firstName": "Fabien"}},
		{"nOt": map[string]interface{}{
			"or": []interface{}{
				map[string]interface{}{"lastName": "Herfray"},
				map[string]interface{}{"money": map[string]interface{}{"gte": float64(0)}},
				map[string]interface{}{"money": map[string]interface{}{"lte": 1000.5}}},
		}},
	},
}

var likeWhereFilter = &Filter{
	Where: []map[string]interface{}{{
		"like": map[string]interface{}{
			"text":             "firstName",
			"search":           "fab%",
			"case_insensitive": true,
		},
	}},
}

// TestProcessFilter runs tests on the filter processor Process method.
func TestProcessFilter(t *testing.T) {
	a := assert.New(t)
	r := require.New(t)
	fp := newFilterProcessor("")

	// Offset and limit filters
	p, err := fp.Process(offsetFilter)
	r.NoError(err)
	a.Equal("1", p.OffsetLimit)

	p, err = fp.Process(limitFilter)
	r.NoError(err)
	a.Equal("2", p.OffsetLimit)

	p, err = fp.Process(offsetLimitFilter)
	r.NoError(err)
	a.Equal("3, 4", p.OffsetLimit)

	p, err = fp.Process(&Filter{Offset: -1})
	r.NoError(err)
	a.NotNil(p)

	p, err = fp.Process(&Filter{Limit: -1})
	r.NoError(err)
	a.NotNil(p)

	// Sort filter
	p, err = fp.Process(sortFilter)
	r.NoError(err)
	a.Equal("var.firstName ASC, var.lastName DESC, var.age ASC", p.Sort)

	p, err = fp.Process(&Filter{Sort: []string{}})
	r.NoError(err)
	a.Equal("", p.Sort)

	p, err = fp.Process(&Filter{Sort: []string{"foo, bar"}})
	r.Error(err)
	a.Nil(p)

	p, err = fp.Process(&Filter{Sort: []string{"INSeRT ASC"}})
	r.Error(err)
	a.Nil(p)

	// Where filter
	p, err = fp.Process(basicWhereFilter)
	r.NoError(err)
	split := strings.Split(p.Where, " && ")
	a.Equal(9, len(split))
	expected := []string{
		`var.awesome == true`,
		`var.graduated IN [2010, 2015]`,
		`var.avg IN [15.5, 13.24]`,
		`var.birthPlace IN ['Chalon', 'Macon']`,
		`var.password == 'qwertyuiop'`,
		`var.age == 22`,
		`var.money == 3000.55`,
		`var.notAwesome == false`,
		`var.bools IN [true, false]`,
	}
	for _, s := range split {
		a.Contains(expected, s)
	}

	p, err = fp.Process(orWhereFilter)
	r.NoError(err)
	a.Equal(`(var.lastName == 'Fabien' || var.age > 23 || var.age < 26)`, p.Where)

	p, err = fp.Process(andWhereFilter)
	r.NoError(err)
	a.Equal(`(var.firstName != 'Toto' && var.money == 200.5)`, p.Where)

	p, err = fp.Process(notWhereFilter)
	r.NoError(err)
	split = strings.Split(p.Where, " && ")
	a.Equal(2, len(split))
	expected = []string{
		`!(var.firstName == 'Fabien')`,
		`!((var.lastName == 'Herfray' || var.money >= 0 || var.money <= 1000.5))`,
	}
	for _, s := range split {
		a.Contains(expected, s)
	}

	p, err = fp.Process(likeWhereFilter)
	r.NoError(err)
	a.Contains(`LIKE(var.firstName, 'fab%', true)`, p.Where)

	p, err = fp.Process(&Filter{Where: []map[string]interface{}{{"var.firstName": []interface{}{"foo", map[string]interface{}{"foo": "bar"}}}}})
	r.Error(err)
	a.Nil(p)

	p, err = fp.Process(&Filter{Where: []map[string]interface{}{{"and": []interface{}{"foo", "bar"}}}})
	r.Error(err)
	a.Nil(p)

	p, err = fp.Process(&Filter{Where: []map[string]interface{}{{"or": []interface{}{"foo", "bar"}}}})
	r.Error(err)
	a.Nil(p)

	p, err = fp.Process(&Filter{Where: []map[string]interface{}{{"and": map[string]interface{}{"var.firstName": "Fabien", "foo": "bar"}}}})
	r.Error(err)
	a.Nil(p)

	p, err = fp.Process(&Filter{Where: []map[string]interface{}{{"and": []interface{}{"INSeRT"}}}})
	r.Error(err)
	a.Nil(p)

	p, err = fp.Process(&Filter{Where: []map[string]interface{}{{"money": map[string]interface{}{"eq": 1}}}})
	r.Error(err)
	a.Nil(p)

	p, err = fp.Process(&Filter{Where: []map[string]interface{}{{"money": map[string]interface{}{"neq": 1}}}})
	r.Error(err)
	a.Nil(p)

	p, err = fp.Process(&Filter{Where: []map[string]interface{}{{"money": map[string]interface{}{"gt": 1}}}})
	r.Error(err)
	a.Nil(p)

	p, err = fp.Process(&Filter{Where: []map[string]interface{}{{"money": map[string]interface{}{"gte": 1}}}})
	r.Error(err)
	a.Nil(p)

	p, err = fp.Process(&Filter{Where: []map[string]interface{}{{"money": map[string]interface{}{"lt": 1}}}})
	r.Error(err)
	a.Nil(p)

	p, err = fp.Process(&Filter{Where: []map[string]interface{}{{"money": map[string]interface{}{"lte": 1}}}})
	r.Error(err)
	a.Nil(p)

	p, err = fp.Process(&Filter{Where: []map[string]interface{}{{"not": 1}}})
	r.Error(err)
	a.Nil(p)

	p, err = fp.Process(nil)
	r.NoError(err)
}
