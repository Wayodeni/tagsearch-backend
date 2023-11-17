package service

import (
	"fmt"
	"strconv"

	"github.com/Wayodeni/tagsearch-backend/internal/storage/models"
	"github.com/blevesearch/bleve/v2"
)

type SearchDocumentRequest struct {
	Query string
	Tags  []models.TagResponse
}

type IndexDocument struct {
	Name string   `json:"name"`
	Body string   `json:"body"`
	Tags []string `json:"tags"`
}

func (documentResponse *IndexDocument) Type() string {
	return "document"
}

type ReadManyer interface {
	ReadMany(IDs []models.ID) (response []models.DocumentResponse, err error)
}

type IndexService struct {
	index              bleve.Index
	documentRepository ReadManyer
}

func NewIndexService(index bleve.Index, documentRepository ReadManyer) *IndexService {
	return &IndexService{
		index:              index,
		documentRepository: documentRepository,
	}
}

func (service *IndexService) Find(searchRequest *SearchDocumentRequest) (response []models.DocumentResponse, err error) {
	booleanQuery := bleve.NewBooleanQuery()

	matchQuery := bleve.NewQueryStringQuery(searchRequest.Query)

	if searchRequest.Query != "" {
		booleanQuery.AddMust(matchQuery)
	}

	if len(searchRequest.Tags) > 0 {
		for _, tag := range searchRequest.Tags {
			termQuery := bleve.NewMatchQuery(tag.Name)
			termQuery.SetField("tags")
			booleanQuery.AddMust(termQuery)
		}
	}

	results, err := service.index.Search(bleve.NewSearchRequest(booleanQuery))
	if err != nil {
		return response, err
	}
	// fmt.Println(results)

	IDs := make([]models.ID, 0, results.Size())
	for _, match := range results.Hits {
		idstr, _ := strconv.Atoi(match.ID)
		IDs = append(IDs, int64(idstr))
	}

	response, _ = service.documentRepository.ReadMany(IDs)

	// for _, doc := range response {
	// 	strdoc, _ := json.MarshalIndent(doc, "", "     ")
	// 	fmt.Println(string(strdoc))
	// }

	return response, nil
}

func (service *IndexService) Index(documents []models.DocumentResponse) error {
	batch := service.index.NewBatch()
	for _, document := range documents {
		batch.Index(
			fmt.Sprint(document.ID),
			IndexDocument{
				Name: document.Name,
				Body: document.Body,
				Tags: document.TagNames(),
			},
		)
	}
	return service.index.Batch(batch)
}
