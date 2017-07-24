package requests

import (
	"encoding/json"
	"fmt"
)

// GetStatistics returns the statistics information.
type GetStatistics struct{}

func (r *GetStatistics) Path() string {
	return "/_admin/statistics"
}

func (r *GetStatistics) Method() string {
	return "GET"
}

func (r *GetStatistics) Generate() []byte {
	return []byte{}
}

type GetStatisticsResult struct {
	Time    float64
	Enabled bool
	// interface{} will be either float64 or DistributionStatistic.
	Statistics map[string]map[string]interface{}
}

type DistributionStatistic struct {
	Sum    float64 `json:"sum"`
	Count  int     `json:"count"`
	Counts []int   `json:"counts"`
}

func (r *GetStatisticsResult) UnmarshalJSON(bytes []byte) (err error) {
	obj := make(map[string]interface{})
	err = json.Unmarshal(bytes, &obj)
	if err != nil {
		return err
	}

	// catch any panics from the type assertions below
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	r.Statistics = make(map[string]map[string]interface{})
	for k, v := range obj {
		if k == "time" {
			r.Time = v.(float64)
			continue
		}

		if k == "enabled" {
			r.Enabled = v.(bool)
			continue
		}

		if group, ok := v.(map[string]interface{}); ok {
			r.Statistics[k] = make(map[string]interface{})
			for gk, gv := range group {
				// A statistic may be either a scalar value (if the statistic's
				// type is 'current' or 'accumulated'), or a DistributionStatistic
				// (if the type is 'distribution').
				switch gv := gv.(type) {
				case map[string]interface{}:
					var ds DistributionStatistic
					ds.Sum = gv["sum"].(float64)
					ds.Count = int(gv["count"].(float64))

					ds.Counts = []int{}
					counts := gv["counts"].([]interface{})
					for _, count := range counts {
						// json.Unmarshal stores all numbers as float64, but the
						// values returned from the API will always be integers,
						// so cast them.
						ds.Counts = append(ds.Counts, int(count.(float64)))
					}

					r.Statistics[k][gk] = ds
				default:
					r.Statistics[k][gk] = gv
				}
			}
		}
	}

	return nil
}

// GetStatisticsDescription fetches descriptive info of statistics.
type GetStatisticsDescription struct{}

func (r *GetStatisticsDescription) Path() string {
	return "/_admin/statistics-description"
}

func (r *GetStatisticsDescription) Method() string {
	return "GET"
}

func (r *GetStatisticsDescription) Generate() []byte {
	return []byte{}
}

type GetStatisticsDescriptionResult struct {
	Groups  []StatisticsGroup  `json:"groups"`
	Figures []StatisticsFigure `json:"figures"`
}

type StatisticsGroup struct {
	Group       string `json:"group"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type StatisticsFigureType string

const (
	StatisticsFigureTypeCurrent      StatisticsFigureType = "current"
	StatisticsFigureTypeAccumulated  StatisticsFigureType = "accumulated"
	StatisticsFigureTypeDistribution StatisticsFigureType = "distribution"
)

type StatisticsFigure struct {
	Group       string               `json:"group"`
	Identifier  string               `json:"identifier"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Type        StatisticsFigureType `json:"type"`
	Cuts        []float64            `json:"cuts,omitempty"`
	Units       string               `json:"units"`
}
