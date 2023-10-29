package controllers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/Wayodeni/tagsearch-backend/internal/storage/models"
	"github.com/Wayodeni/tagsearch-backend/internal/storage/repository"
	"github.com/gin-gonic/gin"
)

type TagController struct {
	repository *repository.TagRepository
}

func NewTagController(tagRepository *repository.TagRepository) *TagController {
	return &TagController{
		repository: tagRepository,
	}
}

func (controller *TagController) Create(c *gin.Context) {
	var createTagRequest models.CreateTagRequest

	if err := c.Bind(&createTagRequest); err != nil {
		log.Println(err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	createdTag, err := controller.repository.Create(createTagRequest)
	if err != nil {
		log.Println(err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, createdTag)
}

func (controller *TagController) Read(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Println(err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	tagResponse, err := controller.repository.Read(id)
	if err != nil {
		log.Println(err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, tagResponse)
}

func (controller *TagController) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Println(err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	var updateTagRequest models.UpdateTagRequest
	if err := c.Bind(&updateTagRequest); err != nil {
		log.Println(err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	tagResponse, err := controller.repository.Update(id, updateTagRequest)
	if err != nil {
		log.Println(err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, tagResponse)
}

func (controller *TagController) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Println(err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if err := controller.repository.Delete(id); err != nil {
		log.Println(err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (controller *TagController) List(c *gin.Context) {
	queryparamIDs, ok := c.GetQueryArray("ids")
	if ok {
		IDs := make([]int, len(queryparamIDs))
		for index, queryparamID := range queryparamIDs { // Check if all of the passed IDs are integers
			id, err := strconv.Atoi(queryparamID)
			if err != nil {
				c.AbortWithError(http.StatusBadRequest, fmt.Errorf("not int id at position '%d': %w", index, err))
				return
			}
			IDs = append(IDs, id)
		}

		response, err := controller.repository.ReadMany(IDs)
		if err != nil {
			log.Println(err)
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.JSON(http.StatusOK, response)
		return
	}

	response, err := controller.repository.List()
	if err != nil {
		log.Println(err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, response)
}
