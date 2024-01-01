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

	res, err := tx.Exec("INSERT INTO tags VALUES (NULL, ?, ?)", request.Name, 0)
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

	if err := tx.Get(&response, "SELECT id, name, assigned FROM tags WHERE id = ?", id); err != nil {
		return response, err
	}

	if err := tx.Commit(); err != nil {
		return response, err
	}

	return response, nil
}

func (repository *TagRepository) ReadMany(IDs []models.ID) (response []models.TagResponse, err error) {
	if len(IDs) == 0 {
		return response, nil
	}

	query, args, err := sqlx.In("SELECT id, name, assigned FROM tags WHERE id IN (?)", IDs)
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

func (repository *TagRepository) ReadManyByNames(names []string) (response []models.TagResponse, err error) {
	if len(names) == 0 {
		return response, nil
	}

	query, args, err := sqlx.In("SELECT id, name, assigned FROM tags WHERE name IN (?)", names)
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

	row := tx.QueryRowx("UPDATE tags SET name = ? WHERE id = ? RETURNING assigned", updateRequest.Name, id)
	var assigned bool
	if err := row.Scan(assigned); err != nil {
		return response, err
	}

	if err := tx.Commit(); err != nil {
		return response, err
	}

	return models.TagResponse{ID: id, Name: updateRequest.Name, Assigned: assigned}, nil
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

	if err := tx.Select(&response, "SELECT id, name, assigned FROM tags"); err != nil {
		return response, err
	}

	if err := tx.Commit(); err != nil {
		return response, err
	}

	return response, nil
}

func (repository *TagRepository) ListForDocument(tx *sqlx.Tx, documentID models.ID) (response []models.TagResponse, err error) {
	query := `
	SELECT id, name, assigned FROM tags
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

	// TODO: IN query to avoid loop
	for _, tag := range tags {
		_, err := tx.Exec("INSERT INTO tags_documents VALUES (?, ?)", tag.ID, documentID)
		if err != nil {
			return err
		}
		if err := repository.toggleTagAssigned(tx, tag.ID); err != nil {
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

	// TODO: IN query to avoid loop
	for _, tag := range tags {
		_, err := tx.Exec("DELETE FROM tags_documents WHERE tag = ? AND document = ?", tag.ID, documentID)
		if err != nil {
			return err
		}
		if err := repository.toggleTagAssigned(tx, tag.ID); err != nil {
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

/*
This method toggles boolean value in `assigned` column in `tags` table.
It is used internally to toggle tag status seamlessly when documents to which this tag is assigned
are attached or detached from it.
*/
func (repository *TagRepository) toggleTagAssigned(tx *sqlx.Tx, tagID models.ID) (err error) {
	query := `
	UPDATE tags SET 
	assigned = CASE
		           WHEN (SELECT COUNT(*) FROM tags_documents WHERE tag = ?) = 0
				       THEN 0
				   ELSE 1
			   END
	WHERE id = ?
`
	if tx == nil {
		tx, err = repository.db.Beginx()
		if err != nil {
			return ErrTransactionOpen
		}
		defer tx.Rollback()
	}

	if _, err := tx.Exec(query, tagID, tagID); err != nil {
		return err
	}

	if tx == nil {
		if err := tx.Commit(); err != nil {
			return err
		}
	}

	return err
}
