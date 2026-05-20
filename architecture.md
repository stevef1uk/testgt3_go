# Architecture for testgt3: Personal Bookmark Manager

## Overview
This architecture defines the structure and organization of `linkshelf`, a personal bookmark manager built using Go for the backend, SQLite for persistence, and Vanilla JavaScript for the frontend. The system follows clean architectural principles with separation of concerns across layers: data store, API handlers, and static web assets. Standard web development practices are followed, including RESTful endpoint design, secure HTTP handling, and modular Go package layout.

The project emphasizes simplicity, testability, and end-to-end verifiability. The backend exposes a JSON API over HTTP, while the frontend consumes this API through client-side JavaScript to render and manipulate bookmarks. SQLite ensures lightweight, file-based storage ideal for single-user use without requiring a database server. All code is intended to be portable and buildable with standard tooling.

This architecture document serves as a blueprint for downstream implementation by Polecat, which will generate source files according to the planned layout and behavioral contracts described below.

## Planned file layout

All source files reside under the `linkshelf/` directory as defined in the SPEC.

- **`linkshelf/go.mod`**  
  Go module definition specifying the project’s module path and Go version. This enables proper dependency management and module scoping. Expected content includes `module linkshelf` and `go 1.22` or similar.

- **`linkshelf/cmd/server/main.go`**  
  Entry point for the HTTP server. This file initializes the SQLite database connection via the store, sets up routing using the `net/http` mux, registers API handlers from the internal API package, and starts the HTTP service on a configurable port (default: 8080). It handles graceful shutdown and logging setup.

- **`linkshelf/internal/store/store.go`**  
  Package `store` encapsulates data access logic and provides an abstraction over SQLite operations. It contains the `Bookmark` struct (with fields: ID, Title, URL, CreatedAt), and methods such as `GetAllBookmarks()`, `GetBookmark(id int)`, `CreateBookmark(b *Bookmark)`, `UpdateBookmark(b *Bookmark)`, and `DeleteBookmark(id int)`. This package opens the SQLite database, ensures schema initialization (via embedded DDL or migration), and uses prepared statements to prevent SQL injection.

- **`linkshelf/internal/api/handlers.go`**  
  Implements HTTP handlers that expose CRUD operations on bookmarks via JSON. Each handler wraps corresponding store methods: `getBookmarksHandler`, `createBookmarkHandler`, etc. Handlers validate input, marshal/unmarshal JSON, and return appropriate HTTP status codes (200, 201, 400, 404, 500). Each endpoint follows REST conventions:
  - `GET /api/bookmarks` → returns list
  - `POST /api/bookmarks` → creates new
  - `GET /api/bookmarks/:id` → fetch one
  - `PUT /api/bookmarks/:id` → update
  - `DELETE /api/bookmarks/:id` → remove

- **`linkshelf/web/index.html`**  
  Static HTML file serving as the SPA entry point. It includes a clean UI for listing, adding, editing, and removing bookmarks. Uses semantic HTML and references `app.js` and `style.css`. Designed mobile-first with responsive layout.

