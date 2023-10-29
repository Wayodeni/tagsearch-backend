package repository

import (
	"errors"

	"github.com/Wayodeni/tagsearch-backend/internal/storage/models"
	"github.com/jmoiron/sqlx"
)

var (
	ErrTransactionOpen = errors.New("error on transaction opening")
)

type DocumentRepository struct {
	db            *sqlx.DB
	tagRepository *TagRepository
}

func NewDocumentRepository(db *sqlx.DB, tagRepository *TagRepository) *DocumentRepository {
	return &DocumentRepository{
		db:            db,
		tagRepository: tagRepository,
	}
}

func (repository *DocumentRepository) Create(request models.CreateDocumentRequest) (response models.DocumentResponse, err error) {
	tx, err := repository.db.Beginx()
	if err != nil {
		return response, ErrTransactionOpen
	}
	defer tx.Rollback()

	res, err := tx.Exec("INSERT INTO documents VALUES (?, ?)", request.Body, request.Tags)
	if err != nil {
		return response, err
	}

	documentID, err := res.LastInsertId()
	if err != nil {
		return response, err
	}

	if err := repository.assignTags(tx, documentID, request.Tags); err != nil {
		return response, err
	}

	if err := tx.Commit(); err != nil {
		return response, err
	}

	return repository.Read(documentID)
}

func (repository *DocumentRepository) assignTags(tx *sqlx.Tx, documentID models.ID, tags []models.TagResponse) error {
	// TODO: Sql generation for bulk insertion
	for _, tag := range tags {
		_, err := tx.Exec("INSERT INTO tags_documents VALUES (?, ?)", tag.ID, documentID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (repository *DocumentRepository) Read(id models.ID) (response models.DocumentResponse, err error) {
	tx, err := repository.db.Beginx()
	if err != nil {
		return response, ErrTransactionOpen
	}
	defer tx.Rollback()

	if err := tx.Get(&response, "SELECT id, name, body FROM documents WHERE id = ?", id); err != nil {
		return response, err
	}

	if err := tx.Commit(); err != nil {
		return response, err
	}

	return response, nil
}

func (repository *DocumentRepository) ReadMany(IDs []models.ID) (response []models.DocumentResponse, err error) {
	tx, err := repository.db.Beginx()
	if err != nil {
		return response, ErrTransactionOpen
	}
	defer tx.Rollback()

	if err := tx.Select(&response, "SELECT id, name, body FROM documents WHERE id IN (?)", IDs); err != nil {
		return response, err
	}

	if err := tx.Commit(); err != nil {
		return response, err
	}

	return response, nil
}

func (repository *DocumentRepository) Update(id models.ID, updateRequest models.UpdateDocumentRequest) (response models.DocumentResponse, err error) {
	tx, err := repository.db.Beginx()
	if err != nil {
		return response, ErrTransactionOpen
	}
	defer tx.Rollback()

	if updateRequest.Body.Valid {
		if _, err := tx.Exec("UPDATE documents SET body = ? WHERE id = ?", updateRequest.Body.String, id); err != nil {
			return response, err
		}
	}

	if len(updateRequest.TagsToAdd) > 0 {
		if err := repository.assignTags(tx, id, updateRequest.TagsToAdd); err != nil {
			return response, err
		}
	}

	if len(updateRequest.TagsToRemove) > 0 {
		if err := repository.removeTags(tx, id, updateRequest.TagsToRemove); err != nil {
			return response, err
		}
	}

	if err := tx.Commit(); err != nil {
		return response, err
	}

	return repository.Read(id)
}

func (repository *DocumentRepository) removeTags(tx *sqlx.Tx, documentID models.ID, tags []models.TagResponse) error {
	// TODO: Sql generation for bulk removal
	for _, tag := range tags {
		_, err := tx.Exec("DELETE FROM tags_documents WHERE tag = ? AND document = ?", tag.ID, documentID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (repository *DocumentRepository) Delete(id models.ID) (err error) {
	tx, err := repository.db.Beginx()
	if err != nil {
		return ErrTransactionOpen
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM documents WHERE id = ?", id); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (repository *DocumentRepository) List() (response []models.DocumentResponse, err error) {
	tx, err := repository.db.Beginx()
	if err != nil {
		return response, ErrTransactionOpen
	}
	defer tx.Rollback()

	if err := tx.Select(&response, "SELECT id, name, body FROM documents ORDER BY name"); err != nil {
		return response, err
	}

	if err := tx.Commit(); err != nil {
		return response, err
	}

	return response, nil
}
