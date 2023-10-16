package repository

import (
	"github.com/Wayodeni/tagsearch-backend/internal/storage/models"
	"github.com/jmoiron/sqlx"
)

type DocumentRepository struct {
	db            *sqlx.DB
	tagRepository *TagRepository
}

func NewDocumentRepository(db *sqlx.DB, tagRepository *TagRepository) *DocumentRepository {
	return &DocumentRepository{
		db:            db,
		tagRepository: tagRepository,
	}
}

func (repository *DocumentRepository) Create(request models.CreateDocumentRequest) (response models.DocumentResponse, err error) {
	return response, err
}

func (repository *DocumentRepository) Read(id int) (response models.DocumentResponse, err error) {
	return response, err
}

func (repository *DocumentRepository) ReadMany(IDs []int) (response []models.DocumentResponse, err error) {
	return response, err
}

func (repository *DocumentRepository) Update(id int, updateRequest models.UpdateDocumentRequest) (response models.DocumentResponse, err error) {
	return response, err
}

func (repository *DocumentRepository) Delete(id int) (err error) {
	return nil
}

func (repository *DocumentRepository) List() (response []models.DocumentResponse, err error) {
	return response, err
}
