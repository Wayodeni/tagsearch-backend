package controllers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	service "github.com/Wayodeni/tagsearch-backend/internal/service/index"
	"github.com/Wayodeni/tagsearch-backend/internal/storage/models"
	"github.com/Wayodeni/tagsearch-backend/internal/storage/repository"
	"github.com/gin-gonic/gin"
)

type DocumentController struct {
	repository   *repository.DocumentRepository
	indexService *service.IndexService
}

func NewDocumentController(documentRepository *repository.DocumentRepository, indexService *service.IndexService) *DocumentController {
	return &DocumentController{
		repository:   documentRepository,
		indexService: indexService,
	}
}

func (controller *DocumentController) Create(c *gin.Context) {
	var createDocumentRequest models.CreateDocumentRequest

	if err := c.Bind(&createDocumentRequest); err != nil {
		err = fmt.Errorf("unable to bind request body during document create: %w", err)
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}

	createdDocument, err := controller.repository.Create(createDocumentRequest)
	if err != nil {
		err = fmt.Errorf("unable to create document in storage: %w", err)
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	if err := controller.indexService.Index([]models.DocumentResponse{createdDocument}); err != nil {
		err = fmt.Errorf("unable to index document after creation: %w", err)
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
	}

	c.JSON(http.StatusCreated, createdDocument)
}

func (controller *DocumentController) Read(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		err = fmt.Errorf("unable to convert id '%s' into int in document read", c.Param("id"))
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}

	documentResponse, err := controller.repository.Read(int64(id))
	if err != nil {
		err = fmt.Errorf("unable to read document with id '%v': %w", id, err)
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, documentResponse)
}

func (controller *DocumentController) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		err = fmt.Errorf("unable to convert id '%s' into int in document update", c.Param("id"))
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	var updateDocumentRequest models.UpdateDocumentRequest
	if err := c.Bind(&updateDocumentRequest); err != nil {
		err = fmt.Errorf("unable to bind request body during document update: %w", err)
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	updateDocumentRequest.RemoveCommonTags()

	documentResponse, err := controller.repository.Update(int64(id), updateDocumentRequest)
	if err != nil {
		err = fmt.Errorf("unable to update document: %w", err)
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	if err := controller.indexService.Index([]models.DocumentResponse{documentResponse}); err != nil {
		err = fmt.Errorf("unable to update document in index: %w", err)
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, documentResponse)
}

func (controller *DocumentController) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		err = fmt.Errorf("unable to convert id '%s' into int in document delete", c.Param("id"))
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	if err := controller.repository.Delete(int64(id)); err != nil {
		err = fmt.Errorf("unable to delete document: %w", err)
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	if err := controller.indexService.Delete([]models.ID{int64(id)}); err != nil {
		err = fmt.Errorf("unable to delete document from index: %w", err)
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

func (controller *DocumentController) List(c *gin.Context) {
	queryparamIDs, ok := c.GetQueryArray("ids")
	if ok {
		IDs := make([]int64, len(queryparamIDs))
		for index, queryparamID := range queryparamIDs { // Check if all of the passed IDs are integers
			id, err := strconv.Atoi(queryparamID)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, fmt.Errorf("not int id at position '%d': %w", index, err))
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
