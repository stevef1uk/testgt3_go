# Project Specification: Link Shelf

## Overview

**Link Shelf** is a small personal bookmark manager: save URLs with titles, list them, and delete entries you no longer need. It is built as a **Go HTTP server** (stdlib `net/http` only — no Gin, no Chi) with a **vanilla HTML/CSS/JavaScript** frontend. Data persists in a local **SQLite** file via `database/sql` and `modernc.org/sqlite` (pure Go driver, no CGO).

This spec is sized for the Gas Town **rig-flow** orchestrator (Mayor → Architect → Planner → Polecat → QA): enough surface area for multiple implementation beads, but small enough to finish in one rig session.

## Goals

- Runnable app: `go run ./cmd/server` serves UI + JSON API on port **8080**
- CRUD for bookmarks (create, list, delete) with validation
- Clean separation: `internal/store` (data), `internal/api` (HTTP), `web/` (static UI)
- Automated verification via `go test ./...` and `go vet ./...`
- No accounts, no auth, no external services

## Non-goals

- Tags, full-text search, import/export, browser extensions
- Docker, Kubernetes, or deployment manifests
- Frontend frameworks (React, Vue) or bundlers (Vite, webpack)
- PostgreSQL or cloud databases

## Tech stack

| Layer | Technology |
|-------|------------|
| Language | Go **1.22+** |
| HTTP | `net/http`, `http.ServeMux` |
| Database | SQLite file `linkshelf.db` (default path next to binary cwd) |
| SQL driver | `modernc.org/sqlite` |
| Frontend | HTML5, vanilla JavaScript, CSS |
| Tests | `testing`, `net/http/httptest`, temp SQLite files |

## Layout root: `linkshelf/`

All implementation paths are **relative to `linkshelf/`** in the mayor/rig worktree (repo root for this project).

```
linkshelf/
├── go.mod
├── README.md
├── cmd/
│   └── server/
│       └── main.go              # flags (-addr, -db), mux, graceful-ish startup
├── internal/
│   ├── store/
│   │   ├── store.go             # Link struct, Store interface
│   │   ├── sqlite.go            # schema migrate, CRUD
│   │   └── sqlite_test.go       # store tests with :memory: or temp file
│   └── api/
│       ├── handlers.go            # routes + JSON helpers
│       └── handlers_test.go       # httptest per route
└── web/
    ├── index.html               # shell + form + list container
    ├── app.js                   # fetch API, DOM updates
    └── style.css                # readable layout, responsive-enough list
```

## Data model

**Link** (JSON and DB row):

| Field | Type | Rules |
|-------|------|-------|
| `id` | int64 | Auto-increment primary key |
| `title` | string | Required, trimmed, 1–200 runes |
| `url` | string | Required, trimmed, must parse as `http` or `https` URL |
| `created_at` | RFC3339 string | Set by server on create |

SQLite table `links`:

```sql
CREATE TABLE IF NOT EXISTS links (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  title TEXT NOT NULL,
  url TEXT NOT NULL,
  created_at TEXT NOT NULL
);
```

## HTTP API

Base URL when running locally: `http://localhost:8080`

| Method | Path | Behavior |
|--------|------|----------|
| GET | `/` | Serve `web/index.html` |
| GET | `/static/{file}` | Serve files from `web/` (only `app.js`, `style.css`; reject path traversal) |
| GET | `/api/links` | `200` + JSON array of links, newest first (`id DESC`) |
| POST | `/api/links` | Body `{"title":"...","url":"..."}` → `201` + created link JSON; `400` on validation error |
| DELETE | `/api/links/{id}` | `204` if deleted; `404` if id missing |

**Content-Type:** `application/json` for API bodies and responses.

**Error shape** (4xx): `{"error":"human-readable message"}`

## Frontend behavior (`web/`)

1. On load, `GET /api/links` and render an unordered list (title as link text opening URL in new tab, show URL, delete button per row).
2. Form: title input, URL input, "Add" button → `POST /api/links` → refresh list on success; show inline error text on failure.
3. Delete button → `DELETE /api/links/{id}` → remove row or refresh list on success.
4. No build step: scripts loaded directly from `/static/app.js`.
5. Basic accessibility: `<label>` for inputs, keyboard-submit on form.

## Run and verify

From `linkshelf/` in the rig worktree:

```bash
go mod tidy
go test ./...
go vet ./...
go run ./cmd/server
# open http://localhost:8080 — add a link, confirm list + delete
```

**Workflow QA command** (run from mayor/rig repo root):

```bash
cd linkshelf && go test ./... && go vet ./...
```

## Polecat implementation requirements

