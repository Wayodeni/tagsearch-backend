package main

import (
	"github.com/Wayodeni/tagsearch-backend/internal/router"
	"github.com/Wayodeni/tagsearch-backend/internal/storage/db"
	"github.com/Wayodeni/tagsearch-backend/internal/storage/repository"
	utilities "github.com/Wayodeni/tagsearch-backend/internal/tests/index"
)

func main() {
	db := db.NewDb(":memory:")
	tagRepository := repository.NewTagRepository(db)
	documentRepository := repository.NewDocumentRepository(db, tagRepository)

	testIndexService, _, _ := utilities.NewTestIndexService("../../internal/tests/index/lenta-ru-news.csv")

	r := router.NewRouter(tagRepository, documentRepository, testIndexService)
	r.Run()
}
