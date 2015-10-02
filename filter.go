package arangolite

import "encoding/json"

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

// func GetAQLFilter(filter *Filter) (string, error) {
// 	gormFilter, err := processFilter(filter)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	if len(gormFilter.Fields) != 0 {
// 		query = query.Select(gormFilter.Fields)
// 	}
//
// 	if gormFilter.Offset != 0 {
// 		query = query.Offset(gormFilter.Offset)
// 	}
//
// 	if gormFilter.Limit != 0 {
// 		query = query.Limit(gormFilter.Limit)
// 	}
//
// 	if gormFilter.Order != "" {
// 		query = query.Order(gormFilter.Order)
// 	}
//
// 	if gormFilter.Where != "" {
// 		query = query.Where(gormFilter.Where)
// 	}
//
// 	for _, include := range gormFilter.Include {
// 		if include.Relation == "" {
// 			break
// 		}
//
// 		if include.Where == "" {
// 			query = query.Preload(include.Relation)
// 		} else {
// 			query = query.Preload(include.Relation, include.Where)
// 		}
// 	}
//
// 	return query, nil
// }
