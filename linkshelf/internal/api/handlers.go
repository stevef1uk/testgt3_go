package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"linkshelf/internal/store"
)

// Handlers encapsulates HTTP handlers for CRUD operations
type Handlers struct {
	store store.Store
}

// NewHandlers returns a new Handlers instance
func NewHandlers(store store.Store) *Handlers {
	return &Handlers{store: store}
}

// getBookmarksHandler handles GET /api/bookmarks
func (h *Handlers) getBookmarksHandler(w http.ResponseWriter, r *http.Request) {
	bookmarks, err := h.store.GetAllBookmarks()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(bookmarks)
}

// createBookmarkHandler handles POST /api/bookmarks
func (h *Handlers) createBookmarkHandler(w http.ResponseWriter, r *http.Request) {
	var b store.Bookmark
	err := json.NewDecoder(r.Body).Decode(&b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = h.store.CreateBookmark(&b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// getBookmarkHandler handles GET /api/bookmarks/:id
func (h *Handlers) getBookmarkHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Path[len("/api/bookmarks/"):])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	b, err := h.store.GetBookmark(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(b)
}

// updateBookmarkHandler handles PUT /api/bookmarks/:id
func (h *Handlers) updateBookmarkHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Path[len("/api/bookmarks/"):])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var b store.Bookmark
	err = json.NewDecoder(r.Body).Decode(&b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	b.ID = id
	err = h.store.UpdateBookmark(&b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// deleteBookmarkHandler handles DELETE /api/bookmarks/:id
func (h *Handlers) deleteBookmarkHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Path[len("/api/bookmarks/"):])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = h.store.DeleteBookmark(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
