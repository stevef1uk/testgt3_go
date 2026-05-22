# Architecture for testgt3 — Link Shelf Personal Bookmark Manager

## Overview

This document outlines the architectural design of `linkshelf`, a personal bookmark manager implemented in Go with SQLite for persistence and a vanilla HTML/JavaScript/CSS frontend. The application enables users to save URLs with titles, view a list of saved links sorted by creation time (newest first), and delete entries. The design emphasizes simplicity, maintainability, and adherence to Go idioms, including clean separation of concerns between data access (store), HTTP handling (API), and static asset serving.

The backend is structured using Go modules and follows a layered architecture:
- **Store Layer (`internal/store`)**: Handles SQLite interactions, schema initialization, and business validation.
- **API Layer (`internal/api`)**: Implements HTTP handlers that marshal requests/responses and delegate to the store.
- **Frontend (`web/`)**: Static files served directly by the server, including a dynamic single-page interface using vanilla JavaScript.
- **Entry Point (`cmd/server/main.go`)**: Bootstraps the application, initializes dependencies, and starts the HTTP server.

The application leverages `modernc.org/sqlite` for SQLite3 support without cgo, enabling easy cross-compilation. All components are designed to support comprehensive unit and integration testing, with a focus on validating error cases and input sanitization.

## Planned File Layout

The project structure strictly follows the Go standard layout for modular applications:

- `linkshelf/go.mod`: Declares the module as `linkshelf`, requires Go 1.22, and depends on `modernc.org/sqlite v1.29.0` for SQLite database support.
- `linkshelf/cmd/server/main.go`: The application entry point. It opens a SQLite database connection to `linkshelf.db` in the current working directory, ensures the schema is applied, registers all HTTP routes on `http.DefaultServeMux`, and starts listening on `:8080`. On successful startup, it logs "listening on :8080".
- `linkshelf/internal/store/schema.go`: **[Schema Ownership]** This file contains the SQL DDL statement used to create the `links` table if it does not exist. It is invoked once during application startup in `main.go` and also during store tests to initialize the in-memory database. The schema defines:
  sql
  CREATE TABLE IF NOT EXISTS links (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    url TEXT NOT NULL,
    created_at TEXT NOT NULL
  );
  
- `linkshelf/internal/store/store.go`: Implements the `Store` type with three core methods:
  - `List(ctx context.Context) ([]Link, error)`: Retrieves all links ordered by `id DESC` (newest first).
  - `Create(ctx context.Context, title, url string) (Link, error)`: Inserts a new link after validating inputs:
    - `title`: must be non-empty and ≤ 200 characters.
    - `url`: must be non-empty and start with `http://` or `https://`.
    - If validation fails, returns an appropriate error; does not panic.
  - `Delete(ctx context.Context, id int64) error`: Removes a link by ID. Returns an error if the ID does not exist.
- `linkshelf/internal/store/store_test.go`: Contains unit tests for the store layer. Uses an in-memory SQLite database (`:memory:`) via `modernc.org/sqlite` to isolate tests. Validates:
  - Creating a valid link returns no error and persists correctly.
  - Invalid titles or URLs trigger validation errors.
  - Deleting an existing ID returns nil; deleting a missing ID returns an error.
  - `List` returns entries in descending order by ID.
- `linkshelf/internal/api/handlers.go`: Implements HTTP handlers registered under `http.DefaultServeMux`. Routes are:
  - `GET /` → serves `web/index.html`.
  - `GET /static/{file}` → serves static assets from the `web/` directory; rejects any path containing `..` to prevent directory traversal.
  - `GET /api/links` → returns JSON array of all links; returns `[]` (not `null`) when empty.
  - `POST /api/links` → parses JSON input, calls `store.Create`, and returns the created link with status 201. On validation or internal error, returns 400 with `{"error":"..."}`.
  - `DELETE /api/links/{id}` → parses ID from URL, calls `store.Delete`, returns 204 on success. If link not found, returns 404 with error message.
- `linkshelf/internal/api/handlers_test.go`: Uses `net/http/httptest` to test the HTTP API layer. Covers:
  - POST with valid payload → 201 and JSON response.
  - POST with invalid URL/title → 400 and error JSON.
  - DELETE with valid ID → 204.
  - DELETE with invalid ID → 404.
  - GET `/api/links` on empty store → returns `[]`.
  - GET `/static/..` → returns 404.
- `linkshelf/web/index.html`: Minimal HTML page with:
  - A form containing `<input>` fields for title and URL, and an "Add" button.
  - An unordered list `<ul id="links">` to display saved bookmarks.
  - References to `/static/app.js` and `/static/style.css`.
