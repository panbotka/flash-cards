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
  - `cards.go` — CRUD, suspend, restore, `GET /tags` for distinct tag list, `PUT /tags` for rename, `DELETE /tags/:tag` for deleting tag + its cards
  - `study.go` — Next card selection (with `?direction=` and `?tag=` filters), review submission, new card introduction
  - `import.go` — Bulk import with preview/commit pattern
  - `stats.go` — Six analytics endpoints
- `internal/auth/auth.go` — HMAC-SHA256 session cookies. Middleware skips `/api/auth/login` and `/api/auth/check`.
- `internal/importer/importer.go` — Delimiter auto-detection and text parsing.

No ORM — all database access uses raw SQL with `database/sql`.

### Frontend (React 19 + Vite + TypeScript)

- `frontend/src/api/client.ts` — All API types and fetch functions in one file. Generic `request<T>()` with 401 redirect handling.
- `frontend/src/App.tsx` — Auth check on mount, protected routing, NavBar rendered on all routes except login.
- `frontend/src/hooks/useStudySession.ts` — Core study flow: card fetching, flip state, review mutations, new-card continuation. Accepts `direction` and `tag` params.
- `frontend/src/hooks/useKeyboardShortcuts.ts` — Space (flip), 1-3 (rate), ignores input fields.
- `frontend/src/hooks/useSwipeRating.ts` — Touch swipe gestures for mobile rating (left=Easy, up=Good, right=Hard). Uses native touch listeners with `passive: false` and immediate `preventDefault()` for iOS Safari compatibility. Reads `enabled`/`onRate` via refs to keep listeners stable. Swipe-off animates card out with fade, new card fades in.
- Pages: `StudyPage`, `CardsPage`, `ImportPage`, `StatsPage`, `LoginPage`
- Components: `FlashCard` (3D CSS flip), `RatingButtons`, `TagFilter` (with edit mode + bottom sheet for rename/delete), `NavBar` (bottom tabs)

State management: TanStack Query for server state, local `useState` for UI state. No global store.

### Key Domain Concepts

**Dual SRS states per card**: Each card has two `srs_state` rows (`cz_en` and `en_cz`) with independent progress. The study page has a direction toggle (CZ→EN / EN→CZ) and a tag filter dropdown. The `StudyContent` component is keyed by `direction-tag` so UI state resets on change. Front/back props on `FlashCard` are swapped based on direction. Rating buttons are always visible (no need to flip first). On mobile, swipe gestures work immediately without flipping. The study page is non-scrollable (`fixed inset-0 overflow-hidden`).

**SM-2 state machine**: Cards progress through `new` → `learning` → `review`. Learning uses sub-day steps (1 min, 10 min). Review uses day-scale intervals multiplied by ease factor. The frontend sends ratings 2-4 (Hard/Good/Easy); the backend still accepts 1-4 but "Again" (1) is not exposed in the UI.

**Tag management**: Tags can be renamed (`PUT /tags`) or deleted with all their cards (`DELETE /tags/:tag`). Rename merges if the target tag already exists on some cards. The TagFilter component has an edit mode (pencil icon toggle) where tapping a tag opens a bottom action sheet with rename/delete options. The Cards page uses fixed viewport layout (`fixed inset-0 flex flex-col`) like the Study page. Tag input accepts comma or space delimiters.

**Stats accuracy values**: The API returns accuracy as a 0–1 fraction. The frontend multiplies by 100 for display.

## PWA & Icons

The app is PWA-installable. Static assets in `frontend/public/`:

- `icon.svg` — Main favicon (SVG, scales to any size)
- `icon-180.png` — Apple touch icon
- `icon-192.png`, `icon-512.png` — Standard PWA icons
- `icon-maskable-192.png`, `icon-maskable-512.png` — Maskable icons for Android adaptive icons (content in 80% safe zone, dark bg)
- `icon-maskable.svg` — Source SVG for maskable variants
- `manifest.json` — Web app manifest (standalone display, dark theme)

`index.html` links the manifest, apple-touch-icon, and SVG favicon. It also sets `viewport-fit=cover` and `apple-mobile-web-app-status-bar-style=black-translucent` for edge-to-edge PWA display.

**Safe area insets**: Each page applies `paddingTop: max(env(safe-area-inset-top, 0px), <default>)` via inline style to avoid content overlapping the iOS status bar/notch. The NavBar handles the bottom inset with `paddingBottom: env(safe-area-inset-bottom)`.

## Vite Dev Proxy

When running `npm run dev` in `frontend/`, API calls to `/api` are proxied to `http://localhost:3011` (configured in `vite.config.ts`). Run the Go backend separately via `APP_PORT=3011 ./bin/flash-cards`.
