package store

import (
	"context"
	"database/sql"
	"time"
)

// Bookmark represents a bookmark in the system.
type Bookmark struct {
	ID        int64
	Title     string
	URL       string
	CreatedAt time.Time
}

// Store provides methods to interact with the bookmarks table.
type Store struct {
	db *sql.DB
}

// NewStore returns a new Store given a database connection.
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// GetAllBookmarks returns all bookmarks from the database.
func (s *Store) GetAllBookmarks(ctx context.Context) ([]Bookmark, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, title, url, created_at FROM bookmarks")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookmarks []Bookmark
	for rows.Next() {
		var b Bookmark
		if err := rows.Scan(&b.ID, &b.Title, &b.URL, &b.CreatedAt); err != nil {
			return nil, err
		}
		bookmarks = append(bookmarks, b)
	}
	return bookmarks, nil
}

// GetBookmark returns a single bookmark by ID.
func (s *Store) GetBookmark(ctx context.Context, id int64) (*Bookmark, error) {
	row := s.db.QueryRowContext(ctx, "SELECT id, title, url, created_at FROM bookmarks WHERE id = ?", id)
	var b Bookmark
	if err := row.Scan(&b.ID, &b.Title, &b.URL, &b.CreatedAt); err != nil {
		return nil, err
	}
	return &b, nil
}

// CreateBookmark inserts a new bookmark and returns its ID.
func (s *Store) CreateBookmark(ctx context.Context, b *Bookmark) (int64, error) {
	result, err := s.db.ExecContext(
		ctx,
		"INSERT INTO bookmarks (title, url, created_at) VALUES (?, ?, ?)",
		b.Title, b.URL, b.CreatedAt,
	)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// UpdateBookmark updates an existing bookmark.
func (s *Store) UpdateBookmark(ctx context.Context, b *Bookmark) error {
	_, err := s.db.ExecContext(
		ctx,
		"UPDATE bookmarks SET title = ?, url = ?, created_at = ? WHERE id = ?",
		b.Title, b.URL, b.CreatedAt, b.ID,
	)
	return err
}

// DeleteBookmark removes a bookmark by ID.
func (s *Store) DeleteBookmark(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM bookmarks WHERE id = ?", id)
	return err
}
