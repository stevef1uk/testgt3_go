package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"linkshelf/internal/store"

	_ "github.com/mattn/go-sqlite3"

	"github.com/stretchr/testify/assert"
)

func newTestHandler(t *testing.T) *handler {
	db, err := sql.Open("sqlite3", ":memory:")
	assert.NoError(t, err)
	err = store.InitSchema(db)
	assert.NoError(t, err)
	s := store.NewStore(db)
	return NewHandler(s)
}

// Helper to perform a request against a handler.
func performRequest(h http.Handler, method, path string, body []byte) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

// Helper to perform a request with a JSON payload and appropriate Content‑Type.
func performJSONRequest(h http.Handler, method, path string, payload interface{}) *httptest.ResponseRecorder {
	var body []byte
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			panic(err)
		}
		body = b
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

// Test that the root path serves the index.html file.
func TestHandler_RootServesIndex(t *testing.T) {
	// Create a temporary web dir with an index.html.
	webDir := t.TempDir()
	indexPath := filepath.Join(webDir, "index.html")
	assert.NoError(t, os.WriteFile(indexPath, []byte("<html>OK</html>"), 0o644))

	// Change working directory so that the handler sees the correct relative path.
	origWD, _ := os.Getwd()
	assert.NoError(t, os.Chdir(webDir))
	defer os.Chdir(origWD)

	h := newTestHandler(t)

	rr := performRequest(h, "GET", "/", nil)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "OK")
}

// Test that static assets are served correctly and directory traversal is blocked.
func TestHandler_StaticAssets(t *testing.T) {
	webDir := t.TempDir()
	staticPath := filepath.Join(webDir, "static", "test.txt")
	assert.NoError(t, os.MkdirAll(filepath.Dir(staticPath), 0o755))
	assert.NoError(t, os.WriteFile(staticPath, []byte("static content"), 0o644))

	origWD, _ := os.Getwd()
	assert.NoError(t, os.Chdir(webDir))
	defer os.Chdir(origWD)

	h := newTestHandler(t)

	rr := performRequest(h, "GET", "/static/test.txt", nil)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "static content", rr.Body.String())

	rr = performRequest(h, "GET", "/static/../secret", nil)
	assert.Equal(t, http.StatusForbidden, rr.Code)
}

// Test that the /api/links endpoint returns an empty slice when no links exist.
func TestHandler_APILinks_Empty(t *testing.T) {
	h := newTestHandler(t)

	rr := performRequest(h, "GET", "/api/links", nil)
	assert.Equal(t, http.StatusOK, rr.Code)

	var links []store.Link
	assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &links))
	assert.NotNil(t, links)
	assert.Empty(t, links)
}

// Test that the /api/links endpoint returns stored links in descending order.
func TestHandler_APILinks_WithData(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	assert.NoError(t, err)
	assert.NoError(t, store.InitSchema(db))

	s := store.NewStore(db)

	// Insert two links.
	_, err = s.Create(context.Background(), "First", "https://first.example")
	assert.NoError(t, err)
	_, err = s.Create(context.Background(), "Second", "https://second.example")
	assert.NoError(t, err)

	h := NewHandler(s)

	rr := performRequest(h, "GET", "/api/links", nil)
	assert.Equal(t, http.StatusOK, rr.Code)

	var links []store.Link
	assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &links))
	assert.Len(t, links, 2)
	// Newest link first.
	assert.Equal(t, "Second", links[0].Title)
	assert.Equal(t, "First", links[1].Title)
}

// ---------------------------------------------------------------------------
// Additional tests required by the architecture specification
// ---------------------------------------------------------------------------

func TestHandler_PostLink_Valid(t *testing.T) {
	// Set up a fresh in‑memory store.
	h := newTestHandler(t)

	// Prepare a valid JSON payload.
	payload := map[string]string{
		"title": "New Link",
		"url":   "https://example.com",
	}
	rr := performJSONRequest(h, http.MethodPost, "/api/links", payload)

	// Expect 201 Created.
	assert.Equal(t, http.StatusCreated, rr.Code)

	var created store.Link
	assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &created))
	assert.Equal(t, payload["title"], created.Title)
	assert.Equal(t, payload["url"], created.URL)
	assert.NotZero(t, created.ID)
	assert.NotEmpty(t, created.CreatedAt)
}

func TestHandler_PostLink_Invalid(t *testing.T) {
	h := newTestHandler(t)

	// Payload missing the required 'url' field.
	payload := map[string]string{
		"title": "Bad Link",
	}

	rr := performJSONRequest(h, http.MethodPost, "/api/links", payload)

	// Expect 400 Bad Request with JSON error.
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var errResp map[string]string
	assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &errResp))
	_, ok := errResp["error"]
	assert.True(t, ok, "expected error field in response")
}

func TestHandler_DeleteLink(t *testing.T) {
	// Create a link first.
	h := newTestHandler(t)

	payload := map[string]string{
		"title": "To Delete",
		"url":   "https://delete.me",
	}
	rrCreate := performJSONRequest(h, http.MethodPost, "/api/links", payload)
	assert.Equal(t, http.StatusCreated, rrCreate.Code)

	var created store.Link
	assert.NoError(t, json.Unmarshal(rrCreate.Body.Bytes(), &created))

	// Now delete it.
	deletePath := "/api/links/" + strconv.FormatInt(created.ID, 10)
	rrDel := performRequest(h, http.MethodDelete, deletePath, nil)
	assert.Equal(t, http.StatusNoContent, rrDel.Code)

	// Subsequent delete should return 404.
	rrDel2 := performRequest(h, http.MethodDelete, deletePath, nil)
	assert.Equal(t, http.StatusNotFound, rrDel2.Code)
}
