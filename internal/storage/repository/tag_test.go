package repository

import (
	"testing"

	"github.com/Wayodeni/tagsearch-backend/internal/storage/db"
	"github.com/Wayodeni/tagsearch-backend/internal/storage/models"
	"github.com/stretchr/testify/require"
)

func newTestTagRepository() (testRepository *TagRepository, cleanupFunc func()) {
	db := db.NewDb(":memory:")
	repository := NewTagRepository(db)
	return repository, func() { db.Close() }
}

func Test_Create_Tag(t *testing.T) {
	repository, cleanupFunc := newTestTagRepository()
	defer cleanupFunc()

	const testTagName = "test tag"
	actual, err := repository.Create(models.CreateTagRequest{
		Name: testTagName,
	})

	require.NoError(t, err)
	require.Equal(t, testTagName, actual.Name)
}

func Test_Read_Tag(t *testing.T) {
	repository, cleanupFunc := newTestTagRepository()
	defer cleanupFunc()

	const testTagName = "test tag"
	createdTag, _ := repository.Create(models.CreateTagRequest{
		Name: testTagName,
	})

	actual, err := repository.Read(createdTag.ID)

	require.NoError(t, err)
	require.Equal(t, createdTag.ID, actual.ID)
}

func Test_ReadMany_Tags(t *testing.T) {
	repository, cleanupFunc := newTestTagRepository()
	defer cleanupFunc()

	createRequests := [...]models.CreateTagRequest{
		{Name: "test tag 0"},
		{Name: "test tag 1"},
		{Name: "test tag 2"},
		{Name: "test tag 3"},
		{Name: "test tag 4"},
		{Name: "test tag 5"},
	}

	createdTagIDs := make([]int64, 0, len(createRequests))
	for _, tag := range createRequests {
		createdTag, _ := repository.Create(models.CreateTagRequest{
			Name: tag.Name,
		})
		createdTagIDs = append(createdTagIDs, createdTag.ID)
	}

	penultimate := len(createRequests) - 1
	IDsToRead := createdTagIDs[0:penultimate]

	actual, err := repository.ReadMany(IDsToRead)

	require.NoError(t, err)
	for i, tag := range actual {
		require.Equal(
			t,
			models.TagResponse{
				ID:   IDsToRead[i],
				Name: createRequests[i].Name,
			},
			tag,
		)
	}
}

func Test_Update_Tag(t *testing.T) {
	repository, cleanupFunc := newTestTagRepository()
	defer cleanupFunc()

	const testTagName = "test tag"
	createdTag, _ := repository.Create(models.CreateTagRequest{
		Name: testTagName,
	})

	const newTagName = "new tag name"
	actual, err := repository.Update(createdTag.ID, models.UpdateTagRequest{
		Name: newTagName,
	})

	require.NoError(t, err)
	require.Equal(t, newTagName, actual.Name)
	require.Equal(t, createdTag.ID, actual.ID)
}

func Test_Delete_Tag(t *testing.T) {
	repository, cleanupFunc := newTestTagRepository()
	defer cleanupFunc()

	const testTagName = "test tag"
	createdTag, _ := repository.Create(models.CreateTagRequest{
		Name: testTagName,
	})

	require.NoError(t, repository.Delete(createdTag.ID))

	actual, err := repository.List()
	require.NoError(t, err)
	require.Equal(t, 0, len(actual))
}

func Test_List_Tags(t *testing.T) {
	repository, cleanupFunc := newTestTagRepository()
	defer cleanupFunc()

	createRequests := [...]models.CreateTagRequest{
		{Name: "test tag 0"},
		{Name: "test tag 1"},
		{Name: "test tag 2"},
	}

	createdTagIDs := make([]int64, 0, len(createRequests))
	for _, tag := range createRequests {
		createdTag, _ := repository.Create(models.CreateTagRequest{
			Name: tag.Name,
		})
		createdTagIDs = append(createdTagIDs, createdTag.ID)
	}

	actual, err := repository.List()
	require.NoError(t, err)
	for i, tag := range actual {
		require.Equal(
			t,
			models.TagResponse{
				ID:   createdTagIDs[i],
				Name: createRequests[i].Name,
			},
			tag,
		)
	}
}
