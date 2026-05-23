package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

type Link struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	CreatedAt string `json:"created_at"` // RFC3339
}

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) List(ctx context.Context) ([]Link, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, title, url, created_at FROM links ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []Link
	for rows.Next() {
		var link Link
		err := rows.Scan(&link.ID, &link.Title, &link.URL, &link.CreatedAt)
		if err != nil {
			return nil, err
		}
		links = append(links, link)
	}

	return links, nil
}

func (s *Store) Create(ctx context.Context, title, url string) (Link, error) {
	// Validation
	if strings.TrimSpace(title) == "" {
		return Link{}, errors.New("title must be non‑empty")
	}
	if len(title) > 200 {
		return Link{}, errors.New("title exceeds 200 characters")
	}
	if strings.TrimSpace(url) == "" {
		return Link{}, errors.New("url must be non‑empty")
	}
	if !(strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")) {
		return Link{}, errors.New("url must start with http:// or https://")
	}

	// Insert row
	result, err := s.db.ExecContext(ctx,
		"INSERT INTO links (title, url, created_at) VALUES (?, ?, datetime('now'))",
		title, url,
	)
	if err != nil {
		return Link{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Link{}, err
	}
	// Fetch the exact stored timestamp to keep consistency with schema
	var createdAt string
	err = s.db.QueryRowContext(ctx,
		"SELECT created_at FROM links WHERE id = ?",
		id,
	).Scan(&createdAt)
	if err != nil {
		return Link{}, err
	}
	return Link{
		ID:        id,
		Title:     title,
		URL:       url,
		CreatedAt: createdAt,
	}, nil
}

func (s *Store) Delete(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM links WHERE id = ?", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
