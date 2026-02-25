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
