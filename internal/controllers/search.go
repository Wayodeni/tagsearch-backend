package controllers

import (
	"fmt"
	"log"
	"net/http"

	service "github.com/Wayodeni/tagsearch-backend/internal/service/index"
	"github.com/gin-gonic/gin"
)

type SearchController struct {
	service *service.IndexService
}

func NewSearchController(service *service.IndexService) *SearchController {
	return &SearchController{
		service: service,
	}
}

func (controller *SearchController) Search(c *gin.Context) {
	searchResults, err := controller.service.Find(&service.SearchDocumentRequest{
		Query: c.Query("query"),
		Tags:  c.QueryArray("tags[]"),
	})
	if err != nil {
		err = fmt.Errorf("error during search: %w", err)
		log.Println(err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, searchResults)
}
