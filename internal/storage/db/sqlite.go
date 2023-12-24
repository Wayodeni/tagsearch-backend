package db

import (
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func NewDb(path string) *sqlx.DB {
	db, err := sqlx.Open("sqlite", "file:"+path+"?"+"_foreign_keys=1")
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
		FOREIGN KEY(tag) REFERENCES tags(id) ON DELETE CASCADE,
		FOREIGN KEY(document) REFERENCES documents(id) ON DELETE CASCADE
	)
	`)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`
	CREATE UNIQUE INDEX IF NOT EXISTS "DOCUMENT_ID" ON "documents" (
		"id"
	)
	`)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`
	CREATE INDEX IF NOT EXISTS "DOCUMENT_ID_M2M" ON "tags_documents" (
		"document"
	)
	`)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`
	CREATE UNIQUE INDEX IF NOT EXISTS "TAG_ID" ON "tags" (
		"id",
		"name"
	)
	`)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`
	CREATE INDEX IF NOT EXISTS "TAG_ID_M2M" ON "tags_documents" (
		"tag"
	)
	`)
	if err != nil {
		panic(err)
	}

	return db
}
