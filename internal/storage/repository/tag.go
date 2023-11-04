package repository

import (
	"fmt"

	"github.com/Wayodeni/tagsearch-backend/internal/storage/models"
	"github.com/jmoiron/sqlx"
)

type TagRepository struct {
	db *sqlx.DB
}

func NewTagRepository(db *sqlx.DB) *TagRepository {
	return &TagRepository{
		db: db,
	}
}

func (repository *TagRepository) Create(request models.CreateTagRequest) (response models.TagResponse, err error) {
	tx, err := repository.db.Beginx()
	if err != nil {
		return response, ErrTransactionOpen
	}
	defer tx.Rollback()

	res, err := tx.Exec("INSERT INTO tags VALUES (NULL, ?)", request.Name)
	if err != nil {
		return response, err
	}

	if err := tx.Commit(); err != nil {
		return response, err
	}

	tagId, err := res.LastInsertId()
	if err != nil {
		return response, err
	}

	return models.TagResponse{
		ID:   tagId,
		Name: request.Name,
	}, nil
}

func (repository *TagRepository) Read(id models.ID) (response models.TagResponse, err error) {
	tx, err := repository.db.Beginx()
	if err != nil {
		return response, ErrTransactionOpen
	}
	defer tx.Rollback()

	if err := tx.Get(&response, "SELECT id, name FROM tags WHERE id = ?", id); err != nil {
		return response, err
	}

	if err := tx.Commit(); err != nil {
		return response, err
	}

	return response, nil
}

func (repository *TagRepository) ReadMany(IDs []models.ID) (response []models.TagResponse, err error) {
	query, args, err := sqlx.In("SELECT id, name FROM tags WHERE id IN (?)", IDs)
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

	if err := tx.Commit(); err != nil {
		return response, err
	}

	return response, nil
}

func (repository *TagRepository) Update(id models.ID, updateRequest models.UpdateTagRequest) (response models.TagResponse, err error) {
	tx, err := repository.db.Beginx()
	if err != nil {
		return response, ErrTransactionOpen
	}
	defer tx.Rollback()

	if _, err := tx.Exec("UPDATE tags SET name = ? WHERE id = ?", updateRequest.Name, id); err != nil {
		return response, err
	}

	if err := tx.Commit(); err != nil {
		return response, err
	}

	return models.TagResponse{ID: id, Name: updateRequest.Name}, nil
}

func (repository *TagRepository) Delete(id models.ID) (err error) {
	tx, err := repository.db.Beginx()
	if err != nil {
		return ErrTransactionOpen
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM tags WHERE id = ?", id); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (repository *TagRepository) List() (response []models.TagResponse, err error) {
	tx, err := repository.db.Beginx()
	if err != nil {
		return response, ErrTransactionOpen
	}
	defer tx.Rollback()

	if err := tx.Select(&response, "SELECT id, name FROM tags"); err != nil {
		return response, err
	}

	if err := tx.Commit(); err != nil {
		return response, err
	}

	return response, nil
}

func (repository *TagRepository) ListForDocument(tx *sqlx.Tx, documentID models.ID) (response []models.TagResponse, err error) {
	query := `
	SELECT id, name FROM tags
	WHERE id IN (
		SELECT tag FROM tags_documents
		WHERE document = ?
	)
		`

	if tx == nil {
		tx, err = repository.db.Beginx()
		if err != nil {
			return response, ErrTransactionOpen
		}
		defer tx.Rollback()
	}

	if err := tx.Select(&response, query, documentID); err != nil {
		return response, err
	}

	if tx == nil {
		if err := tx.Commit(); err != nil {
			return response, err
		}
	}

	return response, nil
}

func (repository *TagRepository) AssignForDocument(tx *sqlx.Tx, documentID models.ID, tags []models.TagResponse) (err error) {
	if tx == nil {
		tx, err = repository.db.Beginx()
		if err != nil {
			return ErrTransactionOpen
		}
		defer tx.Rollback()
	}

	for _, tag := range tags {
		_, err := tx.Exec("INSERT INTO tags_documents VALUES (?, ?)", tag.ID, documentID)
		if err != nil {
			return err
		}
	}

	if tx == nil {
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

func (repository *TagRepository) DeleteForDocument(tx *sqlx.Tx, documentID models.ID, tags []models.TagResponse) (err error) {
	if tx == nil {
		tx, err = repository.db.Beginx()
		if err != nil {
			return ErrTransactionOpen
		}
		defer tx.Rollback()
	}

	for _, tag := range tags {
		_, err := tx.Exec("DELETE FROM tags_documents WHERE tag = ? AND document = ?", tag.ID, documentID)
		if err != nil {
			return err
		}
	}

	if tx == nil {
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}
