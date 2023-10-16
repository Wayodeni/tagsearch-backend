package repository

import (
	"github.com/Wayodeni/tagsearch-backend/internal/storage/models"
	"github.com/jmoiron/sqlx"
)

type TagRepository struct {
	db *sqlx.DB
}

func NewTagRepository(db *sqlx.DB) *TagRepository {
	return &TagRepository{
		db: db,
	}
}

func (repository *TagRepository) Create(request models.CreateTagRequest) (response models.TagResponse, err error) {
	return response, err
}

func (repository *TagRepository) Read(id int) (response models.TagResponse, err error) {
	return response, err
}

func (repository *TagRepository) ReadMany(IDs []int) (response []models.TagResponse, err error) {
	return response, err
}

func (repository *TagRepository) Update(id int, updateRequest models.UpdateTagRequest) (response models.TagResponse, err error) {
	return response, err
}

func (repository *TagRepository) Delete(id int) (err error) {
	return nil
}

func (repository *TagRepository) List() (response []models.TagResponse, err error) {
	return response, err
}
