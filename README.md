# Flash Cards

A Czech-English vocabulary study tool with spaced repetition (SM-2). Single Go binary with an embedded React frontend, designed for mobile-first use.

## Features

- **Spaced repetition** — SM-2 algorithm with learning steps (1m, 10m), graduating intervals, and three rating levels (Hard/Good/Easy)
- **Free-flip study** — Cards show Czech first with unlimited back-and-forth flipping; rate when ready
- **Swipe gestures** — On mobile, swipe left (Hard), up (Good), or right (Easy) with visual feedback
- **Bulk import** — Paste or upload CSV/TSV content with auto-delimiter detection and duplicate checking
- **Statistics** — Activity heatmap, accuracy tracking, maturity distribution, forecast, and hardest cards
- **Tag filtering** — Organize cards with tags and filter study sessions
- **Dark UI** — Mobile-first Apple-style dark interface with 3D card flip animations
- **PWA support** — Installable as a standalone app on mobile and desktop, with offline asset caching and auto-update prompts

## Quick Start

### Development

```bash
./dev.sh
```

Builds the frontend and backend, then runs on port 3011 with no auth required.

### Docker

```bash
docker compose up --build
```

Runs on port 8080 with password authentication.

### Manual Build

```bash
cd frontend && npm install && npm run build && cd ..
CGO_ENABLED=1 go build -o ./bin/flash-cards ./cmd/server
APP_PORT=3011 ./bin/flash-cards
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_PASSWORD` | _(none)_ | Password for app access. If unset, no auth required. |
| `APP_PORT` | `8080` | HTTP listen port. |
| `DB_PATH` | `./data/flash-cards.db` | SQLite database file path. |

## Tech Stack

- **Backend**: Go, Gin, mattn/go-sqlite3 (CGO)
- **Frontend**: React 19, TypeScript, Vite, Tailwind CSS v4, TanStack Query, Recharts
- **Database**: SQLite with WAL mode
- **Deployment**: Multi-stage Dockerfile, single binary with embedded frontend
