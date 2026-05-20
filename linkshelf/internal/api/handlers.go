package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"linkshelf/internal/store"
)

// bookmarkStore is the store for bookmarks.
var bookmarkStore store.Store

// SetStore sets the store for the API handlers.
func SetStore(s store.Store) {
	bookmarkStore = s
}

// ListBookmarksHandler handles GET /api/bookmarks.
func ListBookmarksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method!= http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	bookmarks, err := bookmarkStore.ListBookmarks(context.Background())
	if err!= nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bookmarks)
}

// BookmarkHandler handles GET, POST, PUT, DELETE on /api/bookmarks/{id}.
func BookmarkHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/api/bookmarks/"):]
	if idStr == "" {
		http.Error(w, "missing bookmark ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err!= nil {
		http.Error(w, "invalid bookmark ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		bookmark, err := bookmarkStore.GetBookmark(context.Background(), id)
		if err!= nil {
			if err == store.ErrRecordNotFound {
				http.Error(w, "bookmark not found", http.StatusNotFound)
			} else {
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(bookmark)
	case http.MethodPost:
		var b store.Bookmark
		if err := json.NewDecoder(r.Body).Decode(&b); err!= nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if err := bookmarkStore.CreateBookmark(context.Background(), &b); err!= nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(b)
	case http.MethodPut:
		var b store.Bookmark
		if err := json.NewDecoder(r.Body).Decode(&b); err!= nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		b.ID = id
		if err := bookmarkStore.UpdateBookmark(context.Background(), id, &b); err!= nil {
			if err == store.ErrRecordNotFound {
				http.Error(w, "bookmark not found", http.StatusNotFound)
			} else {
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(b)
	case http.MethodDelete:
		if err := bookmarkStore.DeleteBookmark(context.Background(), id); err!= nil {
			if err == store.ErrRecordNotFound {
				http.Error(w, "bookmark not found", http.StatusNotFound)
			} else {
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
