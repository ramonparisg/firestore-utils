package controller

import (
	"firestore-utils/internal/repository"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
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
	r.POST("/query", func(g *gin.Context) {
		var request QueryRequest
		err := g.ShouldBindJSON(&request)
		if err != nil {
			g.JSON(http.StatusInternalServerError, err)
			log.Printf("Error %v", err)
			return
		}

		filters := mapToRepoFilters(request)

		result, err := c.repo.Query(request.Collection, filters, request.Limit)
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
