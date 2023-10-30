package controller

import (
	"firestore-utils/internal/repository"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

type Controller struct {
	repo repository.Repository
}

var (
	controllerSingleton *Controller
	once                sync.Once
)

func NewController(repo repository.Repository) *Controller {
	once.Do(func() {
		controllerSingleton = &Controller{
			repo: repo,
		}
	})
	return controllerSingleton
}

func (c *Controller) RunController(r *gin.Engine) {
	r.POST("/query/:collection", func(g *gin.Context) {
		var request QueryRequest
		err := g.ShouldBindJSON(&request)
		if err != nil || len(request.Filters) == 0 {
			g.JSON(http.StatusInternalServerError, err)
			log.Printf("Error %v", err)
			return
		}

		collection := g.Param("collection")
		filters := mapToRepoFilters(request)

		result, err := c.repo.Query(collection, filters, request.Limit)
		if err != nil {
			g.JSON(http.StatusInternalServerError, err.Error())
			return
		}

		data := result
		if hasSelectedFields(request) {
			data = filterSelectedFields(request.Select, result)
		}

		if len(data) == 0 || data == nil {
			g.JSON(http.StatusNotFound, "Document(s) not found")
		} else {
			g.JSON(http.StatusOK, data)
		}
	})

}

func hasSelectedFields(request QueryRequest) bool {
	return request.Select != nil && len(request.Select) > 0
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

func flatNestedFields(data map[string]interface{}) {
	for key := range data {
		field := data[key]
		if field != nil {
			if reflect.TypeOf(field).Kind() == reflect.Slice {
				var values []interface{}
				for _, v := range field.([]map[string]interface{}) {
					if v == nil {
						continue
					}

					if len(v) == 1 {
						for nestedKey := range v {
							values = append(values, v[nestedKey])
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

/**
 * This function will populate the response map with the selected fields from the request.
* It will recursively iterate over the selected fields and the object to find the values.
*/
func populateDataWithSelectedFields(response map[string]interface{}, selectedFields map[string][]string, object interface{}) {
	for alias := range selectedFields {
		fieldRoutes := selectedFields[alias]
		for i, field := range fieldRoutes {
			for objectKey := range object.(map[string]interface{}) {
				if valueAlreadyFound(response, alias) {
					break
				}

				response[alias] = getDefaultValue()

				if fieldFound(objectKey, field) {
					objectValue := getRowValue(field, object, objectKey)
					if valueIsArray(objectValue) {
						arrayValue := objectValue.([]interface{})
						if isLastStepInRoute(fieldRoutes) {
							m := make([]map[string]interface{}, len(arrayValue))
							for key, val := range arrayValue {
								m[key] = val.(map[string]interface{})
							}
							response[alias] = m
						} else {
							for j, val := range arrayValue {
								fieldsLeftToFind := map[string][]string{alias: fieldRoutes[i+1:]}
								if isFirstFieldInArray(j) {
									response[alias] = convertToMapInterface(arrayValue)
								}
								populateDataWithSelectedFields(response[alias].([]map[string]interface{})[j], fieldsLeftToFind, val)
							}
						}
					} else {
						if isLastStepInRoute(fieldRoutes) {
							response[alias] = objectValue
						} else if isNestedValue(i, fieldRoutes) {
							fieldsLeftToFind := map[string][]string{alias: fieldRoutes[i+1:]}
							populateDataWithSelectedFields(response, fieldsLeftToFind, objectValue)
						}
					}
				}
			}
		}
	}
}

func isFirstFieldInArray(j int) bool {
	return j == 0
}

func convertToMapInterface(arrayValue []interface{}) []map[string]interface{} {
	var m []map[string]interface{}
	for i := 0; i < len(arrayValue); i++ {
		m = append(m, make(map[string]interface{}))
	}
	return m
}

func getDefaultValue() interface{} {
	return nil
}

func valueAlreadyFound(response map[string]interface{}, responseKey string) bool {
	return response[responseKey] != nil
}

func getRowValue(field string, row interface{}, rowKey string) interface{} {
	var rowValue interface{} = make(map[string]interface{})
	if fieldIsArray(field, row, rowKey) {
		split := strings.Split(field, "[")
		field = split[0]
		index := strings.ReplaceAll(split[1], "]", "")
		arrayValue := row.(map[string]interface{})[rowKey].([]interface{})
		if isNumeric(index) {
			indexInt, _ := strconv.Atoi(index)
			if len(arrayValue) > indexInt {
				rowValue = arrayValue[indexInt]
			}
		} else if isN(index) {
			rowValue = arrayValue
		} else if hasCondition(index) {
			splitCondition := strings.Split(index, "=")
			fieldToFind := strings.Trim(splitCondition[0], " ")
			valueToMatch := strings.Trim(splitCondition[1], " ")
			for _, value := range arrayValue {
				for key, keyVal := range value.(map[string]interface{}) {
					if fieldFound(key, fieldToFind) {
						if strings.EqualFold(keyVal.(string), valueToMatch) {
							rowValue = value.(map[string]interface{})
							break
						}
					}
				}
			}
		}
	} else {
		rowValue = row.(map[string]interface{})[rowKey]
	}
	return rowValue
}

func hasCondition(index string) bool {
	return strings.Contains(index, "=")
}

func fieldIsArray(field string, row interface{}, rowKey string) bool {
	value := row.(map[string]interface{})[rowKey]
	return valueIsArray(value) && strings.Contains(field, "[") && strings.Contains(field, "]")
}

func valueIsArray(value interface{}) bool {
	return reflect.TypeOf(value).Kind() == reflect.Slice
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

func fieldFound(rowKey string, field string) bool {
	fieldFormatted := strings.Split(field, "[")[0]
	return strings.EqualFold(rowKey, fieldFormatted)
}

func isNestedValue(i int, fieldMaps []string) bool {
	return i+1 < len(fieldMaps)
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
