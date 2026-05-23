package store

import (
	"database/sql"
)

// InitSchema creates the links table if it does not exist.
func InitSchema(db *sql.DB) error {
	const schema = `
		CREATE TABLE IF NOT EXISTS links (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			url   TEXT NOT NULL,
			created_at TEXT NOT NULL
		);
	`
	_, err := db.Exec(schema)
	return err
}
