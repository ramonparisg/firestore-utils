package controller

import (
	"firestore-utils/internal/repository"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
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

		response = append(response, data)
	}

	return response
}

func populateDataWithSelectedFields(response map[string]interface{}, selectedFields map[string][]string, row interface{}) interface{} {
	for responseKey := range selectedFields {
		fieldMaps := selectedFields[responseKey]
		for i, field := range fieldMaps {
			for rowKey := range row.(map[string]interface{}) {
				if strings.EqualFold(rowKey, field) {
					rowValue := row.(map[string]interface{})[rowKey]
					if isLastValue(fieldMaps) {
						response[responseKey] = rowValue
					} else if isNestedValue(i, fieldMaps) {
						fieldsLeftToFind := map[string][]string{responseKey: fieldMaps[i+1:]}
						populateDataWithSelectedFields(response, fieldsLeftToFind, rowValue)
					}
					break
				}
			}
		}
	}

	return nil

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
