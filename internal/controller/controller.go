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

		result, err := c.repo.Query(collection, filters)
		if err != nil {
			g.JSON(http.StatusInternalServerError, err)
			return
		}

		if len(result) == 0 || result == nil {
			g.JSON(http.StatusNotFound, "Document(s) not found")
		} else {
			g.JSON(http.StatusOK, result)
		}
	})

	r.POST("v2/query/:collection", func(g *gin.Context) {
		var request QueryRequest
		err := g.ShouldBindJSON(&request)
		if err != nil || len(request.Filters) == 0 {
			g.JSON(http.StatusInternalServerError, err)
			log.Printf("Error %v", err)
			return
		}

		collection := g.Param("collection")
		filters := mapToRepoFilters(request)

		result, err := c.repo.Query(collection, filters)
		data := filterSelectedFields(request.Select, result)
		if err != nil {
			g.JSON(http.StatusInternalServerError, err)
			return
		}

		if len(data) == 0 || data == nil {
			g.JSON(http.StatusNotFound, "Document(s) not found")
		} else {
			g.JSON(http.StatusOK, data)
		}
	})

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

func populateDataWithSelectedFields(response map[string]interface{}, selectedFields map[string][]string, row interface{}) interface{} {
	for responseKey := range selectedFields {
		fields := selectedFields[responseKey]
		for i, field := range fields {
			for rowKey := range row.(map[string]interface{}) {
				fieldFormatted := strings.Split(field, "[")[0]
				if fieldFound(rowKey, fieldFormatted) {
					rowValue := getRowValue(field, row, rowKey)
					if valueIsArray(rowValue) {
						arrayValue := rowValue.([]interface{})
						if isLastValue(fields) {
							m := make([]map[string]interface{}, len(arrayValue))
							for key, val := range arrayValue {
								m[key] = val.(map[string]interface{})
							}
							response[responseKey] = m
							break
						} else {
							for j, val := range arrayValue {
								fieldsLeftToFind := map[string][]string{responseKey: fields[i+1:]}
								if j == 0 {
									var m []map[string]interface{}
									for i := 0; i < len(arrayValue); i++ {
										m = append(m, make(map[string]interface{}))
									}
									response[responseKey] = m
								}
								populateDataWithSelectedFields(response[responseKey].([]map[string]interface{})[j], fieldsLeftToFind, val)
							}
						}
					} else {
						if isLastValue(fields) {
							response[responseKey] = rowValue
						} else if isNestedValue(i, fields) {
							fieldsLeftToFind := map[string][]string{responseKey: fields[i+1:]}
							populateDataWithSelectedFields(response, fieldsLeftToFind, rowValue)
						}
					}
					break
				}
			}
		}
	}

	return nil

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
	return strings.EqualFold(rowKey, field)
}

func isNestedValue(i int, fieldMaps []string) bool {
	return i+1 < len(fieldMaps)
}

func isLastValue(fieldMaps []string) bool {
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
