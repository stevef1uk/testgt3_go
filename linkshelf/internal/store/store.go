package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Bookmark struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
}

type Store interface {
	ListBookmarks(ctx context.Context) ([]*Bookmark, error)
	GetBookmark(ctx context.Context, id int) (*Bookmark, error)
	CreateBookmark(ctx context.Context, bookmark *Bookmark) error
	UpdateBookmark(ctx context.Context, id int, bookmark *Bookmark) error
	DeleteBookmark(ctx context.Context, id int) error
}

var ErrRecordNotFound = errors.New("record not found")

type SQLStore struct {
	db *sql.DB
}

// NewSQLStore creates a new SQLStore and ensures the database schema exists.
func NewSQLStore(db *sql.DB) (*SQLStore, error) {
	s := &SQLStore{db: db}
	if err := s.ensureSchema(); err != nil {
		return nil, fmt.Errorf("failed to ensure schema: %w", err)
	}
	return s, nil
}

// ensureSchema runs the DDL to create the bookmarks table if it does not exist.
func (s *SQLStore) ensureSchema() error {
	const schema = `
	CREATE TABLE IF NOT EXISTS bookmarks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		url TEXT NOT NULL,
		created_at DATETIME NOT NULL
	);
	`
	_, err := s.db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to execute schema DDL: %w", err)
	}
	return nil
}

// ListBookmarks returns all bookmarks ordered by creation time descending.
func (s *SQLStore) ListBookmarks(ctx context.Context) ([]*Bookmark, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, title, url, created_at FROM bookmarks ORDER BY created_at DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to query bookmarks: %w", err)
	}
	defer rows.Close()

	var bookmarks []*Bookmark
	for rows.Next() {
		var b Bookmark
		if err := rows.Scan(&b.ID, &b.Title, &b.URL, &b.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan bookmark row: %w", err)
		}
		bookmarks = append(bookmarks, &b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}
	return bookmarks, nil
}

// GetBookmark retrieves a single bookmark by id.
func (s *SQLStore) GetBookmark(ctx context.Context, id int) (*Bookmark, error) {
	row := s.db.QueryRowContext(ctx, "SELECT id, title, url, created_at FROM bookmarks WHERE id = ?", id)
	var b Bookmark
	if err := row.Scan(&b.ID, &b.Title, &b.URL, &b.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, fmt.Errorf("failed to scan bookmark row: %w", err)
	}
	return &b, nil
}

// CreateBookmark inserts a new bookmark and sets its ID and CreatedAt.
func (s *SQLStore) CreateBookmark(ctx context.Context, bookmark *Bookmark) error {
	if bookmark == nil {
		return fmt.Errorf("bookmark is nil")
	}
	const stmt = `INSERT INTO bookmarks (title, url, created_at) VALUES (?, ?, ?)`
	res, err := s.db.ExecContext(ctx, stmt, bookmark.Title, bookmark.URL, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("failed to insert bookmark: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to retrieve last insert id: %w", err)
	}
	bookmark.ID = int(id)
	bookmark.CreatedAt = time.Now().UTC()
	return nil
}

// UpdateBookmark updates the title and url of an existing bookmark.
func (s *SQLStore) UpdateBookmark(ctx context.Context, id int, bookmark *Bookmark) error {
	if bookmark == nil {
		return fmt.Errorf("bookmark is nil")
	}
	const stmt = `UPDATE bookmarks SET title = ?, url = ? WHERE id = ?`
	res, err := s.db.ExecContext(ctx, stmt, bookmark.Title, bookmark.URL, id)
	if err != nil {
		return fmt.Errorf("failed to update bookmark: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if affected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

// DeleteBookmark removes a bookmark by id.
func (s *SQLStore) DeleteBookmark(ctx context.Context, id int) error {
	const stmt = `DELETE FROM bookmarks WHERE id = ?`
	res, err := s.db.ExecContext(ctx, stmt, id)
	if err != nil {
		return fmt.Errorf("failed to delete bookmark: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if affected == 0 {
		return ErrRecordNotFound
	}
	return nil
}
