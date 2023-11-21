package router

import (
	"github.com/Wayodeni/tagsearch-backend/internal/controllers"
	service "github.com/Wayodeni/tagsearch-backend/internal/service/index"
	"github.com/Wayodeni/tagsearch-backend/internal/storage/repository"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewRouter(tagRepository *repository.TagRepository, documentRepository *repository.DocumentRepository, indexService *service.IndexService) *gin.Engine {
	tagController := controllers.NewTagController(tagRepository)
	documentController := controllers.NewDocumentController(documentRepository, indexService)
	searchController := controllers.NewSearchController(indexService)

	r := gin.Default()
	r.Use(cors.Default())
	api := r.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			tags := v1.Group("/tags")
			{
				tags.POST("", tagController.Create)
				tags.GET("/:id", tagController.Read)
				tags.PATCH("/:id", tagController.Update)
				tags.DELETE("/:id", tagController.Delete)
				tags.GET("", tagController.List)
			}
			documents := v1.Group("/documents")
			{
				documents.POST("", documentController.Create)
				documents.GET("/:id", documentController.Read)
				documents.PATCH("/:id", documentController.Update)
				documents.DELETE("/:id", documentController.Delete)
				documents.GET("", documentController.List)
			}
			search := v1.Group("/search")
			{
				search.GET("", searchController.Search)
			}
		}
	}

	return r
}
