package store

import (
	"database/sql"
	"errors"
	"time"

	_ "modernc.org/sqlite"
)

type Bookmark struct {
	ID        int
	Title     string
	URL       string
	CreatedAt time.Time
}

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

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

func (s *Store) CreateBookmark(b *Bookmark) error {
	res, err := s.db.Exec("INSERT INTO bookmarks (title, url, created_at) VALUES (?,?,?)", b.Title, b.URL, b.CreatedAt)
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

func (s *Store) UpdateBookmark(b *Bookmark) error {
	res, err := s.db.Exec("UPDATE bookmarks SET title = ?, url = ?, created_at = ? WHERE id = ?", b.Title, b.URL, b.CreatedAt, b.ID)
	if err != nil {
		return err
	}
	_, err = res.RowsAffected()
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) DeleteBookmark(id int) error {
	res, err := s.db.Exec("DELETE FROM bookmarks WHERE id = ?", id)
	if err != nil {
		return err
	}
	_, err = res.RowsAffected()
	if err != nil {
		return err
	}
	return nil
}