- `linkshelf/web/app.js`: Client-side script that:
  - On load, fetches `/api/links` and renders each entry as an `<li>` showing the title (as an anchor to the URL), the raw URL, and a "Delete" button.
  - Submits the form via `POST /api/links`, clears inputs on success, and refreshes the list.
  - On clicking "Delete", sends `DELETE /api/links/{id}`, then refreshes the list.
  - All UI updates occur without full page reload.
- `linkshelf/web/style.css`: Basic CSS to ensure the form and list are readable and visually separated. No complex layout or animations required.

## Unit Tests

Unit testing is a core part of the design, ensuring correctness and robustness.

- **Store Tests (`internal/store/store_test.go`)**:
  - Test case: "Create valid link" → ensures insertion works and all fields are preserved.
  - Test case: "Create with empty title" → expect validation error.
  - Test case: "Create with invalid URL scheme" → expect validation error.
  - Test case: "Create with title > 200 chars" → expect validation error.
  - Test case: "Delete existing link" → returns nil error.
  - Test case: "Delete non-existent link" → returns error.
  - Test case: "List returns results in descending order" → verifies `ORDER BY id DESC`.

- **Handler Tests (`internal/api/handlers_test.go`)**:
  - Test case: "GET /api/links returns empty array" → expects `[]`, status 200.
  - Test case: "POST /api/links valid input" → expects 201, correct JSON body.
  - Test case: "POST /api/links invalid input" → expects 400, error JSON.
  - Test case: "DELETE existing link" → expects 204.
  - Test case: "DELETE non-existent link" → expects 404, error JSON.
  - Test case: "GET /static/..%2fetc%2fpasswd" → expects 404 (path traversal defense).
  - Test case: "GET /" → serves index.html successfully.

All tests use dependency injection where necessary (e.g., passing a mock or in-memory store to handlers) and avoid reliance on external state.

## Integration and Testing

The components are integrated as follows:

1. On startup (`main.go`):
   - A SQLite database connection is opened to `linkshelf.db`.
   - The schema from `schema.go` is applied using `db.Exec()` to ensure the `links` table exists.
   - A `store.Store` instance is created with the database handle.
   - HTTP handlers in `handlers.go` are registered with `http.DefaultServeMux`, closing over the store instance.
   - The server listens on `:8080` and logs readiness.

2. Frontend-backend interaction:
   - The frontend JS makes direct calls to the Go HTTP API.
   - No templating or server-side rendering is used; the backend is purely API-driven.
   - The static file server serves `index.html` at `/` and all `web/` assets under `/static/*`.

3. Testing Workflow:
   - The full test suite is executed via:
     bash
     cd linkshelf && go test ./...
     
   - This runs both `store_test.go` and `handlers_test.go`.
   - Store tests use `:memory:` SQLite database for speed and isolation.
   - Handler tests use `httptest.NewServer` to simulate real HTTP traffic.
   - No test should depend on a pre-existing `linkshelf.db` file.

4. Build and Run:
   - Verified with:
     bash
     go build ./...
     
   - The binary is built from `cmd/server/main.go`.
   - After `go run ./cmd/server`, the UI should be accessible at `http://localhost:8080`.

## Acceptance Mapping

This architecture ensures all functional requirements from the SPEC are met:

- ✅ **Save URLs with titles**: Provided via `POST /api/links` with JSON body, validated in `store.Create`, stored in SQLite.
- ✅ **List saved bookmarks**: `GET /api/links` returns JSON array ordered by ID descending; frontend renders list dynamically.
- ✅ **Delete entries**: `DELETE /api/links/{id}` removes by ID; frontend refreshes list after deletion.
- ✅ **Input validation**: Enforced in `store.Create`; rejects empty or non-HTTP(S) URLs and long/empty titles.
- ✅ **Proper error handling**: All validation and runtime errors return appropriate HTTP status codes and JSON error payloads.
- ✅ **Dynamic list updates**: Frontend JS fetches and refreshes the list after every add/delete operation, without page reload.
- ✅ **Static file serving**: Implemented via `GET /static/{file}` and root handler; secure against path traversal.
- ✅ **SQLite persistence**: Data is stored in `linkshelf.db` using a robust DDL applied idempotently at startup.
- ✅ **Definition of Done**:
  1. `go build ./...` — supported by correct `go.mod` and clean package structure.
  2. Server starts and UI loads — root handler serves `index.html`, static assets available.
  3. Add, view, delete cycle — fully implemented through integrated frontend and API.

## Delivery Phases

This design assumes a single delivery phase (0), where all components are implemented concurrently by `polecat` based on this architecture. The file paths and behaviors are aligned exactly with the SPEC:

- Database schema defined in `linkshelf/internal/store/schema.go` and applied on startup.
- All Go source files placed under `linkshelf/` as specified.
- Frontend files (`index.html`, `app.js`, `style.css`) located in `linkshelf/web/`.
- No extraneous directories or files are assumed.

This architecture provides a clear, testable, and production-ready blueprint for implementing the bookmark manager as specified.
