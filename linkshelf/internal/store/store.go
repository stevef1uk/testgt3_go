package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Bookmark represents a bookmark in the linkshelf application.
type Bookmark struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
}

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetAllBookmarks(ctx context.Context) ([]*Bookmark, error) {
	query := `SELECT id, title, url, created_at FROM bookmarks`
	rows, err := s.db.QueryContext(ctx, query)
	if err!= nil {
		return nil, fmt.Errorf("failed to get all bookmarks: %w", err)
	}
	defer rows.Close()

	var bookmarks []*Bookmark
	for rows.Next() {
		var bookmark Bookmark
		err := rows.Scan(&bookmark.ID, &bookmark.Title, &bookmark.URL, &bookmark.CreatedAt)
		if err!= nil {
			return nil, fmt.Errorf("failed to scan bookmark: %w", err)
		}
		bookmarks = append(bookmarks, &bookmark)
	}

	if err := rows.Err(); err!= nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return bookmarks, nil
}

func (s *Store) GetBookmark(ctx context.Context, id int64) (*Bookmark, error) {
	query := `SELECT id, title, url, created_at FROM bookmarks WHERE id = $1`
	bookmark := &Bookmark{}
	err := s.db.QueryRowContext(ctx, query, id).Scan(&bookmark.ID, &bookmark.Title, &bookmark.URL, &bookmark.CreatedAt)
	if err!= nil {
		return nil, fmt.Errorf("failed to get bookmark by id: %w", err)
	}
	return bookmark, nil
}

func (s *Store) CreateBookmark(ctx context.Context, bookmark *Bookmark) error {
	query := `INSERT INTO bookmarks (title, url, created_at) VALUES ($1, $2, $3) RETURNING id`
	err := s.db.QueryRowContext(ctx, query, bookmark.Title, bookmark.URL, time.Now()).Scan(&bookmark.ID)
	if err!= nil {
		return fmt.Errorf("failed to create bookmark: %w", err)
	}
	return nil
}

func (s *Store) UpdateBookmark(ctx context.Context, bookmark *Bookmark) error {
	query := `UPDATE bookmarks SET title = $1, url = $2 WHERE id = $3`
	result, err := s.db.ExecContext(ctx, query, bookmark.Title, bookmark.URL, bookmark.ID)
	if err!= nil {
		return fmt.Errorf("failed to update bookmark: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err!= nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("bookmark not found")
	}
	return nil
}

func (s *Store) DeleteBookmark(ctx context.Context, id int64) error {
	query := `DELETE FROM bookmarks WHERE id = $1`
	result, err := s.db.ExecContext(ctx, query, id)
	if err!= nil {
		return fmt.Errorf("failed to delete bookmark: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err!= nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("bookmark not found")
	}
	return nil
}

