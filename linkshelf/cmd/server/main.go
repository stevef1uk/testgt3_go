package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
	"linkshelf/internal/api"
	"linkshelf/internal/store"
)

func main() {
	// Initialize store
	db, err := sql.Open("sqlite3", "./linkshelf.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	s, err := store.NewSQLStore(db)
	if err != nil {
		log.Fatal(err)
	}
	api.SetStore(s)

	// Serve static files from web directory
	fs := http.FileServer(http.Dir("./web"))
	http.Handle("/", fs)

	// API routes
	http.HandleFunc("/api/bookmarks", api.ListBookmarksHandler)
	http.HandleFunc("/api/bookmarks/", api.BookmarkHandler)

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
