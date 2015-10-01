package arangolite

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
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

const (
	inArrayAQL    = " IN "
	openArrayAQL  = "["
	closeArrayAQL = "]"

	trueBoolAQL  = "true"
	falseBoolAQL = "false"

	notAQL = "!"
	orAQL  = " || "
	andAQL = " && "

	gtAQL  = " > "
	gteAQL = " >= "
	ltAQL  = " < "
	lteAQL = " <= "
	eqAQL  = " == "
	neqAQL = " != "
)

func processFilter(f *Filter) (*ProcessedFilter, error) {
	pf := &ProcessedFilter{}

	if f.Offset != 0 {
		if f.Offset < 0 {
			return nil, fmt.Errorf("invalid offset filter: %d", f.Offset)
		}

		pf.OffsetLimit = strconv.Itoa(f.Offset)
	}

	if f.Limit != 0 {
		if f.Limit < 0 {
			return nil, fmt.Errorf("invalid limit filter: %d", f.Limit)
		}

		if len(pf.OffsetLimit) > 0 {
			pf.OffsetLimit = pf.OffsetLimit + ", " + strconv.Itoa(f.Limit)
		} else {
			pf.OffsetLimit = strconv.Itoa(f.Limit)
		}
	}

	if f.Sort != nil && len(f.Sort) != 0 {
		var processedSort string

		for _, s := range f.Sort {
			matched, err := regexp.MatchString("\\A[0-9a-zA-Z._-]+(\\s(?i)(asc|desc))?\\z", s)
			if err != nil || !matched {
				return nil, errors.New("invalid sort filter: " + s)
			}

			split := strings.Split(s, " ")
			if len(split) == 1 {
				split = append(split, "ASC")
			} else {
				split[1] = strings.ToUpper(split[1])
			}

			processedSort = fmt.Sprintf("%s%s %s, ", processedSort, split[0], split[1])
		}

		pf.Sort = processedSort[:len(processedSort)-2]
	}

	if len(f.Pluck) != 0 {
		matched, err := regexp.MatchString("\\A[0-9a-zA-Z._-]+\\z", f.Pluck)
		if err != nil || !matched {
			return nil, errors.New("invalid pluck filter: " + f.Pluck)
		}

		pf.Pluck = f.Pluck
	}

	if f.Where != nil && len(f.Where) != 0 {
		buffer := &bytes.Buffer{}
		if err := processCondition(buffer, "", andAQL, "", f.Where); err != nil {
			return nil, err
		}

		pf.Where = buffer.String()
	}

	return pf, nil
}

func processCondition(buffer *bytes.Buffer, attribute, operator, sign string, condition interface{}) error {
	switch condition.(type) {
	case map[string]interface{}:
		processUnaryCondition(buffer, attribute, operator, condition.(map[string]interface{}))

	case interface{}:
		if buffer.Len() != 0 {
			buffer.WriteString(operator)
		}
		processOperation(buffer, attribute, operator, sign, condition)
	}

	return nil
}

func processUnaryCondition(buffer *bytes.Buffer, attribute, operator string, condition map[string]interface{}) error {
	for key := range condition {
		lowerKey := strings.ToLower(key)

		switch lowerKey {
		case "gt":
			if buffer.Len() != 0 {
				buffer.WriteString(operator)
			}
			processOperation(buffer, attribute, "", gtAQL, condition[key])
			break

		case "gte":
			if buffer.Len() != 0 {
				buffer.WriteString(operator)
			}
			processOperation(buffer, attribute, "", gteAQL, condition[key])
			break

		case "lt":
			if buffer.Len() != 0 {
				buffer.WriteString(operator)
			}
			processOperation(buffer, attribute, "", ltAQL, condition[key])
			break

		case "lte":
			if buffer.Len() != 0 {
				buffer.WriteString(operator)
			}
			processOperation(buffer, attribute, "", lteAQL, condition[key])
			break

		case "eq":
			if buffer.Len() != 0 {
				buffer.WriteString(operator)
			}
			processOperation(buffer, attribute, "", eqAQL, condition[key])
			break

		case "neq":
			if buffer.Len() != 0 {
				buffer.WriteString(operator)
			}
			processOperation(buffer, attribute, "", neqAQL, condition[key])
			break

		case "not":
			if buffer.Len() != 0 {
				buffer.WriteString(operator)
			}
			newBuffer := &bytes.Buffer{}

			buffer.WriteString(notAQL + "(")
			processCondition(newBuffer, "", andAQL, eqAQL, condition[key])

			buffer.Write(newBuffer.Bytes())
			buffer.WriteString(")")

		case "or":
			if buffer.Len() != 0 {
				buffer.WriteString(operator)
			}
			processOperation(buffer, "", orAQL, eqAQL, condition[key].([]interface{}))

		case "and":
			if buffer.Len() != 0 {
				buffer.WriteString(operator)
			}
			processOperation(buffer, "", andAQL, eqAQL, condition[key].([]interface{}))

		default:
			processCondition(buffer, key, operator, eqAQL, condition[key])
		}
	}

	return nil
}

