
# Link Shelf – Simple Build Spec
 
## What you're building
 
A personal bookmark manager: save URLs with titles, list them, delete them. Go backend + vanilla HTML frontend + SQLite storage.
 
## Run it
 
```bash
cd linkshelf
go mod tidy
go run ./cmd/server
# visit http://localhost:8080
```
 
## File layout
 
```
linkshelf/
├── go.mod
├── cmd/server/main.go
├── internal/
│   ├── store/store.go
│   └── api/handlers.go
└── web/
    ├── index.html
    ├── app.js
    └── style.css
```
 
## go.mod
 
```
module linkshelf
 
go 1.22
 
require modernc.org/sqlite v1.29.0
```
 
## Data model
 
One table, one struct:
 
```go
type Link struct {
    ID        int64  `json:"id"`
    Title     string `json:"title"`
    URL       string `json:"url"`
    CreatedAt string `json:"created_at"` // RFC3339
}
```
 
```sql
CREATE TABLE IF NOT EXISTS links (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  title TEXT NOT NULL,
  url   TEXT NOT NULL,
  created_at TEXT NOT NULL
);
```
 
## Store (`internal/store/store.go`)
 
Implement these three functions backed by SQLite:
 
```go
func (s *Store) List(ctx context.Context) ([]Link, error)      // ORDER BY id DESC
func (s *Store) Create(ctx context.Context, title, url string) (Link, error)
func (s *Store) Delete(ctx context.Context, id int64) error
```
 
Validation rules for `Create`:
- Title: non-empty, max 200 chars
- URL: non-empty, must start with `http://` or `https://`
- Return an error (don't panic) if validation fails
- Return an error if Delete is called with an id that doesn't exist
## HTTP API (`internal/api/handlers.go`)
 
| Method | Path | Success | Error |
|--------|------|---------|-------|
| GET | `/` | 200, serve `web/index.html` | — |
| GET | `/static/{file}` | 200, serve file from `web/` | 404 |
| GET | `/api/links` | 200, JSON array | — |
| POST | `/api/links` | 201, JSON of created link | 400 `{"error":"..."}` |
| DELETE | `/api/links/{id}` | 204 | 404 `{"error":"..."}` |
 
The static handler must reject any path containing `..`.
 
## `cmd/server/main.go`
 
- Open SQLite at `linkshelf.db` (same directory as where you run the command)
- Register routes on `http.DefaultServeMux`
- Listen on `:8080`
- Print `listening on :8080` when ready
## Frontend (`web/`)
 
**index.html** — a form with a title input, a URL input, an Add button, and an empty `<ul id="links">` list.
 
**app.js** — on page load, fetch `/api/links` and populate the list. Each item shows the title (as a link), the URL, and a Delete button. The form submits via `POST /api/links`. Delete calls `DELETE /api/links/{id}`. All updates refresh the list without reloading the page.
 
**style.css** — basic readable styles. Nothing fancy required.
 
## Tests (optional but encouraged)
 
If you write tests, put store tests in `internal/store/store_test.go` using an in-memory SQLite DB (`:memory:`), and handler tests in `internal/api/handlers_test.go` using `net/http/httptest`.
 
Key cases to cover:
- Create + List returns the link
- Create with bad URL returns error
- DELETE existing → 204; DELETE missing → 404
- GET `/api/links` returns `[]` (not `null`) when empty
## Definition of done
 
1. `go mod tidy && go build ./...` succeeds with no errors
2. Server starts and the UI loads at `http://localhost:8080`
3. You can add a bookmark, see it in the list, and delete it
