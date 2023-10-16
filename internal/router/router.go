package router

import (
	service "github.com/Wayodeni/tagsearch-backend/internal/service/index"
	"github.com/Wayodeni/tagsearch-backend/internal/storage/repository"
	"github.com/gin-gonic/gin"
)

func NewRouter(tagRepository *repository.TagRepository, documentRepository *repository.DocumentRepository, indexService *service.IndexService) *gin.Engine {

	r := gin.Default()
	// api := r.Group("/api")
	// {
	// 	v1 := api.Group("/v1")
	// 	{

	// 	}
	// }

	return r
}
