package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"linkshelf/internal/store"
)

type Handlers struct {
	store store.Store
}

func NewHandlers(store store.Store) *Handlers {
	return &Handlers{store: store}
}

func (h *Handlers) GetBookmarksHandler(w http.ResponseWriter, r *http.Request) {
	bookmarks, err := h.store.GetAllBookmarks(r.Context())
	if err!= nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(bookmarks)
}

func (h *Handlers) CreateBookmarkHandler(w http.ResponseWriter, r *http.Request) {
	var b store.Bookmark
	err := json.NewDecoder(r.Body).Decode(&b)
	if err!= nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = h.store.CreateBookmark(r.Context(), &b)
	if err!= nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *Handlers) GetBookmarkHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Path[len("/api/bookmarks/"):])
	if err!= nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	b, err := h.store.GetBookmark(r.Context(), id)
	if err!= nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(b)
}

func (h *Handlers) UpdateBookmarkHandler(w http.ResponseWriter, r *http.Request) {
	var b store.Bookmark
	err := json.NewDecoder(r.Body).Decode(&b)
	if err!= nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = h.store.UpdateBookmark(r.Context(), &b)
	if err!= nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handlers) DeleteBookmarkHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Path[len("/api/bookmarks/"):])
	if err!= nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = h.store.DeleteBookmark(r.Context(), id)
	if err!= nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
