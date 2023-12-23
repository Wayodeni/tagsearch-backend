package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	service "github.com/Wayodeni/tagsearch-backend/internal/service/index"
	"github.com/Wayodeni/tagsearch-backend/internal/storage/db"
	"github.com/Wayodeni/tagsearch-backend/internal/storage/models"
	"github.com/Wayodeni/tagsearch-backend/internal/storage/repository"
	"github.com/blevesearch/bleve/v2"
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

type document struct {
	Title string
	Text  string
	Topic string
	Tag   string
}

func loadTagsInDb(tagRepo *repository.TagRepository, tagNames []string) map[string]models.TagResponse {
	result := make(map[string]models.TagResponse)
	for _, name := range tagNames {
		createdTag, err := tagRepo.Create(models.CreateTagRequest{
			Name: name,
		})

		if liteErr, ok := err.(*sqlite.Error); ok {
			code := liteErr.Code()
			if code == sqlite3.SQLITE_CONSTRAINT_UNIQUE {
				continue
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
	tags = append(tags, createdTags[document.Topic])
	tags = append(tags, createdTags[document.Tag])
	return tags
}

func loadDocumentsInDb(docRepo *repository.DocumentRepository, tagRepo *repository.TagRepository, createdTags map[string]models.TagResponse, documents []document) []models.DocumentResponse {
	const PERCENT_PRINT_PERIOD = 500
	createdDocuments := []models.DocumentResponse{}
	timeStart := time.Now()
	for index, document := range documents {
		if index%PERCENT_PRINT_PERIOD == 0 && int(time.Since(timeStart).Seconds()) != 0 {
			elapsedTime := int(time.Since(timeStart).Seconds())
			additionSpeed := index / elapsedTime
			timeLeft := (RECORDS_TO_INDEX_QUANTITY - index) / additionSpeed
			secondsLeft := time.Duration(timeLeft) * time.Second
			fmt.Printf("added to db %f%% documents\n", (float64(index)/float64(RECORDS_TO_INDEX_QUANTITY))*100)
			fmt.Printf("time left: %.0f hrs %.0f mins %.0f secs\n", secondsLeft.Hours(), secondsLeft.Minutes(), secondsLeft.Seconds())
			fmt.Printf("addition speed is %d docs/sec. WOW! \n", additionSpeed)
			fmt.Printf("docs left: %d \n", RECORDS_TO_INDEX_QUANTITY-index)
		}
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

func main() {
	db := db.NewDb("test_db.sqlite3")

	tagRepository := repository.NewTagRepository(db)
	documentRepository := repository.NewDocumentRepository(db, tagRepository)
	index, err := bleve.New("test_index.bleve", service.GetIndexMapping())
	if err != nil {
		panic(err)
	}
	indexService := service.NewIndexService(index, documentRepository, tagRepository)

	tagNames := getTagsFromFile(FILE_PATH)
	fmt.Println("got all tag names in ram")
	createdTags := loadTagsInDb(tagRepository, tagNames)
	fmt.Println("got all db-written tags in ram")
	documents := getDocumentsFromFile(FILE_PATH)
	fmt.Println("got all documents in ram")
	fmt.Println("starting loading documents with tags into db")
	fmt.Println("there are ~180000. be patient :)")
	loadedDocuments := loadDocumentsInDb(documentRepository, tagRepository, createdTags, documents)
	fmt.Println("parsed successfully and added to db")
	fmt.Println("indexing...")
	indexService.Index(loadedDocuments)
	fmt.Println("successfully finished. grab results!!!")
}
