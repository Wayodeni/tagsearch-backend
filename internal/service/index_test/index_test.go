package index_test

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"testing"

	indexservice "github.com/Wayodeni/tagsearch-backend/internal/service/index"
	"github.com/Wayodeni/tagsearch-backend/internal/storage/models"
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/lang/ru"
	"github.com/stretchr/testify/require"
)

// url,title,text,topic,tags
const (
	URL_COL int = iota
	TITLE_COL
	TEXT_COL
	TOPIC_COL
	TAG_COL
)

// Returns manually added documents that will be checked for equality in FindOne and FindMany tests
func getExpectedSearchResults(latestDocumentIndex int) (results []models.DocumentResponse) {
	expectedSearchResults := []models.DocumentResponse{
		models.DocumentResponse{
			ID:   int64(latestDocumentIndex),
			Name: "Тестовый документ для поиска qwerty",
			Body: "Это тестовое описание документа для индекса",
			Tags: []models.TagResponse{
				{
					ID:   1,
					Name: "Тег 1",
				},
				{
					ID:   2,
					Name: "Тег 2",
				},
				{
					ID:   3,
					Name: "Тег 3",
				},
				{
					ID:   4,
					Name: "Тег 4",
				},
				{
					ID:   999,
					Name: "общий тег",
				},
			},
		},
		models.DocumentResponse{
			ID:   int64(latestDocumentIndex + 1),
			Name: "Тестовый док для поиска 2 qwerty",
			Body: "Второй тестовый документ",
			Tags: []models.TagResponse{
				{
					ID:   5,
					Name: "Тег 5",
				},
				{
					ID:   6,
					Name: "Тег 6",
				},
				{
					ID:   7,
					Name: "Тег 7",
				},
				{
					ID:   8,
					Name: "Тег 8",
				},
				{
					ID:   999,
					Name: "общий тег",
				},
			},
		},
	}

	return expectedSearchResults
}

func loadTestData(filePath string) (testData []models.DocumentResponse) {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Unable to read input file "+filePath, err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	csvReader.Comment = '#'

	tagset := map[string]struct{}{}
	const indexedDocsQuantity = 5000
	for i := 0; i < indexedDocsQuantity; i++ {
		if i == 0 {
			continue
		}
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		if record[TOPIC_COL] != "" {
			tagset[record[TOPIC_COL]] = struct{}{}
		}

		if record[TAG_COL] != "" {
			tagset[record[TAG_COL]] = struct{}{}
		}

		document := models.DocumentResponse{
			ID:   int64(i),
			Name: record[TITLE_COL],
			Body: record[TEXT_COL],
			Tags: []models.TagResponse{
				{
					ID:   int64(i),
					Name: record[TOPIC_COL],
				},
				{
					ID:   int64(i + 1),
					Name: record[TAG_COL],
				},
			},
		}
		testData = append(testData, document)
	}
	if err != nil {
		log.Fatal("Unable to parse file as CSV for "+filePath, err)
	}

	tags := make([]string, 0, len(tagset))
	for tag := range tagset {
		tags = append(tags, fmt.Sprintf("'%s'", tag))
	}
	fmt.Println("Tags: ", tags)

	testData = append(testData, getExpectedSearchResults(indexedDocsQuantity)...)

	return testData
}

type mockDocumentRepository struct {
	store map[models.ID]models.DocumentResponse
}

func newMockDocumentRepository(documents []models.DocumentResponse) *mockDocumentRepository {
	store := map[models.ID]models.DocumentResponse{}
	for _, document := range documents {
		store[int64(document.ID)] = models.DocumentResponse{
			ID:   int64(document.ID),
			Name: document.Name,
			Body: document.Body,
			Tags: document.Tags,
		}
	}
	return &mockDocumentRepository{
		store: store,
	}
}

func (repository *mockDocumentRepository) ReadMany(IDs []models.ID) (response []models.DocumentResponse, err error) {
	for _, id := range IDs {
		if document, ok := repository.store[id]; ok {
			response = append(response, document)
		}
	}
	return response, nil
}

func newTestIndexService() (*indexservice.IndexService, []models.DocumentResponse, func()) {
	testData := loadTestData("lenta-ru-news.csv")

	indexMapping := bleve.NewIndexMapping()
	documentMapping := bleve.NewDocumentMapping()

	documentNameFieldMapping := bleve.NewTextFieldMapping()
	documentNameFieldMapping.Analyzer = ru.AnalyzerName
	documentMapping.AddFieldMappingsAt("name", documentNameFieldMapping)

	documentBodyFieldMapping := bleve.NewTextFieldMapping()
	documentNameFieldMapping.Analyzer = ru.AnalyzerName
	documentMapping.AddFieldMappingsAt("body", documentBodyFieldMapping)

	documentTagsFieldMapping := bleve.NewTextFieldMapping()
	documentNameFieldMapping.Analyzer = "keyword"
	documentMapping.AddFieldMappingsAt("tags", documentTagsFieldMapping)

	indexMapping.AddDocumentMapping("document", documentMapping)

	index, err := bleve.NewMemOnly(indexMapping)
	if err != nil {
		panic(err)
	}

	return indexservice.NewIndexService(
			index,
			newMockDocumentRepository(testData),
		),
		testData,
		func() { index.Close() }
}

func Test_Find_One(t *testing.T) {
	service, testData, cleanupFunc := newTestIndexService()
	expectedSearchResults := getExpectedSearchResults(len(testData) - 1)
	defer cleanupFunc()

	err := service.Index(testData)
	require.NoError(t, err)

	searchResponse, err := service.Find(&indexservice.SearchDocumentRequest{
		Query: expectedSearchResults[0].Name,
		Tags:  expectedSearchResults[0].Tags,
	})

	require.NoError(t, err)
	require.Equal(t, expectedSearchResults[0], searchResponse[0])

}

func Test_Find_Many(t *testing.T) {
	service, testData, cleanupFunc := newTestIndexService()
	expectedSearchResults := getExpectedSearchResults(len(testData) - 1)
	defer cleanupFunc()

	err := service.Index(testData)
	require.NoError(t, err)

	searchResponse, err := service.Find(&indexservice.SearchDocumentRequest{
		Query: "",
		Tags: []models.TagResponse{
			{
				ID:   999,
				Name: "общий тег",
			},
		},
	})

	require.NoError(t, err)
	require.Equal(t, expectedSearchResults, searchResponse)

}
