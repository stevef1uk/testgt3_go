package store

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestInitSchema(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	if err := InitSchema(db); err != nil {
		t.Fatalf("InitSchema failed: %v", err)
	}

	// Verify table exists by querying it
	row := db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='links'`)
	var name string
	if err := row.Scan(&name); err != nil {
		t.Fatalf("links table not found: %v", err)
	}
	if name != "links" {
		t.Fatalf("expected table name 'links', got %q", name)
	}
}
