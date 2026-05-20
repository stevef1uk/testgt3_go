package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type Bookmark struct {
	ID   int    `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
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

func NewSQLStore(db *sql.DB) *SQLStore {
	return &SQLStore{db: db}
}

func (s *SQLStore) ListBookmarks(ctx context.Context) ([]*Bookmark, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, title, url FROM bookmarks")
	if err != nil {
		return nil, fmt.Errorf("failed to query bookmarks: %w", err)
	}
	defer rows.Close()

	var bookmarks []*Bookmark
	for rows.Next() {
		var b Bookmark
		err := rows.Scan(&b.ID, &b.Title, &b.URL)
		if err != nil {
			return nil, fmt.Errorf("failed to scan bookmark row: %w", err)
		}
		bookmarks = append(bookmarks, &b)
	}

	return bookmarks, nil
}

func (s *SQLStore) GetBookmark(ctx context.Context, id int) (*Bookmark, error) {
	row := s.db.QueryRowContext(ctx, "SELECT id, title, url FROM bookmarks WHERE id =?", id)

	var b Bookmark
	err := row.Scan(&b.ID, &b.Title, &b.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to scan bookmark row: %w", err)
	}

	return &b, nil
}

func (s *SQLStore) CreateBookmark(ctx context.Context, bookmark *Bookmark) error {
	result, err := s.db.ExecContext(ctx, "INSERT INTO bookmarks (title, url) VALUES (?,?)", bookmark.Title, bookmark.URL)
	if err != nil {
		return fmt.Errorf("failed to insert bookmark: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last inserted ID: %w", err)
	}
	bookmark.ID = int(id)

	return nil
}

func (s *SQLStore) UpdateBookmark(ctx context.Context, id int, bookmark *Bookmark) error {
	result, err := s.db.ExecContext(ctx, "UPDATE bookmarks SET title =?, url =? WHERE id =?", bookmark.Title, bookmark.URL, id)
	if err != nil {
		return fmt.Errorf("failed to update bookmark: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("bookmark not found")
	}

	return nil
}

func (s *SQLStore) DeleteBookmark(ctx context.Context, id int) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM bookmarks WHERE id =?", id)
	if err != nil {
		return fmt.Errorf("failed to delete bookmark: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("bookmark not found")
	}

	return nil
}
