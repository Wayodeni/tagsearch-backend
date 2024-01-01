package repository

import (
	"errors"
	"fmt"

	"github.com/Wayodeni/tagsearch-backend/internal/storage/models"
	"github.com/jmoiron/sqlx"
)

var (
	ErrTransactionOpen = errors.New("error on transaction opening")
)

type TagAssigner interface {
	AssignForDocument(tx *sqlx.Tx, documentID models.ID, tags []models.TagResponse) (err error)
	ListForDocument(tx *sqlx.Tx, documentID models.ID) (response []models.TagResponse, err error)
	DeleteForDocument(tx *sqlx.Tx, documentID models.ID, tags []models.TagResponse) (err error)
}

type DocumentRepository struct {
	db            *sqlx.DB
	tagRepository TagAssigner
}

func NewDocumentRepository(db *sqlx.DB, tagRepository TagAssigner) *DocumentRepository {
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

	res, err := tx.Exec("INSERT INTO documents VALUES (NULL, ?, ?)", request.Name, request.Body)
	if err != nil {
		return response, err
	}

	documentID, err := res.LastInsertId()
	if err != nil {
		return response, err
	}

	if err := repository.tagRepository.AssignForDocument(tx, documentID, request.Tags); err != nil {
		return response, err
	}

	if err := tx.Commit(); err != nil {
		return response, err
	}

	return models.DocumentResponse{
		ID:   documentID,
		Name: request.Name,
		Body: request.Body,
		Tags: request.Tags,
	}, nil
}

func (repository *DocumentRepository) Read(id models.ID) (response models.DocumentResponse, err error) {
	// TODO: Investigate who is faster: single SQL with join or two queries for nested structure
	tx, err := repository.db.Beginx()
	if err != nil {
		return response, ErrTransactionOpen
	}
	defer tx.Rollback()

	if err := tx.Get(&response, "SELECT id, name, body FROM documents WHERE id = ?", id); err != nil {
		return response, err
	}

	if err := repository.setDocumentTags(tx, &response); err != nil {
		return response, err
	}

	if err := tx.Commit(); err != nil {
		return response, err
	}

	return response, nil
}

func (repository *DocumentRepository) setDocumentTags(tx *sqlx.Tx, documentResponse *models.DocumentResponse) (err error) {
	tags, err := repository.tagRepository.ListForDocument(tx, documentResponse.ID)
	if err != nil {
		return err
	}
	if len(tags) > 0 {
		documentResponse.Tags = tags
	}
	return nil
}

func (repository *DocumentRepository) ReadMany(IDs []models.ID) (response []models.DocumentResponse, err error) {
	if len(IDs) == 0 {
		return response, nil
	}

	query, args, err := sqlx.In("SELECT id, name, body FROM documents WHERE id IN (?)", IDs)
	if err != nil {
		return response, fmt.Errorf("unable to rebind query for slice usage in sqlx.In: %w", err)
	}

	tx, err := repository.db.Beginx()
	if err != nil {
		return response, ErrTransactionOpen
	}
	defer tx.Rollback()

	if err := tx.Select(&response, tx.Rebind(query), args...); err != nil {
		return response, err
	}

	for i := 0; i < len(response); i++ {
		if err := repository.setDocumentTags(tx, &response[i]); err != nil {
			return response, err
		}
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

	if updateRequest.Name.Valid {
		if _, err := tx.Exec("UPDATE documents SET name = ? WHERE id = ?", updateRequest.Name.String, id); err != nil {
			return response, err
		}
	}

	if updateRequest.Body.Valid {
		if _, err := tx.Exec("UPDATE documents SET body = ? WHERE id = ?", updateRequest.Body.String, id); err != nil {
			return response, err
		}
	}

	if len(updateRequest.TagsToAdd) > 0 {
		if err := repository.tagRepository.AssignForDocument(tx, id, updateRequest.TagsToAdd); err != nil {
			return response, err
		}
	}

	if len(updateRequest.TagsToRemove) > 0 {
		if err := repository.tagRepository.DeleteForDocument(tx, id, updateRequest.TagsToRemove); err != nil {
			return response, err
		}
	}

	if err := tx.Commit(); err != nil {
		return response, err
	}

	return repository.Read(id)
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

	for i := 0; i < len(response); i++ {
		repository.setDocumentTags(tx, &response[i])
	}

	if err := tx.Commit(); err != nil {
		return response, err
	}

	return response, nil
}
