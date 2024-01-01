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

type DocumentLister interface {
	ListForTag(tagID models.ID) (response []models.DocumentResponse, err error)
	ReadMany(IDs []models.ID) (response []models.DocumentResponse, err error)
}

type Indexer interface {
	Index(documents []models.DocumentResponse) error
}

type TagController struct {
	repository         *repository.TagRepository
	documentRepository DocumentLister
	indexService       Indexer
}

func NewTagController(tagRepository *repository.TagRepository, documentRepository DocumentLister, indexService Indexer) *TagController {
	return &TagController{
		repository:         tagRepository,
		documentRepository: documentRepository,
		indexService:       indexService,
	}
}

func (controller *TagController) Create(c *gin.Context) {
	var createTagRequest models.CreateTagRequest

	if err := c.Bind(&createTagRequest); err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}

	createdTag, err := controller.repository.Create(createTagRequest)
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, createdTag)
}

func (controller *TagController) Read(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	tagResponse, err := controller.repository.Read(int64(id))
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, tagResponse)
}

func (controller *TagController) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	var updateTagRequest models.UpdateTagRequest
	if err := c.Bind(&updateTagRequest); err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	// TODO: tag update, document listing and reindexing in one transaction
	tagResponse, err := controller.repository.Update(int64(id), updateTagRequest)
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	tagDocuments, err := controller.documentRepository.ListForTag(int64(id))
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	if err := controller.indexService.Index(tagDocuments); err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, tagResponse)
}

func (controller *TagController) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	// TODO: listing, deleting, reindexing in one transaction
	tagDocuments, err := controller.documentRepository.ListForTag(int64(id))
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	documentsIDs := make([]models.ID, 0, len(tagDocuments))
	for _, document := range tagDocuments {
		documentsIDs = append(documentsIDs, document.ID)
	}

	if err := controller.repository.Delete(int64(id)); err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	documentsWithoutDeletedTag, err := controller.documentRepository.ReadMany(documentsIDs)
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	if err := controller.indexService.Index(documentsWithoutDeletedTag); err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

func (controller *TagController) List(c *gin.Context) {
	queryparamIDs, ok := c.GetQueryArray("ids")
	if ok {
		IDs := make([]int64, len(queryparamIDs))
		for index, queryparamID := range queryparamIDs { // Check if all of the passed IDs are integers
			id, err := strconv.Atoi(queryparamID)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, fmt.Errorf("not int id at position '%d': %w", index, err).Error())
				return
			}
			IDs = append(IDs, int64(id))
		}

		response, err := controller.repository.ReadMany(IDs)
		if err != nil {
			log.Println(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
			return
		}

		c.JSON(http.StatusOK, response)
		return
	}

	response, err := controller.repository.List()
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, response)
}
