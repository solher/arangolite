package filters

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var aqlOp = []string{
	"FOR", "RETURN", "FILTER", "SORT", "LIMIT", "LET", "COLLECT", "INTO",
	"KEEP", "WITH", "COUNT", "OPTIONS", "REMOVE", "UPDATE", "REPLACE", "INSERT",
	"UPSERT",
}

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

type filterProcessor struct {
	VarName string
}

func newFilterProcessor(varName string) *filterProcessor {
	if len(varName) == 0 {
		varName = "var"
	}

	return &filterProcessor{VarName: varName}
}

func (fp *filterProcessor) Process(f *Filter) (*processedFilter, error) {
	if f == nil {
		return &processedFilter{}, nil
	}

	if err := fp.checkFilter(f); err != nil {
		return nil, err
	}

	pf := &processedFilter{}

	if f.Offset > 0 {
		pf.OffsetLimit = strconv.Itoa(f.Offset)
	}

	if f.Limit > 0 {
		if len(pf.OffsetLimit) > 0 {
			pf.OffsetLimit = pf.OffsetLimit + ", " + strconv.Itoa(f.Limit)
		} else {
			pf.OffsetLimit = strconv.Itoa(f.Limit)
		}
	}

	if f.Sort != nil && len(f.Sort) != 0 {
		var processedSort string

		for _, s := range f.Sort {
			matched, err := regexp.MatchString("\\A[0-9a-zA-Z_][0-9a-zA-Z._-]*(\\s(?i)(asc|desc))?\\z", s)
			if err != nil || !matched {
				return nil, errors.New("invalid sort filter: " + s)
			}

			split := strings.Split(s, " ")
			if len(split) == 1 {
				split = append(split, "ASC")
			} else {
				split[1] = strings.ToUpper(split[1])
			}

			processedSort = fmt.Sprintf("%s%s.%s %s, ", processedSort, fp.VarName, split[0], split[1])
		}

		pf.Sort = processedSort[:len(processedSort)-2]
	}

	if f.Where != nil && len(f.Where) != 0 {
		buffer := &bytes.Buffer{}
		if err := fp.processCondition(buffer, "", andAQL, "", f.Where); err != nil {
			return nil, err
		}

		pf.Where = buffer.String()
		pf.Where = pf.Where[1 : len(pf.Where)-1]
	}

	return pf, nil
}

func (fp *filterProcessor) processCondition(buffer *bytes.Buffer, attribute, operator, sign string, condition interface{}) error {
	switch condition.(type) {
	case map[string]interface{}:
		if err := fp.processUnaryCondition(buffer, attribute, operator, condition.(map[string]interface{})); err != nil {
			return err
		}

	case interface{}:
		if buffer.Len() != 0 {
			buffer.WriteString(operator)
		}
		if err := fp.processOperation(buffer, attribute, operator, sign, condition); err != nil {
			return err
		}
	}

	return nil
}

