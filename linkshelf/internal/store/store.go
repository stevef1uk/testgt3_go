package store

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

// Bookmark represents a bookmark with ID, Title, URL, and CreatedAt timestamp.
type Bookmark struct {
	ID        int
	Title     string
	URL       string
	CreatedAt time.Time
}

// Store provides an abstraction over SQLite operations for bookmarks.
type Store struct {
	db *sql.DB
}

// NewStore returns a new Store instance with the given database connection.
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// GetAllBookmarks retrieves all bookmarks from the database.
func (s *Store) GetAllBookmarks() ([]*Bookmark, error) {
	rows, err := s.db.Query("SELECT id, title, url, created_at FROM bookmarks")
	if err!= nil {
		return nil, err
	}
	defer rows.Close()

	var bookmarks []*Bookmark
	for rows.Next() {
		var b Bookmark
		err := rows.Scan(&b.ID, &b.Title, &b.URL, &b.CreatedAt)
		if err!= nil {
			return nil, err
		}
		bookmarks = append(bookmarks, &b)
	}
	return bookmarks, nil
}

// GetBookmark retrieves a bookmark by ID from the database.
func (s *Store) GetBookmark(id int) (*Bookmark, error) {
	row := s.db.QueryRow("SELECT id, title, url, created_at FROM bookmarks WHERE id =?", id)
	var b Bookmark
	err := row.Scan(&b.ID, &b.Title, &b.URL, &b.CreatedAt)
	if err!= nil {
		return nil, err
	}
	return &b, nil
}

// CreateBookmark creates a new bookmark in the database.
func (s *Store) CreateBookmark(b *Bookmark) error {
	_, err := s.db.Exec("INSERT INTO bookmarks (title, url, created_at) VALUES (?,?,?)", b.Title, b.URL, time.Now())
	return err
}

// UpdateBookmark updates an existing bookmark in the database.
func (s *Store) UpdateBookmark(b *Bookmark) error {
	_, err := s.db.Exec("UPDATE bookmarks SET title =?, url =? WHERE id =?", b.Title, b.URL, b.ID)
	return err
}

// DeleteBookmark deletes a bookmark by ID from the database.
func (s *Store) DeleteBookmark(id int) error {
	_, err := s.db.Exec("DELETE FROM bookmarks WHERE id =?", id)
	return err
}

// OpenDatabase opens the SQLite database and ensures schema initialization.
func OpenDatabase() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "./linkshelf.db")
	if err!= nil {
		return nil, err
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS bookmarks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			url TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err!= nil {
		return nil, err
	}
	return db, nil
}
