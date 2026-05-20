package store

import (
	"context"
	"database/sql"
	"errors"
)

var ErrNotFound = errors.New("bookmark not found")

// Bookmark represents a saved URL with title and metadata.
type Bookmark struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

// Store manages database operations for bookmarks.
type Store struct {
	db *sql.DB
}

// NewStore creates a new store with the given database connection.
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// OpenDB opens a connection to the SQLite database.
func OpenDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "bookmarks.db")
	if err != nil {
		return nil, err
	}
	// Verify the connection.
	if err := db.Ping(); err != nil {
		return nil, err
	}
	// Create the bookmarks table if it doesn't exist.
	if err := createTable(db); err != nil {
		return nil, err
	}
	return db, nil
}

// createTable ensures the bookmarks table exists.
func createTable(db *sql.DB) error {
	stmt := `
	CREATE TABLE IF NOT EXISTS bookmarks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		url TEXT NOT NULL,
		description TEXT
	);`
	_, err := db.Exec(stmt)
	return err
}

// ListBookmarks returns all stored bookmarks.
func (s *Store) ListBookmarks(ctx context.Context) ([]Bookmark, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, title, url, description FROM bookmarks")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookmarks []Bookmark
	for rows.Next() {
		var b Bookmark
		if err := rows.Scan(&b.ID, &b.Title, &b.URL, &b.Description); err != nil {
			return nil, err
		}
		bookmarks = append(bookmarks, b)
	}
	return bookmarks, nil
}

// CreateBookmark inserts a new bookmark into the database.
func (s *Store) CreateBookmark(ctx context.Context, b *Bookmark) error {
	err := s.db.QueryRowContext(ctx, "INSERT INTO bookmarks (title, url, description) VALUES (?, ?, ?) RETURNING id",
		b.Title, b.URL, b.Description).Scan(&b.ID)
	return err
}

// GetBookmark retrieves a bookmark by ID.
func (s *Store) GetBookmark(ctx context.Context, id int) (*Bookmark, error) {
	var b Bookmark
	err := s.db.QueryRowContext(ctx, "SELECT id, title, url, description FROM bookmarks WHERE id = ?", id).
		Scan(&b.ID, &b.Title, &b.URL, &b.Description)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &b, nil
}

// UpdateBookmark modifies an existing bookmark.
func (s *Store) UpdateBookmark(ctx context.Context, b *Bookmark) error {
	result, err := s.db.ExecContext(ctx, "UPDATE bookmarks SET title = ?, url = ?, description = ? WHERE id = ?",
		b.Title, b.URL, b.Description, b.ID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteBookmark removes a bookmark by ID.
func (s *Store) DeleteBookmark(ctx context.Context, id int) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM bookmarks WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
