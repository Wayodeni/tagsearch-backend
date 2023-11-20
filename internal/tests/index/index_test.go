package utilities

import (
	"testing"

	indexService "github.com/Wayodeni/tagsearch-backend/internal/service/index"
	"github.com/stretchr/testify/require"
)

const testDataPath = "lenta-ru-news.csv"

func Test_Find_One(t *testing.T) {
	service, testData, cleanupFunc := NewTestIndexService(testDataPath)
	expectedSearchResults := GetExpectedSearchResults(len(testData) - 1)
	defer cleanupFunc()

	err := service.Index(testData)
	require.NoError(t, err)

	searchResponse, err := service.Find(&indexService.SearchDocumentRequest{
		Query: expectedSearchResults[0].Name,
		Tags:  expectedSearchResults[0].TagNames(),
	})

	require.NoError(t, err)
	require.Equal(t, expectedSearchResults[0], searchResponse[0])

}

func Test_Find_Many(t *testing.T) {
	service, testData, cleanupFunc := NewTestIndexService(testDataPath)
	expectedSearchResults := GetExpectedSearchResults(len(testData) - 1)
	defer cleanupFunc()

	err := service.Index(testData)
	require.NoError(t, err)

	searchResponse, err := service.Find(&indexService.SearchDocumentRequest{
		Query: "",
		Tags:  []string{"общий тег"},
	},
	)
	require.NoError(t, err)
	require.Equal(t, expectedSearchResults, searchResponse)

}
