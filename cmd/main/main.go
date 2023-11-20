package main

import (
	"errors"
	"fmt"

	"github.com/Wayodeni/tagsearch-backend/internal/config"
	"github.com/Wayodeni/tagsearch-backend/internal/router"
	service "github.com/Wayodeni/tagsearch-backend/internal/service/index"
	"github.com/Wayodeni/tagsearch-backend/internal/storage/db"
	"github.com/Wayodeni/tagsearch-backend/internal/storage/repository"
	"github.com/blevesearch/bleve/v2"
)

func main() {
	config, err := config.NewConfig()
	if err != nil {
		panic(err)
	}
	fmt.Println(config)

	db := db.NewDb(config.Db.Path)

	index, err := bleve.Open(config.Index.Path)
	if errors.Is(err, bleve.ErrorIndexPathDoesNotExist) {
		index, err = bleve.New(config.Index.Path, bleve.NewIndexMapping())
		if err != nil {
			panic(err)
		}
	} else if err != nil {
		panic(err)
	}

	tagRepository := repository.NewTagRepository(db)
	documentRepository := repository.NewDocumentRepository(db, tagRepository)
	indexService := service.NewIndexService(index, documentRepository, tagRepository)

	router := router.NewRouter(tagRepository, documentRepository, indexService)
	router.Run(fmt.Sprintf("%s:%s", config.App.Host, config.App.Port))
}
