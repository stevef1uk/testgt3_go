package store

import (
	"context"
	"database/sql"
	"testing"
)

func TestStore_List(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	err = InitSchema(db)
	if err != nil {
		t.Fatal(err)
	}

	store := NewStore(db)

	// Create some links
	_, err = store.Create(context.Background(), "Link 1", "https://example.com")
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.Create(context.Background(), "Link 2", "https://example.com")
	if err != nil {
		t.Fatal(err)
	}

	links, err := store.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(links) != 2 {
		t.Errorf("expected 2 links, got %d", len(links))
	}

	if links[0].Title != "Link 2" {
		t.Errorf("expected first link to be 'Link 2', got '%s'", links[0].Title)
	}

	if links[1].Title != "Link 1" {
		t.Errorf("expected second link to be 'Link 1', got '%s'", links[1].Title)
	}
}

func TestStore_Create(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	err = InitSchema(db)
	if err != nil {
		t.Fatal(err)
	}

	store := NewStore(db)

	link, err := store.Create(context.Background(), "Link 1", "https://example.com")
	if err != nil {
		t.Fatal(err)
	}

	if link.Title != "Link 1" {
		t.Errorf("expected title to be 'Link 1', got '%s'", link.Title)
	}

	if link.URL != "https://example.com" {
		t.Errorf("expected URL to be 'https://example.com', got '%s'", link.URL)
	}

	if link.CreatedAt == "" {
		t.Errorf("expected created at to be set, got empty string")
	}
}

func TestStore_Delete(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	err = InitSchema(db)
	if err != nil {
		t.Fatal(err)
	}

	store := NewStore(db)

	_, err = store.Create(context.Background(), "Link 1", "https://example.com")
	if err != nil {
		t.Fatal(err)
	}

	err = store.Delete(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}

	links, err := store.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(links) != 0 {
		t.Errorf("expected 0 links, got %d", len(links))
	}
}
