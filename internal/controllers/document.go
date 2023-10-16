package controllers

import (
	service "github.com/Wayodeni/tagsearch-backend/internal/service/index"
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
	return
}

func (controller *DocumentController) Read(c *gin.Context) {
	return
}

func (controller *DocumentController) ReadMany(c *gin.Context) {
	return
}

func (controller *DocumentController) Update(c *gin.Context) {
	return
}

func (controller *DocumentController) Delete(c *gin.Context) {
	return
}

func (controller *DocumentController) List(c *gin.Context) {
	return
}
