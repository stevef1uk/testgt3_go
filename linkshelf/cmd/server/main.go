package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"linkshelf/internal/api"
	"linkshelf/internal/store"
	"database/sql"
	_ "modernc.org/sqlite"
)

func main() {
	// Get port from environment or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	db, err := sql.Open("sqlite", "linkshelf.db")
	if err!= nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	store := store.NewStore(db)
	h := api.NewHandlers(*store)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/bookmarks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.GetBookmarksHandler(w, r)
		case http.MethodPost:
			h.CreateBookmarkHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/bookmarks/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.GetBookmarkHandler(w, r)
		case http.MethodPut:
			h.UpdateBookmarkHandler(w, r)
		case http.MethodDelete:
			h.DeleteBookmarkHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		log.Printf("Server starting on port %s", port)
		if err := srv.ListenAndServe(); err!= nil && err!= http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err!= nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	log.Println("Server shutdown complete")
}
