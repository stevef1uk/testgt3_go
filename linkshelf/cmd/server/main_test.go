package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"linkshelf/internal/api"
	"linkshelf/internal/store"

	_ "github.com/mattn/go-sqlite3"
)

func TestMainHandler(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	err = store.InitSchema(db)
	if err != nil {
		t.Fatal(err)
	}

	storeInstance := store.NewStore(db)

	// Create some links
	_, err = storeInstance.Create(context.Background(), "Link 1", "https://example.com")
	if err != nil {
		t.Fatal(err)
	}
	_, err = storeInstance.Create(context.Background(), "Link 2", "https://example.com")
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("GET", "/api/links", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create the API handler directly (main.go registers it on "/").
	h := api.NewHandler(storeInstance)

	// Serve the request using the handler.
	// Create a response recorder.
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response []store.Link
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatal(err)
	}

	if len(response) != 2 {
		t.Errorf("expected 2 links, got %d", len(response))
	}

	// Check that the links are returned in descending order of ID (which is creation order here)
	if response[0].Title != "Link 2" {
		t.Errorf("expected first link title to be 'Link 2', got '%s'", response[0].Title)
	}
	if response[1].Title != "Link 1" {
		t.Errorf("expected second link title to be 'Link 1', got '%s'", response[1].Title)
	}
}
