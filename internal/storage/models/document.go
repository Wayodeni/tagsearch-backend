package models

import "gopkg.in/guregu/null.v4"

type CreateDocumentRequest struct {
	Name string        `json:"name" binding:"required"`
	Body string        `json:"body" binding:"required"`
	Tags []TagResponse `json:"tags"`
}

type UpdateDocumentRequest struct {
	Body         null.String   `json:"body"`
	TagsToAdd    []TagResponse `json:"tagsToAdd" binding:"unique"`
	TagsToRemove []TagResponse `json:"tagsToRemove" binding:"unique"`
}

/*
For each set of tags (TagsToAdd, TagsToRemove) finds complement of the opposite set.

Example:
If some tags that are inside TagsToAdd slice also exist in TagsToRemove slice
this function removes these common tags from both of sets.

Before call:
TagsToAdd := []int{1, 2, 3, 4}
TagsToRemove := []int{2, 3, 5}

After call:
TagsToAdd := []int{1, 4}
TagsToRemove := []int{5}
*/
func (udr *UpdateDocumentRequest) RemoveCommonTags() {
	tagsSliceToMap := func(tags []TagResponse) map[ID]TagResponse {
		result := make(map[ID]TagResponse, len(tags))
		for _, tag := range tags {
			result[tag.ID] = tag
		}
		return result
	}

	tagsMapToSlice := func(tagsMap map[ID]TagResponse) []TagResponse {
		result := make([]TagResponse, 0, len(tagsMap))
		for _, tag := range tagsMap {
			result = append(result, tag)
		}
		return result
	}

	tagsToAddMap := tagsSliceToMap(udr.TagsToAdd)
	tagsToRemoveMap := tagsSliceToMap(udr.TagsToRemove)
	for tagToRemoveID := range tagsToRemoveMap {
		if _, ok := tagsToAddMap[tagToRemoveID]; ok {
			delete(tagsToAddMap, tagToRemoveID)
			delete(tagsToRemoveMap, tagToRemoveID)
		}
	}

	udr.TagsToAdd = tagsMapToSlice(tagsToAddMap)
	udr.TagsToRemove = tagsMapToSlice(tagsToRemoveMap)
}

type DocumentResponse struct {
	ID   ID            `json:"id" db:"id"`
	Name string        `json:"name" db:"name"`
	Body string        `json:"body" db:"body"`
	Tags []TagResponse `json:"tags"`
}
