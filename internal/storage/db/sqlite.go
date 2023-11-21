package db

import (
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func NewDb(path string) *sqlx.DB {
	db, err := sqlx.Open("sqlite", path)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS tags (
		id   INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		UNIQUE(name)
	)
	`)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS documents (
		id   INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		body TEXT NOT NULL,
		UNIQUE(name)
	)
	`)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS tags_documents (
		tag INTEGER NOT NULL,
		document INTEGER NOT NULL,
		FOREIGN KEY(tag) REFERENCES tags(id) ON DELETE CASCADE
		FOREIGN KEY(document) REFERENCES documents(id) ON DELETE CASCADE
	)
	`)
	if err != nil {
		panic(err)
	}

	return db
}
