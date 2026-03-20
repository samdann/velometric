# Velometric

Cycling performance analytics platform. Upload .FIT files, get deep performance insights.

## Project Structure

```
velometric/
  frontend/     # Next.js (App Router) + React + TypeScript + Tailwind
  backend/      # Go REST API
  BRIEF.md      # Full project brief and roadmap
```

## Tech Stack

- **Frontend**: Next.js 16, React, TypeScript, Tailwind CSS 4
- **Backend**: Go, Chi router, PostgreSQL + TimescaleDB
- **Infrastructure**: Docker Compose, Hetzner VPS

## Development

```bash
# Frontend (from frontend/)
npm run dev -- -p 3001   # http://localhost:3001

# Backend (from backend/)
go run cmd/api/main.go   # http://localhost:8081

# Full stack (from root)
docker-compose up
```

## Key Principles

- **Precompute everything** at upload time. Frontend never does heavy computation.
- Dark theme with cycling-inspired colors (orange primary, metric-specific colors)
- Fonts: Instrument Sans (UI), DM Mono (data/numbers)

## API Endpoints (planned)

```
GET  /health              # Health check
POST /api/activities      # Upload FIT file
GET  /api/activities      # List activities
GET  /api/activities/:id  # Activity detail with computed metrics
```


## Engineering Standards

### Frontend
- component reusability is very important. avoid code duplication until it is strongly justified. Search existing components, functions or pattern before implementing.
- If the same UI element or logic exists elsewhere in the codebase, extract it into a shared component/utility and use it in both places. Never write two implementations of the same thing.
- This applies to charts, widgets, helpers, styles, and API calls.
- Always suggest best practices to implement for every new feature or core changes. 

### Tests are not optional
- Every new backend function or behaviour must have corresponding tests.
- When changing existing logic, update existing tests and add new ones to cover the change.
- Run `go test ./...` before considering backend work done.

### Respect the design system
- Always use design token classes (`bg-background-subtle`, `border-border`, `text-foreground-muted`, etc.) — never hardcode colours like `bg-zinc-900` or `border-zinc-800` in product UI.
- Hardcoded zinc/slate values are only acceptable for one-off dev utilities or skeletons that have no design token equivalent.

### Keep localization at the back of your mind
- Site will most likely be translated to different languages at a later time