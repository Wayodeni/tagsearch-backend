package service

import (
	"github.com/Wayodeni/tagsearch-backend/internal/storage/models"
	"github.com/Wayodeni/tagsearch-backend/internal/storage/repository"
	"github.com/blevesearch/bleve/v2"
)

type SearchDocumentRequest struct {
	Query string
	Tags  []models.TagResponse
}

type IndexService struct {
	index              *bleve.Index
	documentRepository *repository.DocumentRepository
}

func NewIndexService(index *bleve.Index, documentRepository *repository.DocumentRepository) *IndexService {
	return &IndexService{
		index:              index,
		documentRepository: documentRepository,
	}
}

func (service *IndexService) Find(searchRequest *SearchDocumentRequest) (response []models.DocumentResponse, err error) {
	return response, nil
}

func (service *IndexService) Index(documents []models.DocumentResponse) error {
	return nil
}
