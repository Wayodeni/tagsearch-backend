package models

type ID = int64

type CreateTagRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateTagRequest struct {
	Name string `json:"name" binding:"required"`
}

type TagResponse struct {
	ID   ID     `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}
