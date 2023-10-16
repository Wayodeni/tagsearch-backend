package models

type CreateDocumentRequest struct {
	Name string
	Body string
	Tags []TagResponse
}

type UpdateDocumentRequest struct {
	Body         string
	TagsToAdd    []TagResponse
	TagsToRemove []TagResponse
}

type DocumentResponse struct {
	ID   int
	Name string
	Body string
	Tags []TagResponse
}
