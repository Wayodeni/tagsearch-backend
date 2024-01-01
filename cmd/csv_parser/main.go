package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"time"

	service "github.com/Wayodeni/tagsearch-backend/internal/service/index"
	"github.com/Wayodeni/tagsearch-backend/internal/storage/db"
	"github.com/Wayodeni/tagsearch-backend/internal/storage/models"
	"github.com/Wayodeni/tagsearch-backend/internal/storage/repository"
	"github.com/blevesearch/bleve/v2"
	"github.com/jmoiron/sqlx"
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
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
const FILE_PATH = "lenta-ru-news.csv"
const INDEX_BATCH_SIZE = 30000

type document struct {
	Title string
	Text  string
	Topic string
	Tag   string
}

type tagCreator interface {
	Create(request models.CreateTagRequest) (response models.TagResponse, err error)
}

type alwaysAssignedTagRepository struct {
	db            *sqlx.DB
	tagRepository *repository.TagRepository
}

func newAlwaysAssignedTagRepository(db *sqlx.DB, tagRepository *repository.TagRepository) *alwaysAssignedTagRepository {
	return &alwaysAssignedTagRepository{
		db:            db,
		tagRepository: tagRepository,
	}
}

func (repository *alwaysAssignedTagRepository) Create(request models.CreateTagRequest) (response models.TagResponse, err error) {
	res, err := repository.db.Exec("INSERT INTO tags VALUES (NULL, ?, true)", request.Name)
	if err != nil {
		return response, err
	}

	tagId, err := res.LastInsertId()
	if err != nil {
		return response, err
	}

	return models.TagResponse{
		ID:   tagId,
		Name: request.Name,
	}, nil
}

func (repository *alwaysAssignedTagRepository) AssignForDocument(tx *sqlx.Tx, documentID models.ID, tags []models.TagResponse) (err error) {
	return repository.tagRepository.AssignForDocument(tx, documentID, tags)
}
func (repository *alwaysAssignedTagRepository) ListForDocument(tx *sqlx.Tx, documentID models.ID) (response []models.TagResponse, err error) {
	return repository.tagRepository.ListForDocument(tx, documentID)
}
func (repository *alwaysAssignedTagRepository) DeleteForDocument(tx *sqlx.Tx, documentID models.ID, tags []models.TagResponse) (err error) {
	return repository.tagRepository.DeleteForDocument(tx, documentID, tags)
}
func (repository *alwaysAssignedTagRepository) List() (response []models.TagResponse, err error) {
	return repository.tagRepository.List()
}
func (repository *alwaysAssignedTagRepository) ReadManyByNames(names []string) (response []models.TagResponse, err error) {
	return repository.tagRepository.ReadManyByNames(names)
}

func loadTagsInDb(tagRepo tagCreator, tagNames []string) map[string]models.TagResponse {
	result := make(map[string]models.TagResponse)
	for _, name := range tagNames {
		createdTag, err := tagRepo.Create(models.CreateTagRequest{
			Name: name,
		})
		createdTag.Assigned = true // Setting to true because all of parsed tags will be assigned to documents

		if sqliteErr, ok := err.(*sqlite.Error); ok {
			code := sqliteErr.Code()
			if code == sqlite3.SQLITE_CONSTRAINT_UNIQUE {
				continue
			} else {
				panic(err)
			}
		}
		result[createdTag.Name] = createdTag
	}
	return result
}

func getTagsFromFile(filePath string) []string {
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

	}
	if err != nil {
		log.Fatal("Unable to parse file as CSV for "+filePath, err)
	}

	tags := make([]string, 0, len(tagset))
	for tag := range tagset {
		tags = append(tags, tag)
	}
	fmt.Println("Tags: ", tags)
	return tags
}

func getDocumentTags(document document, createdTags map[string]models.TagResponse) []models.TagResponse {
	tags := []models.TagResponse{}
	if tag, ok := createdTags[document.Topic]; ok {
		tags = append(tags, tag)
	}
	if tag, ok := createdTags[document.Tag]; ok {
		tags = append(tags, tag)
	}
	return tags
}

func loadDocumentsInDb(docRepo *repository.DocumentRepository, createdTags map[string]models.TagResponse, documents []document) []models.DocumentResponse {
	createdDocuments := []models.DocumentResponse{}
	pb := NewProgressBar(time.Now(), len(documents)-1, 1000)
	for _, document := range documents {
		pb.Increment()
		createdDocument, _ := docRepo.Create(models.CreateDocumentRequest{
			Name: document.Title,
			Body: document.Text,
			Tags: getDocumentTags(document, createdTags),
		})
		createdDocuments = append(createdDocuments, createdDocument)
	}
	return createdDocuments
}

func getDocumentsFromFile(filePath string) []document {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Unable to read input file "+filePath, err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	csvReader.Comment = '#'

	documents := []document{}
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

		documents = append(documents, document{
			Title: record[TITLE_COL],
			Text:  record[TEXT_COL],
			Topic: record[TOPIC_COL],
			Tag:   record[TAG_COL],
		})

	}
	if err != nil {
		log.Fatal("Unable to parse file as CSV for "+filePath, err)
	}

	return documents
}

func indexDocuments(indexService *service.IndexService, documents []models.DocumentResponse) {
	batchesQuantity := math.Ceil(float64(len(documents)) / float64(INDEX_BATCH_SIZE))
	batchStartIndex := 0
	batchStopIndex := batchStartIndex + INDEX_BATCH_SIZE
	pb := NewProgressBar(time.Now(), int(batchesQuantity), 1000)
	for i := 0; i < int(batchesQuantity); i++ {
		if batchStopIndex > len(documents) {
			batchStopIndex = len(documents)
		}
		indexService.Index(documents[batchStartIndex:batchStopIndex])
		pb.Increment()
		batchStartIndex = batchStopIndex
		batchStopIndex += INDEX_BATCH_SIZE
	}
}

func main() {
	db := db.NewDb("test_db.sqlite3")

	alwaysAssignedtagRepository := newAlwaysAssignedTagRepository(db, repository.NewTagRepository(db))
	documentRepository := repository.NewDocumentRepository(db, alwaysAssignedtagRepository)
	index, err := bleve.New("test_index.bleve", service.GetIndexMapping())
	if err != nil {
		panic(err)
	}
	tagNames := getTagsFromFile(FILE_PATH)
	fmt.Println("got all tag names in ram")
	createdTags := loadTagsInDb(alwaysAssignedtagRepository, tagNames)
	fmt.Println("got all db-written tags in ram")
	documents := getDocumentsFromFile(FILE_PATH)
	fmt.Println("got all documents in ram")
	fmt.Println("starting loading documents with tags into db")
	fmt.Printf("there are ~%d. be patient :)", RECORDS_TO_INDEX_QUANTITY)
	dbDocuments := loadDocumentsInDb(documentRepository, createdTags, documents)
	fmt.Println("successfully added to db")
	fmt.Println("starting to index...")
	indexDocuments(service.NewIndexService(index, documentRepository, alwaysAssignedtagRepository), dbDocuments)
	fmt.Println("FINISHED!!!")
}
