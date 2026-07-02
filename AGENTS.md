# AGENTS.md

## Project Overview

Hot Contracts Dashboard (热门合约看板) — a Binance futures data aggregator with a Go backend and Vue 3 frontend. The backend fetches market data concurrently from Binance APIs, derives high-level trading metrics (relative strength, taker buy ratio, funding rate, premium rate), caches results in-memory, and serves them via a single REST endpoint. The frontend renders a high-density dark-themed data table with real-time polling.

## Architecture

Two independent projects with no shared tooling, no monorepo setup, and no root-level package manager config:

- **`backend/`** — Go (Gin) HTTP server on port **8082**
- **`frontend/`** — Vite dev server on port **5180**, Vue 3 + TypeScript + Tailwind CSS 4 + Element Plus

API contract: `GET /api/v1/contracts/hot?sort_by=<field>&order=<ascending|descending>`

The frontend hardcodes `http://localhost:8082/api/v1/contracts/hot` as the API URL (`src/App.vue:232`). Both servers must run simultaneously for the app to work.

## Commands

### Backend (Go)

```bash
cd backend
go run main.go          # Start server on :8082
go build -o server.exe  # Build binary (server.exe already exists in repo)
```

No test files exist. No linter configured.

### Frontend (Vue/Vite)

```bash
cd frontend
npm run dev       # Vite dev server on :5180
npm run build     # vue-tsc -b && vite build
npm run preview   # Preview production build
```

No test files exist. No eslint or prettier configured. TypeScript is strict: `noUnusedLocals`, `noUnusedParameters`, `erasableSyntaxOnly`.

## Key Gotchas

- **Backend caching**: Data is cached for 60s (`go-cache`). Sorting happens *after* cache read on a shallow copy — the cache is never mutated by sort requests.
- **Hardcoded funding intervals**: `backend/service/binance.go:188` has 4 symbols (`STORJUSDT`, `ARKMUSDT`, `TRB-USDT`, `GASUSDT`) with 4-hour funding intervals hardcoded. Adding new symbols with non-standard intervals requires updating this list.
- **In-memory volume filter**: Only contracts with 24h quote volume > 1M USDT are included (`binance.go:143`). This threshold is hardcoded.
- **Tailwind CSS 4**: Uses the new `@tailwindcss/vite` plugin and `@import "tailwindcss"` syntax — not the v3 `@tailwind` directives. Config in `tailwind.config.js` and `postcss.config.js` both exist but the Vite plugin takes precedence.
- **Element Plus dark theme**: Custom CSS variables override Element Plus defaults in `src/App.vue:334-344`. The `<html>` tag has `class="dark"` in `index.html`.
- **Frontend polling**: 10-second interval via `useIntervalFn`. Sorting triggers an immediate refetch with new query params.
- **No CI/CD, no pre-commit hooks, no Makefile.**

## Directory Structure

```
heyue/
├── AGENTS.md
├── .cursor/plans/          # Cursor plan files (design docs)
├── backend/
│   ├── main.go             # Entry point, Gin router, CORS, handler
│   ├── model/
│   │   ├── contract.go     # HotContract struct (API response shape)
│   │   └── raw.go          # Binance raw response structs
│   └── service/
│       ├── binance.go      # API client, concurrent fetch, metric derivation
│       └── sort.go         # Server-side sorting logic
└── frontend/
    ├── src/
    │   ├── App.vue         # Main component — table, formatting, polling (all-in-one)
    │   ├── main.ts         # App bootstrap, Element Plus plugin
    │   ├── style.css       # Tailwind import + base styles
    │   └── components/     # Only HelloWorld.vue (unused scaffold)
    ├── vite.config.ts      # Dev server on :5180
    ├── tailwind.config.js
    └── postcss.config.js
```

## UI Language

The entire user-facing interface is in **Chinese (Simplified)**. Column labels, tooltips, and status text are hardcoded in Chinese in `App.vue`.
