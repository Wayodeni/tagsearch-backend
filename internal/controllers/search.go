package controllers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

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
	const SYMBOLS_TO_REMOVE = "+-=&|><!(){}[]^\"~*?:\\/ "
	queryString := c.Query("query")

	if len(queryString) != 0 {
		// Removing querystring query syntax special symbols to prevent 500 error
		for _, symbolToRemove := range SYMBOLS_TO_REMOVE {
			if queryString[len(queryString)-1] == byte(symbolToRemove) {
				queryString = queryString[:len(queryString)-1]
				break
			}
		}
	}

	pageSizeString, ok := c.GetQuery("pageSize")
	pageSizeInt := 10
	var err error
	if ok {
		pageSizeInt, err = strconv.Atoi(pageSizeString)
		if err != nil {
			err = fmt.Errorf("error during search: %w", err)
			log.Println(err)
			c.AbortWithError(http.StatusBadRequest, err)
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
			c.AbortWithError(http.StatusBadRequest, err)
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

	if err != nil {
		if err.Error() == "syntax error" {
			err = fmt.Errorf("syntax error detected, check your query")
			log.Println(err)
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		err = fmt.Errorf("error during search: %w", err)
		log.Println(err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, searchResults)
}
