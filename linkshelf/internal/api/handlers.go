package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"linkshelf/internal/store"
)

var bookmarkStore *store.Store

// SetStore injects the store dependency into the API package.
func SetStore(s *store.Store) {
	bookmarkStore = s
}

// ListBookmarksHandler returns a list of all bookmarks.
func ListBookmarksHandler(w http.ResponseWriter, r *http.Request) {
	bookmarks, err := bookmarkStore.ListBookmarks(r.Context())
	if err != nil {
		http.Error(w, "failed to list bookmarks", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bookmarks)
}

// CreateBookmarkHandler creates a new bookmark.
func CreateBookmarkHandler(w http.ResponseWriter, r *http.Request) {
	var bookmark store.Bookmark
	if err := json.NewDecoder(r.Body).Decode(&bookmark); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}
	if err := bookmarkStore.CreateBookmark(r.Context(), &bookmark); err != nil {
		http.Error(w, "failed to create bookmark", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(bookmark)
}

// GetBookmarkHandler returns a single bookmark by ID.
func GetBookmarkHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid bookmark ID", http.StatusBadRequest)
		return
	}
	bookmark, err := bookmarkStore.GetBookmark(r.Context(), id)
	if err != nil {
		if err == store.ErrNotFound {
			http.Error(w, "bookmark not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get bookmark", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bookmark)
}

// UpdateBookmarkHandler updates an existing bookmark.
func UpdateBookmarkHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid bookmark ID", http.StatusBadRequest)
		return
	}
	var bookmark store.Bookmark
	if err := json.NewDecoder(r.Body).Decode(&bookmark); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}
	bookmark.ID = id
	if err := bookmarkStore.UpdateBookmark(r.Context(), &bookmark); err != nil {
		if err == store.ErrNotFound {
			http.Error(w, "bookmark not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to update bookmark", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bookmark)
}

// DeleteBookmarkHandler deletes a bookmark by ID.
func DeleteBookmarkHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid bookmark ID", http.StatusBadRequest)
		return
	}
	if err := bookmarkStore.DeleteBookmark(r.Context(), id); err != nil {
		if err == store.ErrNotFound {
			http.Error(w, "bookmark not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to delete bookmark", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
