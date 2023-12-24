package service

import (
	"fmt"
	"math"
	"slices"
	"strconv"

	"github.com/Wayodeni/tagsearch-backend/internal/storage/models"
	"github.com/blevesearch/bleve/v2"
)

type SearchDocumentRequest struct {
	Query      string   `form:"query" json:"query"`
	Tags       []string `form:"tags" json:"tags"`
	PageSize   int      `form:"pageSize" json:"pageSize"`
	PageNumber int      `form:"pageNumber" json:"pageNumber"`
}

type TagName = string
type DocumentCount = int
type SearchResponse struct {
	Documents                []models.DocumentResponse `json:"documents,omitempty"`
	Tags                     []TagBucket               `json:"tags,omitempty"`
	DocumentsFound           int64                     `json:"documentsFound"`
	Pages                    int                       `json:"pages"`
	RequestPageIsOutOfBounds bool                      `json:"requestPageIsOutOfBounds"` // this flag tells frontend to change current page to Pages field of response
}

type TagBucket struct {
	models.TagResponse
	DocumentCount DocumentCount `json:"documentCount"`
	Selected      bool          `json:"selected"`
}

type IndexDocument struct {
	Name string   `json:"name"`
	Body string   `json:"body"`
	Tags []string `json:"tags"`
}

func (documentResponse *IndexDocument) Type() string {
	return "document"
}

type DocumentReadManyer interface {
	ReadMany(IDs []models.ID) (response []models.DocumentResponse, err error)
}

type TagNameLister interface {
	List() (response []models.TagResponse, err error)
	ReadManyByNames(names []string) (response []models.TagResponse, err error)
}

type IndexService struct {
	index              bleve.Index
	documentRepository DocumentReadManyer
	tagRepository      TagNameLister
}

func NewIndexService(index bleve.Index, documentRepository DocumentReadManyer, tagRepository TagNameLister) *IndexService {
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
	if searchQuery.Query != "" && searchQuery.Query[len(searchQuery.Query)-1] != '-' {
		booleanQuery.AddMust(matchQuery)
	}

	// Adding term query per tag if any tags present
	queryTags := make([]models.TagResponse, 0, len(searchQuery.Tags))
	if len(searchQuery.Tags) > 0 {
		for _, tag := range searchQuery.Tags {
			termQuery := bleve.NewTermQuery(tag)
			termQuery.SetField("tags")
			booleanQuery.AddMust(termQuery)
		}
		queryTags, err = service.tagRepository.ReadManyByNames(searchQuery.Tags)
		if err != nil {
			return response, fmt.Errorf("unable to get tags from database by names: %w", err)
		}
	}

	// If search request donesn't contain querystring or tags we searching for all docs or using built query otherwise
	var searchRequest *bleve.SearchRequest
	if len(searchQuery.Tags) == 0 && searchQuery.Query == "" {
		searchRequest = bleve.NewSearchRequestOptions(bleve.NewMatchAllQuery(), searchQuery.PageSize, searchQuery.PageNumber*searchQuery.PageSize, false)
	} else {
		searchRequest = bleve.NewSearchRequestOptions(booleanQuery, searchQuery.PageSize, searchQuery.PageNumber*searchQuery.PageSize, false)
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
	response.DocumentsFound = int64(results.Total)

	// Getting number of pages for result
	pages := int(math.Ceil(float64(results.Total) / float64(searchQuery.PageSize)))
	response.Pages = pages

	// If requested page is out of bounds change it to max page of search result
	// and tell frontend to change its page. Also update search results for new page number
	if searchQuery.PageNumber >= pages {
		response.RequestPageIsOutOfBounds = true
		searchRequest.From = pages
		results, err = service.index.Search(searchRequest)
		if err != nil {
			return response, err
		}
	}

	// Collecting document IDs from search result to get them from DB
	IDs := make([]models.ID, 0, results.Size())
	for _, match := range results.Hits {
		idstr, _ := strconv.Atoi(match.ID)
		IDs = append(IDs, int64(idstr))
	}
	// Getting found docs by id from DB
	foundDocuments, err := service.documentRepository.ReadMany(IDs)
	if err != nil {
		return response, fmt.Errorf("unable to ReadMany documents by IDs: %w", err)
	}
	response.Documents = foundDocuments

	// Getting all tags with count from documents found by query
	terms := results.Facets["tags"].Terms.Terms()
	foundTagsCount := make(map[TagName]DocumentCount, len(allTags))
	foundTagsNames := make([]TagName, 0, len(allTags))
	for _, term := range terms {
		foundTagsCount[term.Term] = term.Count
		foundTagsNames = append(foundTagsNames, term.Term)
	}

	// Getting additional metadata for tags from database
	tagResponses, err := service.tagRepository.ReadManyByNames(foundTagsNames)
	if err != nil {
		return response, fmt.Errorf("unable to get tags from db: %w", err)
	}
	for _, tag := range tagResponses {
		response.Tags = append(response.Tags, TagBucket{
			TagResponse:   tag,
			DocumentCount: foundTagsCount[tag.Name],
			Selected:      slices.Contains(searchQuery.Tags, tag.Name),
		})
	}

	// Adding selected tags with zero document count to response
	for _, tag := range queryTags {
		if !slices.Contains(tagResponses, tag) {
			response.Tags = append(response.Tags, TagBucket{
				TagResponse:   tag,
				DocumentCount: 0,
				Selected:      true,
			})
		}
	}

	return response, nil
}

// Perform batch document indexing or update
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
	// for _, document := range documents {
	// 	service.index.Index(
	// 		fmt.Sprint(document.ID),
	// 		IndexDocument{
	// 			Name: document.Name,
	// 			Body: document.Body,
	// 			Tags: document.TagNames(),
	// 		})
	// }
	// return nil
}

func (service *IndexService) Delete(IDs []models.ID) error {
	batch := service.index.NewBatch()
	for _, ID := range IDs {
		batch.Delete(
			fmt.Sprint(ID),
		)
	}
	return service.index.Batch(batch)
}
