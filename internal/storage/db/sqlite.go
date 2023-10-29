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
		name TEXT NOT NULL
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

	return db
}