func (fp *filterProcessor) processUnaryCondition(buffer *bytes.Buffer, attribute, operator string, condition map[string]interface{}) error {
	for key := range condition {
		lowerKey := strings.ToLower(key)

		switch lowerKey {
		case "gt":
			if buffer.Len() != 0 {
				buffer.WriteString(operator)
			}
			if err := fp.processOperation(buffer, attribute, "", gtAQL, condition[key]); err != nil {
				return err
			}
			break

		case "gte":
			if buffer.Len() != 0 {
				buffer.WriteString(operator)
			}
			if err := fp.processOperation(buffer, attribute, "", gteAQL, condition[key]); err != nil {
				return err
			}
			break

		case "lt":
			if buffer.Len() != 0 {
				buffer.WriteString(operator)
			}
			if err := fp.processOperation(buffer, attribute, "", ltAQL, condition[key]); err != nil {
				return err
			}
			break

		case "lte":
			if buffer.Len() != 0 {
				buffer.WriteString(operator)
			}
			if err := fp.processOperation(buffer, attribute, "", lteAQL, condition[key]); err != nil {
				return err
			}
			break

		case "eq":
			if buffer.Len() != 0 {
				buffer.WriteString(operator)
			}
			if err := fp.processOperation(buffer, attribute, "", eqAQL, condition[key]); err != nil {
				return err
			}
			break

		case "neq":
			if buffer.Len() != 0 {
				buffer.WriteString(operator)
			}
			if err := fp.processOperation(buffer, attribute, "", neqAQL, condition[key]); err != nil {
				return err
			}
			break

		case "not":
			if buffer.Len() != 0 {
				buffer.WriteString(operator)
			}
			newBuffer := &bytes.Buffer{}

			buffer.WriteString(notAQL + "(")
			if err := fp.processCondition(newBuffer, "", andAQL, eqAQL, condition[key]); err != nil {
				return err
			}

			buffer.Write(newBuffer.Bytes())
			buffer.WriteString(")")

		case "or":
			mapArr, err := fp.checkAndOrCondition(condition[key])
			if err != nil {
				return err
			}

			if buffer.Len() != 0 {
				buffer.WriteString(operator)
			}

			if err := fp.processOperation(buffer, "", orAQL, eqAQL, mapArr); err != nil {
				return err
			}

		case "and":
			mapArr, err := fp.checkAndOrCondition(condition[key])
			if err != nil {
				return err
			}

			if buffer.Len() != 0 {
				buffer.WriteString(operator)
			}

			if err := fp.processOperation(buffer, "", andAQL, eqAQL, mapArr); err != nil {
				return err
			}

		case "like":
			paramMap, err := fp.checkFunctionCondition(condition[key])
			if err != nil {
				return err
			}

			buffer.WriteString("LIKE(")
			buffer.WriteString(fp.VarName)
			buffer.WriteRune('.')
			buffer.WriteString(paramMap["text"].(string))
			buffer.WriteString(", ")
			fp.writeQuotedString(buffer, paramMap["search"].(string))
			buffer.WriteString(", ")
			caseInsensitive, ok := paramMap["case_insensitive"]
			if ok && caseInsensitive.(bool) {
				buffer.WriteString("true")
			}
			buffer.WriteRune(')')

		default:
			if err := fp.processCondition(buffer, key, operator, eqAQL, condition[key]); err != nil {
				return err
			}
		}
	}

	return nil
}

func (fp *filterProcessor) processOperation(buffer *bytes.Buffer, attribute, operator, sign string, condition interface{}) error {
	switch condition := condition.(type) {
	case bool:
		if condition {
			fp.processSimpleOperation(buffer, attribute, sign, trueBoolAQL)
		} else {
			fp.processSimpleOperation(buffer, attribute, sign, falseBoolAQL)
		}

	case string:
		if condition != "null" {
			fp.processSimpleOperationStr(buffer, attribute, sign, condition)
		} else {
			fp.processSimpleOperation(buffer, attribute, sign, condition)
		}

	case float64:
		fp.processSimpleOperation(buffer, attribute, sign, strconv.FormatFloat(condition, 'f', -1, 64))

	case []map[string]interface{}:
		newBuffer := &bytes.Buffer{}

		buffer.WriteString("(")

		for _, c := range condition {
			if err := fp.processCondition(newBuffer, "", operator, sign, c); err != nil {
				return err
			}
		}
		buffer.Write(newBuffer.Bytes())

		buffer.WriteString(")")

	// When a JSON is unmarshalled in the Where field of the Filter, all the arrays
	// are given as []interface{}. We have to check the elem types manually.
	case []interface{}:
		buffer.WriteString(fp.VarName)
		buffer.WriteRune('.')
		buffer.WriteString(attribute)
		buffer.WriteString(inArrayAQL + openArrayAQL)

		for i, c := range condition {
			switch c := c.(type) {
			case bool:
				if c {
					buffer.WriteString(trueBoolAQL)
				} else {
					buffer.WriteString(falseBoolAQL)
				}

			case string:
				fp.writeQuotedString(buffer, c)

			case float64:
				buffer.WriteString(strconv.FormatFloat(c, 'f', -1, 64))

			default:
				return fmt.Errorf("unrecognized type in: %v", reflect.TypeOf(condition))
			}

			if i < len(condition)-1 {
				buffer.WriteString(", ")
			}
		}

		buffer.WriteString(closeArrayAQL)

	default:
		return fmt.Errorf("unrecognized type: %v", reflect.TypeOf(condition))
	}

	return nil
}

