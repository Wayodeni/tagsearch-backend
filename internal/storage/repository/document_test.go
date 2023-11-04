package repository

import (
	"sort"
	"testing"

	"github.com/Wayodeni/tagsearch-backend/internal/storage/db"
	"github.com/Wayodeni/tagsearch-backend/internal/storage/models"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

func testDocumentRepository() (repository *DocumentRepository, cleanupFunc func()) {
	db := db.NewDb(":memory:")

	repository = NewDocumentRepository(db, NewTagRepository(db))
	return repository, func() { repository.db.Close() }
}

func Test_Create_Document(t *testing.T) {
	repository, cleanupFunc := testDocumentRepository()
	defer cleanupFunc()

	testTags := [...]models.CreateTagRequest{
		{
			Name: "tag 0",
		},
		{
			Name: "tag 1",
		},
		{
			Name: "tag 2",
		},
	}

	createdTags := make([]models.TagResponse, 0, len(testTags))
	for _, tag := range testTags {
		createdTag, _ := repository.tagRepository.Create(tag)
		createdTags = append(createdTags, createdTag)
	}

	testDocuments := [...]models.CreateDocumentRequest{
		{
			Name: "test document without tag",
			Body: "this is test document body",
			Tags: []models.TagResponse{},
		},
		{
			Name: "test document with single tag",
			Body: "this is test document body",
			Tags: []models.TagResponse{
				createdTags[0],
			},
		},
		{
			Name: "test document with multiple tags",
			Body: "this is test document body",
			Tags: []models.TagResponse{
				createdTags[0],
				createdTags[1],
				createdTags[2],
			},
		},
	}

	for i, createDocumentRequest := range testDocuments {
		actual, err := repository.Create(createDocumentRequest)
		require.NoError(t, err)
		require.Equal(
			t,
			models.DocumentResponse{
				ID:   actual.ID,
				Name: testDocuments[i].Name,
				Body: testDocuments[i].Body,
				Tags: testDocuments[i].Tags,
			},
			actual,
		)
	}
}

func Test_Read_Document(t *testing.T) {
	repository, cleanupFunc := testDocumentRepository()
	defer cleanupFunc()

	testTags := [...]models.CreateTagRequest{
		{
			Name: "tag 0",
		},
		{
			Name: "tag 1",
		},
		{
			Name: "tag 2",
		},
	}

	createdTags := make([]models.TagResponse, 0, len(testTags))
	for _, tag := range testTags {
		createdTag, _ := repository.tagRepository.Create(tag)
		createdTags = append(createdTags, createdTag)
	}

	testDocuments := [...]models.CreateDocumentRequest{
		{
			Name: "test document without tag",
			Body: "this is test document body",
		},
		{
			Name: "test document with single tag",
			Body: "this is test document body",
			Tags: []models.TagResponse{
				createdTags[0],
			},
		},
		{
			Name: "test document with multiple tags",
			Body: "this is test document body",
			Tags: []models.TagResponse{
				createdTags[0],
				createdTags[1],
				createdTags[2],
			},
		},
	}

	for _, createDocumentRequest := range testDocuments {
		expected, _ := repository.Create(createDocumentRequest)

		actual, err := repository.Read(expected.ID)
		require.NoError(t, err)
		require.Equal(
			t,
			expected,
			actual,
		)
	}
}

func Test_ReadMany_Documents(t *testing.T) {
	repository, cleanupFunc := testDocumentRepository()
	defer cleanupFunc()

	testTags := [...]models.CreateTagRequest{
		{
			Name: "tag 0",
		},
		{
			Name: "tag 1",
		},
		{
			Name: "tag 2",
		},
	}

	createdTags := make([]models.TagResponse, 0, len(testTags))
	for _, tag := range testTags {
		createdTag, _ := repository.tagRepository.Create(tag)
		createdTags = append(createdTags, createdTag)
	}

	testDocuments := [...]models.CreateDocumentRequest{
		{
			Name: "test document without tag",
			Body: "this is test document body",
		},
		{
			Name: "test document with single tag",
			Body: "this is test document body",
			Tags: []models.TagResponse{
				createdTags[0],
			},
		},
		{
			Name: "test document with multiple tags",
			Body: "this is test document body",
			Tags: []models.TagResponse{
				createdTags[0],
				createdTags[1],
				createdTags[2],
			},
		},
	}

	createdDocuments := make([]models.DocumentResponse, 0, len(testDocuments))
	for _, createDocumentRequest := range testDocuments {
		createdDocument, _ := repository.Create(createDocumentRequest)
		createdDocuments = append(createdDocuments, createdDocument)
	}

	expected := []models.DocumentResponse{createdDocuments[0], createdDocuments[2]}

	actual, err := repository.ReadMany([]models.ID{createdDocuments[0].ID, createdDocuments[2].ID})
	require.NoError(t, err)
	require.Equal(
		t,
		expected,
		actual,
	)
}

func Test_Update_Document(t *testing.T) {
	repository, cleanupFunc := testDocumentRepository()
	defer cleanupFunc()

	testTags := [...]models.CreateTagRequest{
		{
			Name: "tag 0",
		},
		{
			Name: "tag 1",
		},
		{
			Name: "tag 2",
		},
		{
			Name: "tag 3",
		},
		{
			Name: "tag 4",
		},
		{
			Name: "tag 5",
		},
	}

	createdTags := make([]models.TagResponse, 0, len(testTags))
	for _, tag := range testTags {
		createdTag, _ := repository.tagRepository.Create(tag)
		createdTags = append(createdTags, createdTag)
	}

	createdDocument, _ := repository.Create(models.CreateDocumentRequest{
		Name: "test document",
		Body: "test document body",
		Tags: []models.TagResponse{createdTags[0], createdTags[1], createdTags[2], createdTags[3]},
	})

	updatedDocumentName := "updated document"
	updatedDocumentBody := "updated body"
	expected := models.DocumentResponse{
		ID:   createdDocument.ID,
		Name: updatedDocumentName,
		Body: updatedDocumentBody,
		Tags: []models.TagResponse{createdTags[4], createdTags[5]},
	}

	actual, err := repository.Update(createdDocument.ID, models.UpdateDocumentRequest{
		Name:         null.NewString(updatedDocumentName, true),
		Body:         null.NewString(updatedDocumentBody, true),
		TagsToAdd:    []models.TagResponse{createdTags[4], createdTags[5]},
		TagsToRemove: []models.TagResponse{createdTags[0], createdTags[1], createdTags[2], createdTags[3]},
	})
	require.NoError(t, err)
	require.Equal(
		t,
		expected,
		actual,
	)
}

func Test_Delete_Document(t *testing.T) {
	repository, cleanupFunc := testDocumentRepository()
	defer cleanupFunc()

	createdDocument, _ := repository.Create(models.CreateDocumentRequest{
		Name: "test name",
		Body: "test body",
	})

	require.NoError(t, repository.Delete(createdDocument.ID))

	actual, _ := repository.Read(createdDocument.ID)
	require.Equal(t, models.DocumentResponse{}, actual)
}

func Test_List_Documents(t *testing.T) {
	repository, cleanupFunc := testDocumentRepository()
	defer cleanupFunc()

	testTags := [...]models.CreateTagRequest{
		{
			Name: "tag 0",
		},
		{
			Name: "tag 1",
		},
		{
			Name: "tag 2",
		},
	}

	createdTags := make([]models.TagResponse, 0, len(testTags))
	for _, tag := range testTags {
		createdTag, _ := repository.tagRepository.Create(tag)
		createdTags = append(createdTags, createdTag)
	}

	testDocuments := [...]models.CreateDocumentRequest{
		{
			Name: "test document without tag",
			Body: "this is test document body",
		},
		{
			Name: "test document with single tag",
			Body: "this is test document body",
			Tags: []models.TagResponse{
				createdTags[0],
			},
		},
		{
			Name: "test document with multiple tags",
			Body: "this is test document body",
			Tags: []models.TagResponse{
				createdTags[0],
				createdTags[1],
				createdTags[2],
			},
		},
	}

	createdDocuments := make([]models.DocumentResponse, 0, len(testDocuments))
	for _, createDocumentRequest := range testDocuments {
		createdDocument, _ := repository.Create(createDocumentRequest)
		createdDocuments = append(createdDocuments, createdDocument)
	}

	sort.Slice(createdDocuments, func(i, j int) bool {
		return createdDocuments[i].Name < createdDocuments[j].Name
	})

	actual, err := repository.List()
	require.NoError(t, err)
	require.Equal(
		t,
		createdDocuments,
		actual,
	)
}
