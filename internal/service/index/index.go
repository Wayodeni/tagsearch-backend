package service

import (
	"fmt"
	"strconv"

	"github.com/Wayodeni/tagsearch-backend/internal/storage/models"
	"github.com/blevesearch/bleve/v2"
)

type SearchDocumentRequest struct {
	Query string   `form:"query" json:"query"`
	Tags  []string `form:"tags" json:"tags"`
}

type TagName = string
type DocumentCount = int
type SearchResponse struct {
	Documents []models.DocumentResponse `json:"documents"`
	Tags      []TagBucket               `json:"tags"`
}

type TagBucket struct {
	TagName       TagName       `json:"name"`
	DocumentCount DocumentCount `json:"documentCount"`
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

type Lister interface {
	List() (response []models.TagResponse, err error)
}

type IndexService struct {
	index              bleve.Index
	documentRepository ReadManyer
	tagRepository      Lister
}

func NewIndexService(index bleve.Index, documentRepository ReadManyer, tagRepository Lister) *IndexService {
	return &IndexService{
		index:              index,
		documentRepository: documentRepository,
		tagRepository:      tagRepository,
	}
}

func (service *IndexService) Find(searchQuery *SearchDocumentRequest) (response SearchResponse, err error) {
	booleanQuery := bleve.NewBooleanQuery()
	matchQuery := bleve.NewQueryStringQuery(searchQuery.Query)

	// Adding match query only if it presents
	if searchQuery.Query != "" {
		booleanQuery.AddMust(matchQuery)
	}

	// Adding term query per tag if any tags present
	if len(searchQuery.Tags) > 0 {
		for _, tag := range searchQuery.Tags {
			termQuery := bleve.NewTermQuery(tag)
			termQuery.SetField("tags")
			booleanQuery.AddMust(termQuery)
		}
	}

	// If search request don't contain querystring or tags we searching for all docs or using built query otherwise
	var searchRequest *bleve.SearchRequest
	if len(searchQuery.Tags) == 0 && searchQuery.Query == "" {
		searchRequest = bleve.NewSearchRequest(bleve.NewMatchAllQuery())
	} else {
		searchRequest = bleve.NewSearchRequest(booleanQuery)
	}

	// Getting all tags list to get tags quantity for facet request
	allTags, err := service.tagRepository.List()
	if err != nil {
		return response, fmt.Errorf("unable to get List of all tags: %w", err)
	}

	// Adding facet request to include all tags in response
	searchRequest.AddFacet("tags", bleve.NewFacetRequest("tags", len(allTags)))

	// Getting search results with using search request
	results, err := service.index.Search(searchRequest)
	if err != nil {
		return response, err
	}

	// Collecting document IDs from search result to get them from DB
	IDs := make([]models.ID, 0, results.Size())
	for _, match := range results.Hits {
		idstr, _ := strconv.Atoi(match.ID)
		IDs = append(IDs, int64(idstr))
	}

	// Getting found docs by id from DB
	foundDocuments, _ := service.documentRepository.ReadMany(IDs)

	// Getting all tags with count from documents found by query
	terms := results.Facets["tags"].Terms.Terms()
	foundTags := make([]TagBucket, 0, len(allTags))
	for _, term := range terms {
		foundTags = append(foundTags, TagBucket{
			TagName:       term.Term,
			DocumentCount: term.Count,
		})
	}

	return SearchResponse{
		Documents: foundDocuments,
		Tags:      foundTags,
	}, nil
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