func (fp *filterProcessor) processSimpleOperation(buffer *bytes.Buffer, attribute, sign, condition string) {
	buffer.WriteString(fp.VarName)
	buffer.WriteRune('.')
	buffer.WriteString(attribute)
	buffer.WriteString(sign)
	buffer.WriteString(condition)
}

func (fp *filterProcessor) processSimpleOperationStr(buffer *bytes.Buffer, attribute, sign, condition string) {
	buffer.WriteString(fp.VarName)
	buffer.WriteRune('.')
	buffer.WriteString(attribute)
	buffer.WriteString(sign)
	fp.writeQuotedString(buffer, condition)
}

func (fp *filterProcessor) writeQuotedString(buffer *bytes.Buffer, str string) {
	buffer.WriteRune('\'')
	buffer.WriteString(str)
	buffer.WriteRune('\'')
}

func (fp *filterProcessor) checkAndOrCondition(condition interface{}) ([]map[string]interface{}, error) {
	condType := reflect.TypeOf(condition)

	if condType == reflect.TypeOf([]map[string]interface{}{}) {
		return condition.([]map[string]interface{}), nil
	}

	if condType != reflect.TypeOf([]interface{}{}) {
		return nil, fmt.Errorf("invalid condition, must be an array: %v", condition)
	}

	arrCondition := condition.([]interface{})
	mapArr := []map[string]interface{}{}
	mapType := reflect.TypeOf(map[string]interface{}{})

	for _, c := range arrCondition {
		if reflect.TypeOf(c) != mapType {
			return nil, fmt.Errorf("invalid condition, values are present: %v", condition)
		}

		mapArr = append(mapArr, c.(map[string]interface{}))
	}

	return mapArr, nil
}

func (fp *filterProcessor) checkFunctionCondition(condition interface{}) (map[string]interface{}, error) {
	condType := reflect.TypeOf(condition)

	if condType == reflect.TypeOf(map[string]interface{}{}) {
		return condition.(map[string]interface{}), nil
	}

	return nil, fmt.Errorf("invalid condition, must be a map: %v", condition)
}

func (fp *filterProcessor) checkFilter(filter *Filter) error {
	c := make(chan error)

	for _, v := range filter.Sort {
		go func(v string) { fp.checkAQLOperators(v, c) }(v)
	}

	counter := len(filter.Sort) + fp.checkWhereFilter(filter.Where, c)

	for i := 0; i < counter; i++ {
		if err := <-c; err != nil {
			return err
		}
	}

	return nil
}

func (fp *filterProcessor) checkWhereFilter(filter interface{}, c chan error) int {
	var counter int

	switch f := filter.(type) {
	case map[string]interface{}:
		for k := range f {
			counter++
			go func() { fp.checkAQLOperators(k, c) }()
			fp.checkWhereFilter(f[k], c)
		}
	case []interface{}:
		for _, v := range f {
			fp.checkWhereFilter(v, c)
		}
	}

	return counter
}

func (fp *filterProcessor) checkAQLOperators(op string, c chan error) {
	upperOp := strings.ToUpper(op)

	for _, op := range aqlOp {
		if strings.Contains(upperOp, op) {
			c <- errors.New("forbidden AQL operator detected: " + op)
			return
		}
	}

	c <- nil
}
