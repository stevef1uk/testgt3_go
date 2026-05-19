package store

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Bookmark struct {
	ID        int64
	Title     string
	URL       string
	CreatedAt time.Time
}

func Initialize() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "linkshelf.db")
	if err!= nil {
		return nil, err
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS bookmarks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		url TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`)
	if err!= nil {
		return nil, err
	}

	return db, nil
}

func GetAllBookmarks(db *sql.DB) ([]Bookmark, error) {
	rows, err := db.Query("SELECT id, title, url, created_at FROM bookmarks")
	if err!= nil {
		return nil, err
	}
	defer rows.Close()

	var bookmarks []Bookmark
	for rows.Next() {
		var b Bookmark
		err := rows.Scan(&b.ID, &b.Title, &b.URL, &b.CreatedAt)
		if err!= nil {
			return nil, err
		}
		bookmarks = append(bookmarks, b)
	}

	return bookmarks, nil
}

func GetBookmark(db *sql.DB, id int64) (*Bookmark, error) {
	row := db.QueryRow("SELECT id, title, url, created_at FROM bookmarks WHERE id =?", id)

	var b Bookmark
	err := row.Scan(&b.ID, &b.Title, &b.URL, &b.CreatedAt)
	if err!= nil {
		return nil, err
	}

	return &b, nil
}

func CreateBookmark(db *sql.DB, title, url string) (*Bookmark, error) {
	result, err := db.Exec("INSERT INTO bookmarks (title, url) VALUES (?,?)", title, url)
	if err!= nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err!= nil {
		return nil, err
	}

	return GetBookmark(db, id)
}

func UpdateBookmark(db *sql.DB, b *Bookmark) error {
	_, err := db.Exec("UPDATE bookmarks SET title =?, url =? WHERE id =?", b.Title, b.URL, b.ID)
	return err
}

func DeleteBookmark(db *sql.DB, id int64) error {
	_, err := db.Exec("DELETE FROM bookmarks WHERE id =?", id)
	return err
}
