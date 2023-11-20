package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	"github.com/Wayodeni/tagsearch-backend/internal/storage/db"
	"github.com/Wayodeni/tagsearch-backend/internal/storage/models"
	"github.com/Wayodeni/tagsearch-backend/internal/storage/repository"
	utilities "github.com/Wayodeni/tagsearch-backend/internal/tests/index"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newTestRouter() (*gin.Engine, func()) {
	db := db.NewDb(":memory:")
	tagRepository := repository.NewTagRepository(db)
	documentRepository := repository.NewDocumentRepository(db, tagRepository)

	testIndexService, _, indexCleanupFunc := utilities.NewTestIndexService("../tests/index/lenta-ru-news.csv")

	return NewRouter(tagRepository, documentRepository, testIndexService),
		func() {
			db.Close()
			indexCleanupFunc()
		}
}

const apiURL = "/api/v1/search?"

func Test_Find_One(t *testing.T) {
	router, cleanupFunc := newTestRouter()
	defer cleanupFunc()

	testCases := []string{
		"query=Тестовый&tags[]=Тег 6&tags[]=общий тег",
		"tags[]=Тег 6&tags[]=общий тег",
		"tags[]=Тег 6",
	}

	for _, queryParams := range testCases {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", apiURL+queryParams, nil)
		router.ServeHTTP(w, req)

		var actual []models.DocumentResponse
		json.Unmarshal(w.Body.Bytes(), &actual)

		require.Equal(
			t,
			[]models.DocumentResponse{utilities.GetExpectedSearchResults(utilities.RECORDS_TO_INDEX_QUANTITY)[1]},
			actual,
			"test case '%s' failed",
			queryParams,
		)
	}
}

func Test_Find_Many(t *testing.T) {
	router, cleanupFunc := newTestRouter()
	defer cleanupFunc()

	testCases := []string{
		"query=Тестовый&tags[]=общий тег",
		"tags[]=общий тег",
	}

	for _, queryParams := range testCases {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", apiURL+queryParams, nil)
		router.ServeHTTP(w, req)

		var actual []models.DocumentResponse
		json.Unmarshal(w.Body.Bytes(), &actual)

		sort.Slice(actual, func(i, j int) bool {
			return actual[i].ID < actual[j].ID
		})

		require.Equal(
			t,
			utilities.GetExpectedSearchResults(utilities.RECORDS_TO_INDEX_QUANTITY),
			actual,
			"test case '%s' failed",
			queryParams,
		)
	}
}

func Test_Find_None(t *testing.T) {
	router, cleanupFunc := newTestRouter()
	defer cleanupFunc()

	testCases := []string{
		"query=Тестовый&tags[]=Тег",
		"query=Тестовый&tags[]=общий",
		"query=Тестовый&tags[]=Тег",
		"query=Тестовый&tags[]=общий",
	}

	for _, queryParams := range testCases {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", apiURL+queryParams, nil)
		router.ServeHTTP(w, req)

		var actual []models.DocumentResponse
		json.Unmarshal(w.Body.Bytes(), &actual)

		require.Equal(t, len(actual), 0, "test case '%s' failed", queryParams)
	}
}
