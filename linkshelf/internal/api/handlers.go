package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"linkshelf/internal/store"
)

type handler struct {
	store *store.Store
}

func NewHandler(store *store.Store) *handler {
	return &handler{store: store}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Root page
	if r.URL.Path == "/" {
		http.ServeFile(w, r, "index.html")
		return
	}

	// Static assets: /static/{file}
	if strings.HasPrefix(r.URL.Path, "/static/") {
		// Prevent directory traversal
		if strings.Contains(r.URL.Path, "..") {
			http.Error(w, "invalid path", http.StatusForbidden)
			return
		}
		// Serve the file relative to the current working directory.
		// The request path includes the "/static/" prefix; remove the leading slash
		// to obtain a filesystem path like "static/test.txt".
		filePath := strings.TrimPrefix(r.URL.Path, "/")
		http.ServeFile(w, r, filePath)
		return
	}

	// API endpoints
	// We support three methods on /api/links:
	//   GET    – list all links
	//   POST   – create a new link
	//   DELETE – delete a link by ID (path: /api/links/{id})
	if strings.HasPrefix(r.URL.Path, "/api/links") {
		switch r.Method {
		case http.MethodGet:
			links, err := h.store.List(r.Context())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if links == nil {
				links = []store.Link{}
			}
			json.NewEncoder(w).Encode(links)
			return

		case http.MethodPost:
			var payload struct {
				Title string `json:"title"`
				URL   string `json:"url"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
				return
			}
			created, err := h.store.Create(r.Context(), payload.Title, payload.URL)
			if err != nil {
				http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(created)
			return

		case http.MethodDelete:
			parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/links/"), "/")
			if len(parts) == 0 || parts[0] == "" {
				http.Error(w, `{"error":"missing id"}`, http.StatusBadRequest)
				return
			}
			idStr := parts[0]
			var id int64
			if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
				http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
				return
			}
			if err := h.store.Delete(r.Context(), id); err != nil {
				if err == sql.ErrNoRows {
					http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
				} else {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		// Unsupported method for this path.
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// If nothing matched, return 404.
	http.Error(w, "not found", http.StatusNotFound)
}
