# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
# Full build (frontend + backend)
make build

# Development: builds both and runs on port 3011, no auth
./dev.sh

# Frontend only
cd frontend && npm install && npm run build

# Backend only (requires frontend/dist/ to exist for embed)
CGO_ENABLED=1 go build -o ./bin/flash-cards ./cmd/server

# Docker
docker compose up --build    # port 8080, APP_PASSWORD=changeme

# Frontend lint and type-check
cd frontend && npm run lint && npx tsc --noEmit
```

CGO is required (`CGO_ENABLED=1`) because of the `mattn/go-sqlite3` driver.

## Environment Variables

- `APP_PORT` — Listen port (default: `8080`)
- `APP_PASSWORD` — Enables auth when set; empty = no auth (dev mode)
- `DB_PATH` — SQLite path (default: `./data/flash-cards.db`, Docker: `/data/flash-cards.db`)

## Architecture

**Single binary deployment**: The React frontend is compiled to `frontend/dist/`, then embedded into the Go binary via `embed.FS` in `embed.go`. The server serves the SPA with a fallback to `index.html` for client-side routing.

### Backend (Go + Gin)

- `cmd/server/main.go` — Entry point. Wires DB, auth middleware, all handler groups, and embedded frontend serving.
- `internal/db/` — SQLite connection (WAL mode, foreign keys) and version-based schema migrations.
- `internal/models/models.go` — All structs and request/response types. Shared across handlers.
- `internal/srs/sm2.go` — SM-2 spaced repetition algorithm. `ProcessReview()` is pure (returns new state, doesn't mutate input).
- `internal/handlers/` — Five handler files, each with a struct holding `*sql.DB` and a `Register(r *gin.RouterGroup)` method:
  - `auth.go` — Login/logout/check
  - `cards.go` — CRUD, suspend, restore
  - `study.go` — Next card selection, review submission, new card introduction
  - `import.go` — Bulk import with preview/commit pattern
  - `stats.go` — Six analytics endpoints
- `internal/auth/auth.go` — HMAC-SHA256 session cookies. Middleware skips `/api/auth/login` and `/api/auth/check`.
- `internal/importer/importer.go` — Delimiter auto-detection and text parsing.

No ORM — all database access uses raw SQL with `database/sql`.

### Frontend (React 19 + Vite + TypeScript)

- `frontend/src/api/client.ts` — All API types and fetch functions in one file. Generic `request<T>()` with 401 redirect handling.
- `frontend/src/App.tsx` — Auth check on mount, protected routing, NavBar rendered on all routes except login.
- `frontend/src/hooks/useStudySession.ts` — Core study flow: card fetching, flip state, review mutations, new-card continuation.
- `frontend/src/hooks/useKeyboardShortcuts.ts` — Space (flip), 1-4 (rate), ignores input fields.
- Pages: `StudyPage`, `CardsPage`, `ImportPage`, `StatsPage`, `LoginPage`
- Components: `FlashCard` (3D CSS flip), `RatingButtons`, `TagFilter`, `NavBar` (bottom tabs)

State management: TanStack Query for server state, local `useState` for UI state. No global store.

### Key Domain Concepts

**Single SRS state per card**: Each card has one `srs_state` row (direction `cz_en`). During study, Czech is shown first; the card can be flipped back and forth freely. Rating buttons appear after the first flip and remain visible.

**SM-2 state machine**: Cards progress through `new` → `learning` → `review`. Learning uses sub-day steps (1 min, 10 min). Review uses day-scale intervals multiplied by ease factor. "Again" rating on a review card sends it back to learning.

**Stats accuracy values**: The API returns accuracy as a 0–1 fraction. The frontend multiplies by 100 for display.

## PWA & Icons

The app is PWA-installable. Static assets in `frontend/public/`:

- `icon.svg` — Main favicon (SVG, scales to any size)
- `icon-180.png` — Apple touch icon
- `icon-192.png`, `icon-512.png` — Standard PWA icons
- `icon-maskable-192.png`, `icon-maskable-512.png` — Maskable icons for Android adaptive icons (content in 80% safe zone, dark bg)
- `icon-maskable.svg` — Source SVG for maskable variants
- `manifest.json` — Web app manifest (standalone display, dark theme)

`index.html` links the manifest, apple-touch-icon, and SVG favicon.

## Vite Dev Proxy

When running `npm run dev` in `frontend/`, API calls to `/api` are proxied to `http://localhost:3011` (configured in `vite.config.ts`). Run the Go backend separately via `APP_PORT=3011 ./bin/flash-cards`.