- **`linkshelf/web/app.js``  
  Client-side JavaScript that interacts with the Go backend via `fetch()` calls. Handles form submissions, populates the bookmark list on load, manages DOM updates, and provides basic validation and user feedback. Uses event delegation and avoids jQuery or frameworks.

- **`linkshelf/web/style.css`**  
  Minimal stylesheet defining layout, spacing, and visual consistency. Includes styles for form inputs, buttons, bookmark cards, and responsive breakpoints. Intentionally kept small and dependency-free.

- **`linkshelf/internal/store/store_test.go`**  
  Unit tests for the `store` package. Uses an in-memory SQLite instance (`:memory:`) to test all CRUD operations in isolation. Each test function validates correctness and edge cases (e.g., retrieving nonexistent ID, duplicate URLs, empty titles). Relies on `testing` package and ensures full coverage of public methods.

- **`linkshelf/internal/api/handlers_test.go`**  
  Integration-style tests for HTTP handlers. Uses `net/http/httptest` to simulate requests against wrapped handler functions. Tests verify correct JSON responses, status codes, and behavior under invalid input (malformed JSON, missing fields). Uses a real store instance (in-memory) to ensure API ↔ store contract integrity.

## Integration and testing

Components integrate through well-defined interfaces:
- `main.go` depends on `store` and `handlers`, wiring them into the HTTP server.
- `handlers.go` calls methods on a `store.Store` interface, allowing for mock injection in tests.
- The frontend (`app.js`) communicates with the backend over HTTP, with no direct database access.

Polecat will implement and verify the following:

1. **Build and run server:**
   sh
   cd linkshelf
   go build -o bin/server cmd/server/main.go
   ./bin/server
   

2. **Test execution:**
   sh
   go test ./internal/store/...
   go test ./internal/api/...
   

3. **End-to-end verification:**
   - Start server and open `http://localhost:8080`
   - Confirm `index.html` loads, form works, bookmarks persist
   - Use `curl` to validate API:
     sh
     curl http://localhost:8080/api/bookmarks
     curl -X POST http://localhost:8080/api/bookmarks -H "Content-Type: application/json" -d '{"title":"Test","url":"https://example.com"}'
     

4. **SQLite schema verification:**
   sql
   CREATE TABLE IF NOT EXISTS bookmarks (
       id INTEGER PRIMARY KEY AUTOINCREMENT,
       title TEXT NOT NULL,
       url TEXT NOT NULL UNIQUE,
       created_at DATETIME DEFAULT CURRENT_TIMESTAMP
   );
   

## Acceptance mapping

| SPEC Requirement                                    | Architecture Coverage                                                                 |
|-----------------------------------------------------|----------------------------------------------------------------------------------------|
| Backend with Go HTTP server                         | Implemented in `main.go` and `handlers.go` using standard `net/http`.                  |
| SQLite for persistence                              | Handled by `store.go` using `modernc.org/sqlite` or Go’s `database/sql` with sqlite3.  |
| Frontend with Vanilla JS                            | `app.js` uses native DOM and `fetch`; no frameworks required.                          |
| Support bookmark CRUD operations                    | Full REST API implemented in handlers + store methods.                                 |
| Store title, URL, creation timestamp                | `Bookmark` struct includes all three fields; enforced in SQLite schema.                |
| Web interface to view and manage bookmarks          | `index.html` and `app.js` provide full UI interaction.                                 |
| End-to-end testing coverage                         | Both unit (`store_test.go`) and integration (`handlers_test.go`) tests are defined.   |
| Standard web development concepts                   | Clean separation of backend, API, frontend; stateless HTTP; client-server model.       |
| Buildable Go project                                | `go.mod` enables `go build`/`go test`; project follows Go layout standards.            |
| Testable components                                 | All core logic is testable using in-memory DB and `httptest`; 100% coverage expected.  |
| No unnecessary dependencies                         | Only standard library and minimal externals (e.g., SQLite driver).                     |

## Error handling strategy

Each layer implements appropriate error handling:
- **Store**: Returns Go error types; handles DB connection, constraint violations, and transaction failures.
- **API handlers**: Log errors, return 500 for internal errors, parse/validation errors return 400.
- **Frontend**: `app.js` catches fetch and input errors, displays user-friendly messages.
- **Startup**: Server fails fast on DB connection or bind errors, logs and exits.

## Data model

go
type Bookmark struct {
    ID        int       `json:"id"`
    Title     string    `json:"title"`
    URL       string    `json:"url"`
    CreatedAt time.Time `json:"created_at"`
}


Enforced uniqueness on `url`. Input validation ensures non-empty title and valid URL format.

## Security considerations

- No authentication (per spec) but designed for single-user localhost use.
- SQL injection prevented via prepared statements.
- CORS not enabled (assumes frontend and backend on same origin).
- Content-Type enforcement in handlers rejects malformed requests.

## Development workflow alignment

This architecture supports iterative development:
1. First, build and test the store layer.
2. Then implement and test handlers.
3. Build and connect frontend last.
4. Run all tests before deployment.

Logging via `log` package enables debugging; structured logging can be added later if needed.

