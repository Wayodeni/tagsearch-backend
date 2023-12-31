package controllers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

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
	queryString := c.Query("query")

	pageSizeString, ok := c.GetQuery("pageSize")
	pageSizeInt := 10
	var err error
	if ok {
		pageSizeInt, err = strconv.Atoi(pageSizeString)
		if err != nil {
			err = fmt.Errorf("error during search: %w", err)
			log.Println(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
	}

	pageNumberString, ok := c.GetQuery("pageNumber")
	pageNumberInt := 1
	if ok {
		pageNumberInt, err = strconv.Atoi(pageNumberString)
		if err != nil {
			err = fmt.Errorf("error during search: %w", err)
			log.Println(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
	}
	if pageNumberInt > 0 {
		pageNumberInt -= 1 // substituting because frontend does not have 0 in paginator
	}
	searchResults, err := controller.service.Find(&service.SearchDocumentRequest{
		Query:      queryString,
		Tags:       c.QueryArray("tags[]"),
		PageSize:   pageSizeInt,
		PageNumber: pageNumberInt,
	})

	if err != nil && strings.Contains(err.Error(), "parse error") {
		err = fmt.Errorf("error during querystring parsing: %w", err)
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	} else if err != nil {
		err = fmt.Errorf("error during search: %w", err)
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, searchResults)
}
