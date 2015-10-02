package arangolite

import (
	"encoding/json"
	"fmt"
)

type Filter struct {
	Offset  int                    `json:"offset"`
	Limit   int                    `json:"limit"`
	Sort    []string               `json:"sort"`
	Where   map[string]interface{} `json:"where"`
	Pluck   string                 `json:"pluck"`
	Options []string               `json:"options"`
}

type ProcessedFilter struct {
	OffsetLimit string
	Sort        string
	Where       string
	Pluck       string
}

func GetFilter(jsonFilter string) (*Filter, error) {
	filter := &Filter{}

	if err := json.Unmarshal([]byte(jsonFilter), filter); err != nil {
		return nil, err
	}

	return filter, nil
}

func GetAQLFilter(f *Filter) (string, error) {
	fp := NewFilterProcessor("var")
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

	if len(filter.Pluck) != 0 {
		aqlFilter = fmt.Sprintf("%s COLLECT var2 = %s OPTIONS { method: 'sorted' } RETURN var2", aqlFilter, filter.Pluck)
	} else {
		aqlFilter = fmt.Sprintf("%s %s", aqlFilter, `RETURN var`)
	}

	return aqlFilter, nil
}
