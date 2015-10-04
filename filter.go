package arangolite

import (
	"encoding/json"
	"fmt"
)

// Filter defines a way of filtering AQL queries.
type Filter struct {
	Offset  int                    `json:"offset"`
	Limit   int                    `json:"limit"`
	Sort    []string               `json:"sort"`
	Where   map[string]interface{} `json:"where"`
	Options []string               `json:"options"`
}

type processedFilter struct {
	OffsetLimit string
	Sort        string
	Where       string
}

// GetFilter converts a JSON filter to a Filter object.
func GetFilter(jsonFilter string) (*Filter, error) {
	filter := &Filter{}

	if err := json.Unmarshal([]byte(jsonFilter), filter); err != nil {
		return nil, err
	}

	return filter, nil
}

// GetAQLFilter converts a Filter object to its translation in AQL.
func GetAQLFilter(f *Filter) (string, error) {
	fp := newFilterProcessor("var")
	filter, err := fp.Process(f)
	if err != nil {
		return "", err
	}

	aqlFilter := "FOR var IN result"

	if len(filter.OffsetLimit) != 0 {
		aqlFilter = fmt.Sprintf("%s LIMIT %s", aqlFilter, filter.OffsetLimit)
	}

	if len(filter.Sort) != 0 {
		aqlFilter = fmt.Sprintf("%s SORT %s", aqlFilter, filter.Sort)
	}

	if len(filter.Where) != 0 {
		aqlFilter = fmt.Sprintf("%s FILTER %s", aqlFilter, filter.Where)
	}

	aqlFilter = aqlFilter + ` RETURN var`

	return aqlFilter, nil
}
