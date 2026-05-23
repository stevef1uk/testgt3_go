package store

import (
	"context"
	"strings"
	"testing"
)

// helper creates a fresh in‑memory store with schema initialized.
func newTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := NewStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	// Ensure the schema is present. NewStore may already do this,
	// but calling InitSchema explicitly is safe and idempotent.
	if err := InitSchema(s.db); err != nil {
		t.Fatalf("failed to init schema: %v", err)
	}
	return s
}

// Test that List on a brand‑new store returns an empty slice (not nil).
func TestStore_List_Empty(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	links, err := s.List(ctx)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if links == nil {
		t.Fatalf("List returned nil slice, want empty slice")
	}
	if len(links) != 0 {
		t.Fatalf("List returned %d links, want 0", len(links))
	}
}

// Test successful creation of a link and that it appears in List.
func TestStore_Create_List_Success(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	const (
		title = "OpenAI"
		url   = "https://openai.com"
	)

	created, err := s.Create(ctx, title, url)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if created.Title != title || created.URL != url {
		t.Fatalf("Created link fields mismatch: got %+v", created)
	}
	if created.ID == 0 {
		t.Fatalf("Created link has zero ID")
	}
	if created.CreatedAt == "" {
		t.Fatalf("Created link missing CreatedAt timestamp")
	}

	// Verify List now returns exactly this link.
	links, err := s.List(ctx)
	if err != nil {
		t.Fatalf("List after Create returned error: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("List after Create returned %d links, want 1", len(links))
	}
	got := links[0]
	if got.ID != created.ID || got.Title != title || got.URL != url {
		t.Fatalf("List returned unexpected link: got %+v, want %+v", got, created)
	}
}

// Test validation: empty title should be rejected.
func TestStore_Create_EmptyTitle(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	_, err := s.Create(ctx, "", "https://example.com")
	if err == nil {
		t.Fatalf("Create with empty title succeeded, want error")
	}
}

// Test validation: title longer than 200 characters should be rejected.
func TestStore_Create_TitleTooLong(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	longTitle := strings.Repeat("x", 201) // 201 > 200
	_, err := s.Create(ctx, longTitle, "https://example.com")
	if err == nil {
		t.Fatalf("Create with too‑long title succeeded, want error")
	}
}

// Test validation: URL must start with http:// or https://.
func TestStore_Create_InvalidURL(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	_, err := s.Create(ctx, "Bad URL", "ftp://example.com")
	if err == nil {
		t.Fatalf("Create with invalid URL scheme succeeded, want error")
	}
}

// Test successful deletion of an existing link.
func TestStore_Delete_Success(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	created, err := s.Create(ctx, "To Delete", "https://example.com")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if err := s.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	// After deletion, List should be empty.
	links, err := s.List(ctx)
	if err != nil {
		t.Fatalf("List after Delete returned error: %v", err)
	}
	if len(links) != 0 {
		t.Fatalf("List after Delete returned %d links, want 0", len(links))
	}
}

// Test deletion of a non‑existent ID returns an error.
func TestStore_Delete_NonExistentID(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	if err := s.Delete(ctx, 9999); err == nil {
		t.Fatalf("Delete of non‑existent ID succeeded, want error")
	}
}