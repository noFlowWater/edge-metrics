package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// OpenDB opens a SQLite database connection at the given path.
func OpenDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	log.Println("Database connection opened successfully")
	return db, nil
}
