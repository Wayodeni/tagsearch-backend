package controllers

import (
	service "github.com/Wayodeni/tagsearch-backend/internal/service/index"
	"github.com/Wayodeni/tagsearch-backend/internal/storage/repository"
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
