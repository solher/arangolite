package filters

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// Filter defines a way of filtering AQL queries.
type Filter struct {
	Offset  int                      `json:"offset"`
	Limit   int                      `json:"limit"`
	Sort    []string                 `json:"sort"`
	Where   []map[string]interface{} `json:"where"`
	Options map[string]interface{}   `json:"options"`
}

type processedFilter struct {
	OffsetLimit string
	Sort        string
	Where       string
}

// FromRequest returns a filter object from a http request.
func FromRequest(r *http.Request) (*Filter, error) {
	param := r.URL.Query().Get("filter")

	if len(param) == 0 {
		param = r.URL.Query().Get("Filter")
	}

	if len(param) == 0 {
		return nil, nil
	}

	filter, err := FromJSON(param)
	if err != nil {
		return nil, err
	}

	return filter, nil
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
