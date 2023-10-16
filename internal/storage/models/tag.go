package models

type CreateTagRequest struct {
	Name string
}

type UpdateTagRequest struct {
	Name string
}

type TagResponse struct {
	ID int
	Name string
}