func processOperation(buffer *bytes.Buffer, attribute, operator, sign string, condition interface{}) error {
	switch condition.(type) {
	case bool:
		if condition.(bool) {
			processSimpleOperation(buffer, attribute, sign, trueBoolAQL)
		} else {
			processSimpleOperation(buffer, attribute, sign, falseBoolAQL)
		}

	case string:
		processSimpleOperationStr(buffer, attribute, sign, condition.(string))

	case int:
		processSimpleOperation(buffer, attribute, sign, strconv.FormatInt(int64(condition.(int)), 10))

	case float64:
		processSimpleOperation(buffer, attribute, sign, strconv.FormatFloat(condition.(float64), 'f', -1, 64))

	case []int:
		intArray := condition.([]int)
		lenArray := len(intArray)

		buffer.WriteString(attribute)
		buffer.WriteString(inArrayAQL + openArrayAQL)

		for i, value := range intArray {
			buffer.WriteString(strconv.FormatInt(int64(value), 10))
			if i < lenArray-1 {
				buffer.WriteString(", ")
			}
		}

		buffer.WriteString(closeArrayAQL)

	case []float64:
		floatArray := condition.([]float64)
		lenArray := len(floatArray)

		buffer.WriteString(attribute)
		buffer.WriteString(inArrayAQL + openArrayAQL)

		for i, value := range floatArray {
			buffer.WriteString(strconv.FormatFloat(value, 'f', -1, 64))
			if i < lenArray-1 {
				buffer.WriteString(", ")
			}
		}

		buffer.WriteString(closeArrayAQL)

	case []interface{}:
		conditions := condition.([]interface{})

		arrStr := []string{}
		strType := reflect.TypeOf("")

		for _, condition := range conditions {
			if reflect.TypeOf(condition) == strType {
				arrStr = append(arrStr, condition.(string))
			}
		}

		if len(arrStr) == 0 {
			newBuffer := &bytes.Buffer{}

			buffer.WriteString("(")

			for _, condition := range conditions {
				processCondition(newBuffer, "", operator, sign, condition)
			}
			buffer.Write(newBuffer.Bytes())

			buffer.WriteString(")")
		} else {
			lenArray := len(arrStr)

			buffer.WriteString(attribute)
			buffer.WriteString(inArrayAQL + openArrayAQL)

			for i, value := range arrStr {
				buffer.WriteRune('\'')
				buffer.WriteString(value)
				buffer.WriteRune('\'')

				if i < lenArray-1 {
					buffer.WriteString(", ")
				}
			}

			buffer.WriteString(closeArrayAQL)
		}
	}

	return nil
}

func processSimpleOperation(buffer *bytes.Buffer, attribute, sign, condition string) {
	buffer.WriteString(attribute)
	buffer.WriteString(sign)
	buffer.WriteString(condition)
}

func processSimpleOperationStr(buffer *bytes.Buffer, attribute, sign, condition string) {
	buffer.WriteString(attribute)
	buffer.WriteString(sign)
	buffer.WriteRune('\'')
	buffer.WriteString(condition)
	buffer.WriteRune('\'')
}
