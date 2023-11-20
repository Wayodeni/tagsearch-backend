package utilities

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"

	service "github.com/Wayodeni/tagsearch-backend/internal/service/index"
	"github.com/Wayodeni/tagsearch-backend/internal/storage/models"
	"github.com/blevesearch/bleve/v2"
)

// url,title,text,topic,tags
const (
	URL_COL int = iota
	TITLE_COL
	TEXT_COL
	TOPIC_COL
	TAG_COL
)

const RECORDS_TO_INDEX_QUANTITY = 180000

// Returns manually added documents that will be checked for equality in FindOne and FindMany tests
func GetExpectedSearchResults(latestDatasetDocumentIndex int) (results []models.DocumentResponse) {
	expectedSearchResults := []models.DocumentResponse{
		models.DocumentResponse{
			ID:   int64(latestDatasetDocumentIndex),
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
			ID:   int64(latestDatasetDocumentIndex + 1),
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

func LoadTestData(filePath string) (testData []models.DocumentResponse) {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Unable to read input file "+filePath, err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	csvReader.Comment = '#'

	tagset := map[string]struct{}{}
	for i := 0; i < RECORDS_TO_INDEX_QUANTITY; i++ {
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

	testData = append(testData, GetExpectedSearchResults(RECORDS_TO_INDEX_QUANTITY)...)

	return testData
}

type MockDocumentRepository struct {
	store map[models.ID]models.DocumentResponse
}

func NewMockDocumentRepository(documents []models.DocumentResponse) *MockDocumentRepository {
	store := map[models.ID]models.DocumentResponse{}
	for _, document := range documents {
		store[int64(document.ID)] = models.DocumentResponse{
			ID:   int64(document.ID),
			Name: document.Name,
			Body: document.Body,
			Tags: document.Tags,
		}
	}
	return &MockDocumentRepository{
		store: store,
	}
}

type MockTagRepository struct {
	store map[models.ID]models.TagResponse
}

func NewMockTagRepository(documents []models.DocumentResponse) *MockTagRepository {
	store := map[models.ID]models.TagResponse{}
	for _, document := range documents {
		tags := document.Tags
		for _, tag := range tags {
			store[int64(tag.ID)] = models.TagResponse{
				ID:   int64(tag.ID),
				Name: tag.Name,
			}
		}
	}
	return &MockTagRepository{
		store: store,
	}
}

func (repository *MockTagRepository) List() (response []models.TagResponse, err error) {
	for _, tag := range repository.store {
		response = append(response, tag)
	}
	return response, nil
}

func (repository *MockDocumentRepository) ReadMany(IDs []models.ID) (response []models.DocumentResponse, err error) {
	for _, id := range IDs {
		if document, ok := repository.store[id]; ok {
			response = append(response, document)
		}
	}
	return response, nil
}

func getTestIndex() (index bleve.Index, indexTestData bool) {
	const testIndexName = "test_index.bleve"

	if _, err := os.Stat(testIndexName); os.IsNotExist(err) {
		indexTestData = true
		fmt.Println("test index not found - creating new")
		index, err = bleve.New(testIndexName, service.GetIndexMapping())
		if err != nil {
			panic(err)
		}
	} else {
		fmt.Println("using already existing test bleve index")
		index, err = bleve.Open(testIndexName)
		if err != nil {
			panic(err)
		}
	}

	return index, indexTestData
}

func NewTestIndexService(testDataPath string) (*service.IndexService, []models.DocumentResponse, func()) {

	testData := LoadTestData(testDataPath)

	index, indexTestData := getTestIndex()

	indexService := service.NewIndexService(
		index,
		NewMockDocumentRepository(testData),
		NewMockTagRepository(testData),
	)

	if indexTestData {
		indexService.Index(testData)
	}

	return indexService,
		testData,
		func() { index.Close() }
}