Implement **real, runnable code** — not placeholders (`TODO`, empty handlers, or `panic("not implemented")`). Each implementation bead title should reference a concrete path under `linkshelf/`.

### Per-file minimum acceptance

**`go.mod`** — module path `linkshelf` (or `github.com/<owner>/linkshelf` if rig remote dictates); require Go 1.22+; depend on `modernc.org/sqlite`.

**`cmd/server/main.go`** — Parse flags `-addr` (default `:8080`), `-db` (default `linkshelf.db`); open store; register mux routes from `api`; serve; log listen address on start.

**`internal/store/store.go`** — `Link` struct with JSON tags; `Store` interface: `List(ctx) ([]Link, error)`, `Create(ctx, title, url) (Link, error)`, `Delete(ctx, id int64) error`.

**`internal/store/sqlite.go`** — Open DB, run schema migration on init, implement `Store`; validate title/url on create; normalize URL (trim, require scheme).

**`internal/store/sqlite_test.go`** — Tests: create + list returns link; duplicate titles allowed; invalid URL rejected; delete removes row; delete missing id returns error.

**`internal/api/handlers.go`** — Constructor taking `Store`; methods for each route; static file handler safe against `..` in path.

**`internal/api/handlers_test.go`** — `httptest`: GET `/api/links` empty `[]`; POST valid → 201; POST invalid URL → 400; DELETE existing → 204; DELETE missing → 404; GET `/` returns 200 HTML.

**`web/index.html`** — Form + `#link-list` (or equivalent); links `/static/style.css` and `/static/app.js`.

**`web/app.js`** — `loadLinks`, `addLink`, `deleteLink` using `fetch`; DOM updates without full page reload.

**`web/style.css`** — Readable typography, spaced form, list rows with clear delete control (not unstyled browser default only).

**`README.md`** — Prerequisites (Go 1.22+), `go run`, `go test`, default URL, example `curl` for POST/GET/DELETE.

### Suggested implementation beads (planner)

Planner may split work roughly as:

1. `Implement linkshelf/go.mod` + module bootstrap
2. `Implement linkshelf/internal/store/` (interface + sqlite + tests)
3. `Implement linkshelf/internal/api/` (handlers + tests)
4. `Implement linkshelf/cmd/server/main.go`
5. `Implement linkshelf/web/` (index, app.js, style.css)
6. `Implement linkshelf/README.md`

Titles must start with **`Implement linkshelf/`** so `bd list` and rig-flow QA can filter them.

### Verification before `bd close`

From `linkshelf/`:

```bash
go test ./...
go vet ./...
```

If the bead touched `web/` only, still run tests (should stay green). If the bead touched backend, manual smoke test in browser is encouraged but **QA gate is `go test` + `go vet`**.

Do **not** close implementation beads until the bead's file(s) exist, are non-stub, and `go test ./...` passes from `linkshelf/`.

## Testing scope for automated rig QA

**Required for workflow pass:**

```bash
cd linkshelf && go test ./... && go vet ./...
```

### Go tests (must exist)

**Store**

- Create link → List contains it with `id > 0` and non-empty `created_at`
- Create with empty title → error
- Create with `not-a-url` → error
- Delete existing id → success; List no longer contains it
- Delete unknown id → error

**API (httptest + in-memory or temp DB store)**

- GET `/api/links` → `200`, `[]` initially
- POST valid JSON → `201`, body has `id` and fields
- POST missing url → `400`
- DELETE after POST → `204`; subsequent GET list empty
- GET `/` → `200`, body contains `<form` or add-link UI marker

## Security notes (keep simple)

- Static handler must reject `..` and absolute paths
- No CORS configuration required (same-origin UI)
- No secrets in repo; DB file is local dev data only

## Success criteria (definition of done)

1. `go test ./...` and `go vet ./...` pass from `linkshelf/`
2. Server starts on `:8080` and serves the UI
3. User can add a link, see it in the list, open it, and delete it
4. All paths in the layout tree exist with substantive code
5. `README.md` documents run/test commands accurately

## Operator workflow

```bash
# Register rig (example remote — use your fork)
gt rig add linkshelf https://github.com/<you>/linkshelf.git

# Copy this SPEC into the rig worktree
cp /path/to/gastown/docs/examples/orchestrator/link-shelf/SPEC.md \
   ~/gt/linkshelf/mayor/rig/SPEC.md

# Index profile for orchestrator prompts
gt rig spec-index linkshelf --force

# Start rig-flow
gt mayor workflow start rig-flow --rig linkshelf
gt nudge mayor "Kick off Link Shelf per SPEC.md"
```

Mayor coordinates: Architect (`architecture.md`), Planner (`plan.md` + beads), Polecat (implementation beads), QA (review + test command).

