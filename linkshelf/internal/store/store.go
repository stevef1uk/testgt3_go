package store

import (
	"database/sql"
	"errors"
	"time"

	_ "modernc.org/sqlite"
)

// Bookmark represents a saved link with metadata.
type Bookmark struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
}

// Store encapsulates database operations for bookmarks.
type Store struct {
	db *sql.DB
}

// NewStore returns a new Store using the provided database connection.
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// InitSchema ensures the bookmarks table exists.
func (s *Store) InitSchema() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS bookmarks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			url TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

// OpenDB opens the SQLite database and returns the connection.
func OpenDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "./linkshelf.db")
	if err != nil {
		return nil, err
	}
	// Ensure schema is initialized
	if err := NewStore(db).InitSchema(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

// GetAllBookmarks retrieves all bookmarks from the database.
func (s *Store) GetAllBookmarks() ([]Bookmark, error) {
	rows, err := s.db.Query("SELECT id, title, url, created_at FROM bookmarks")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookmarks []Bookmark
	for rows.Next() {
		var b Bookmark
		err := rows.Scan(&b.ID, &b.Title, &b.URL, &b.CreatedAt)
		if err != nil {
			return nil, err
		}
		bookmarks = append(bookmarks, b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return bookmarks, nil
}

// GetBookmark retrieves a single bookmark by ID.
func (s *Store) GetBookmark(id int) (*Bookmark, error) {
	row := s.db.QueryRow("SELECT id, title, url, created_at FROM bookmarks WHERE id = ?", id)
	var b Bookmark
	err := row.Scan(&b.ID, &b.Title, &b.URL, &b.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("bookmark not found")
		}
		return nil, err
	}
	return &b, nil
}

// CreateBookmark inserts a new bookmark into the database.
func (s *Store) CreateBookmark(b *Bookmark) error {
	res, err := s.db.Exec("INSERT INTO bookmarks (title, url, created_at) VALUES (?, ?, ?)", b.Title, b.URL, time.Now())
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	b.ID = int(id)
	return nil
}

// UpdateBookmark modifies an existing bookmark.
func (s *Store) UpdateBookmark(b *Bookmark) error {
	_, err := s.db.Exec("UPDATE bookmarks SET title = ?, url = ? WHERE id = ?", b.Title, b.URL, b.ID)
	return err
}

// DeleteBookmark removes a bookmark by ID.
func (s *Store) DeleteBookmark(id int) error {
	_, err := s.db.Exec("DELETE FROM bookmarks WHERE id = ?", id)
	return err
}
