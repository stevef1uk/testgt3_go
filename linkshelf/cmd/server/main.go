package main

import (
	"database/sql"
	"linkshelf/internal/api"
	"linkshelf/internal/store"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Open a SQLite database connection to linkshelf.db
	db, err := sql.Open("sqlite3", "./linkshelf.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Ensure the schema is applied
	err = store.InitSchema(db)
	if err != nil {
		log.Fatal(err)
	}

	// Create a store instance with the database handle
	storeInstance := store.NewStore(db)

	// Create API handlers with store dependency
	handler := api.NewHandler(storeInstance)

	// Register all HTTP routes on http.DefaultServeMux
	http.Handle("/", handler)

	// Start listening on :8080
	log.Println("listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
