package store

import (
	"context"
	"database/sql"
)

type Store struct {
	db *sql.DB
}

type Bookmark struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	Url      string `json:"url"`
	Notes    string `json:"notes"`
	Username string `json:"username"`
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetAllBookmarks(ctx context.Context) ([]Bookmark, error) {
	// implement GetAllBookmarks
	return []Bookmark{}, nil
}

func (s *Store) GetBookmark(ctx context.Context, id int64) (*Bookmark, error) {
	// implement GetBookmark
	return &Bookmark{}, nil
}

func (s *Store) CreateBookmark(ctx context.Context, b *Bookmark) error {
	// implement CreateBookmark
	return nil
}

func (s *Store) UpdateBookmark(ctx context.Context, b *Bookmark) error {
	// implement UpdateBookmark
	return nil
}

func (s *Store) DeleteBookmark(ctx context.Context, id int64) error {
	// implement DeleteBookmark
	return nil
}

func Initialize() {
	// initialization code here
}
