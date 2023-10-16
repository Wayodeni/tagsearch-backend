package controllers

import "github.com/Wayodeni/tagsearch-backend/internal/storage/repository"

type TagController struct {
	repository *repository.TagRepository
}

func NewTagController(tagRepository *repository.TagRepository) *TagController {
	return &TagController{
		repository: tagRepository,
	}
}
