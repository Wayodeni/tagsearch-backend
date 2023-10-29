package models

type CreateTagRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateTagRequest struct {
	Name string `json:"name" binding:"required"`
}

type tagResponseID = int
type TagResponse struct {
	ID   tagResponseID `json:"id" db:"id" binding:"required"`
	Name string        `json:"name" db:"name" binding:"required"`
}
