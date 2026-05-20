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
)

func main() {
	// Determine the port to listen on.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Set up HTTP routes.
	mux := http.NewServeMux()
	mux.HandleFunc("/api/bookmarks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			api.ListBookmarksHandler(w, r)
		case http.MethodPost:
			api.CreateBookmarkHandler(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/bookmarks/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			api.GetBookmarkHandler(w, r)
		case http.MethodPut, http.MethodPatch:
			api.UpdateBookmarkHandler(w, r)
		case http.MethodDelete:
			api.DeleteBookmarkHandler(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Create the server.
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// Run the server in a goroutine so we can handle shutdown signals.
	go func() {
		log.Printf("Server listening on %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced shutdown: %v", err)
	}

	log.Println("Server stopped")
}
