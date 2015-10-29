package filters

import (
	"bytes"
	"encoding/json"
)

// Filter defines a way of filtering AQL queries.
type Filter struct {
	Offset  int                      `json:"offset"`
	Limit   int                      `json:"limit"`
	Sort    []string                 `json:"sort"`
	Where   []map[string]interface{} `json:"where"`
	Options []string                 `json:"options"`
}

type processedFilter struct {
	OffsetLimit string
	Sort        string
	Where       string
}

// FromJSON converts a JSON filter to a Filter object.
func FromJSON(jsonFilter string) (*Filter, error) {
	filter := &Filter{}

	if err := json.Unmarshal([]byte(jsonFilter), filter); err != nil {
		return nil, err
	}

	return filter, nil
}

// ToAQL converts a Filter object to its translation in AQL.
// "tmpVar" is the AQL var name to apply the filter on.
func ToAQL(tmpVar string, f *Filter) (string, error) {
	fp := newFilterProcessor(tmpVar)
	filter, err := fp.Process(f)
	if err != nil {
		return "", err
	}

	aqlFilter := bytes.NewBuffer(nil)

	if len(filter.Where) != 0 {
		aqlFilter.WriteString("FILTER ")
		aqlFilter.WriteString(filter.Where)
	}

	if len(filter.Sort) != 0 {
		aqlFilter.WriteString(" SORT ")
		aqlFilter.WriteString(filter.Sort)
	}

	if len(filter.OffsetLimit) != 0 {
		aqlFilter.WriteString(" LIMIT ")
		aqlFilter.WriteString(filter.OffsetLimit)
	}

	return aqlFilter.String(), nil
}
