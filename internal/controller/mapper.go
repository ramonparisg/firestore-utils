package controller

import (
	"firestore-utils/internal/repository"
	"reflect"
	"strconv"
	"strings"
)

func getValueFromData(data map[string]interface{}, fieldRoutes []string) (interface{}, bool) {
	for i, field := range fieldRoutes {
		value, exists := getValueByField(field, data)

		if !exists {
			return getDefaultValue(), false
		}

		if isLastStepInRoute(fieldRoutes) {
			return value, true
		}

		switch v := value.(type) {
		case map[string]interface{}:
			if result, ok := getValueFromData(v, fieldRoutes[i+1:]); ok {
				return result, true
			}
		case []interface{}:
			m := make([]interface{}, len(v))
			for j, item := range v {
				if subData, ok := item.(map[string]interface{}); ok { // todo cuando tengo un {n} y este retorna un array, no lo estoy manejando
					if result, ok := getValueFromData(subData, fieldRoutes[i+1:]); ok {
						m[j] = result
					}
				}
			}
			return m, true
		default:
			if isLastStepInRoute(fieldRoutes) { // todo is this really necessary?
				return value, true
			}
		}
	}

	return getDefaultValue(), false
}

func getValueByField(field string, data map[string]interface{}) (interface{}, bool) {
	fieldType := getRouteFieldType(field)
	fieldFormatted := strings.Split(field, "[")[0]
	switch fieldType {
	case "array":
		return getArrayValue(field, data, fieldFormatted)
	case "map":
		i, b, done := getMapValue(field, data)
		if done {
			return i, b
		}
	default:
		value, exists := data[fieldFormatted]
		return value, exists

	}

	value, exists := data[fieldFormatted]
	return value, exists
}

func getMapValue(field string, data map[string]interface{}) (interface{}, bool, bool) {
	split := strings.Split(field, "{")
	index := strings.ReplaceAll(split[1], "}", "")
	if isNumeric(index) {
		// todo esto está roto. Arreglarlo. Lo que está mal es que el map retorna las keys en cualquier orden. Por lo que buscar por orden funciona
		if indexInt, _ := strconv.Atoi(index); len(data) > indexInt {
			keys := make([]string, len(data))
			i := 0
			for k := range data {
				keys[i] = k
				i++
			}

			return data[keys[indexInt]], true, true
		}

	} else if isN(index) {
		m := make([]interface{}, len(data))
		i := 0
		for key := range data {
			m[i] = data[key]
			i++
		}

		return m, true, true
	}
	return nil, false, false
}

func getArrayValue(field string, data map[string]interface{}, fieldFormatted string) (interface{}, bool) {
	split := strings.Split(field, "[")
	index := strings.ReplaceAll(split[1], "]", "")
	val, exists := data[fieldFormatted]
	if !exists {
		return getDefaultValue(), false
	}

	arrayValue := val.([]interface{})
	if isNumeric(index) {
		indexInt, _ := strconv.Atoi(index)
		if len(arrayValue) > indexInt {
			return arrayValue[indexInt], true
		}
	} else if isN(index) {
		return arrayValue, true
	} else if hasCondition(index) {
		splitCondition := strings.Split(index, "=")
		fieldToFind := strings.Trim(splitCondition[0], " ")
		valueToMatch := strings.Trim(splitCondition[1], " ")
		for _, value := range arrayValue {
			formattedValue := value.(map[string]interface{})
			v, exists := formattedValue[fieldToFind]

			if strings.EqualFold(v.(string), valueToMatch) {
				return formattedValue, exists
			}
		}
	}
	return val, true
}

func getRouteFieldType(field string) string {
	if strings.Contains(field, "{") {
		return "map"
	}

	if strings.Contains(field, "[") {
		return "array"
	}

	return "value"
}

func filterSelectedFields(selected map[string]string, rows []interface{}) []interface{} {

	selectedMapped := make(map[string][]string)
	for s := range selected {
		selectedMapped[s] = strings.Split(selected[s], ".")
	}

	var response []interface{}
	for _, row := range rows {
		data := make(map[string]interface{})
		populateDataWithSelectedFields(data, selectedMapped, row)

		flatNestedFields(data)

		response = append(response, data)
	}

	return response
}

/**
 * This function will populate the response map with the selected fields from the request.
* It will recursively iterate over the selected fields and the object to find the values.
*/
func populateDataWithSelectedFields(response map[string]interface{}, selectedFields map[string][]string, object interface{}) {
	for alias := range selectedFields { // todo run this in parallel
		fieldRoutes := selectedFields[alias]
		data, _ := getValueFromData(object.(map[string]interface{}), fieldRoutes)
		response[alias] = data
	}
}

func getDefaultValue() interface{} {
	return nil
}

func hasCondition(index string) bool {
	return strings.Contains(index, "=")
}

func isN(index string) bool {
	return strings.EqualFold(index, "n")
}

func isNumeric(index string) bool {
	if _, err := strconv.Atoi(index); err == nil {
		return true
	}
	return false
}

func isLastStepInRoute(fieldMaps []string) bool {
	return len(fieldMaps) == 1
}

func mapToRepoFilters(request QueryRequest) []repository.Filter {
	var filters []repository.Filter
	for _, filter := range request.Filters {
		filters = append(filters, repository.Filter{
			Field:     filter.Field,
			Operation: filter.Operation,
			Value:     filter.Value,
		})
	}
	return filters
}

func flatNestedFields(data map[string]interface{}) {
	for key := range data {
		field := data[key]
		if field != nil {
			if reflect.TypeOf(field).Kind() == reflect.Slice {
				var values []interface{}
				for _, v := range field.([]interface{}) {
					if v == nil {
						continue
					}

					if reflect.TypeOf(v).Kind() == reflect.Slice {
						nestedValues := v.([]interface{})
						for i := range nestedValues {
							values = append(values, nestedValues[i])
						}
					} else {
						values = append(values, v)
					}
				}
				data[key] = values
			}
		}
	}
}
