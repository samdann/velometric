# Velometric — Project Brief

## Vision
Custom cycling analytics dashboard. Upload .FIT files (Garmin), get deep performance insights that Strava doesn't offer. Eventually: the biggest competitor to Strava.

## Architecture

```
Frontend: Next.js (App Router) + React + TypeScript + Tailwind
Backend:  Go (REST API)
Database: PostgreSQL + TimescaleDB (time-series data)
Cache:    Redis
Storage:  S3-compatible (Hetzner Object Storage or Cloudflare R2)
Infra:    Hetzner VPS, Docker Compose, future migration to AWS/GCP
```

## Monorepo Structure

```
velometric/
  frontend/           # Next.js app
  backend/            # Go API
  BRIEF.md            # This file
  CLAUDE.md           # AI assistant context
  docker-compose.yml  # Local dev orchestration
```

## Data Model

Activities contain ~9,000+ data points per ride with fields:
- GPS (lat, lon), altitude, distance, speed
- Power, heart rate, cadence, temperature
- Left/right balance, torque effectiveness, pedal smoothness
- Gear change events

Key principle: **precompute everything at upload time**. Frontend never does heavy computation.

## Database Schema (planned)

```sql
users (id, email, name, ftp, max_hr, weight, created_at)

activities (id, user_id, sport, start_time,
  distance, duration, elevation_gain, avg_power, np, tss, intensity_factor,
  avg_hr, max_hr, calories, fit_file_url, created_at)

activity_records — TimescaleDB hypertable
  (activity_id, timestamp, lat, lon, altitude, distance,
   power, heart_rate, cadence, speed, temperature,
   l_torque_eff, r_torque_eff, l_pedal_smooth, r_pedal_smooth,
   left_right_balance)

activity_laps (id, activity_id, lap_number, start_time, duration, distance,
  avg_power, np, avg_hr, max_hr, ascent, descent, ...)

activity_power_curve (activity_id, duration_seconds, best_power)

activity_events (activity_id, timestamp, event_type, data_json)
```

## Sprint 1 — Foundation

### Step 1 ✅ — Next.js scaffold
Created with: TypeScript, Tailwind, App Router, src dir, ESLint

### Step 2 ✅ — Project structure & core layout
- Set up folder structure: components/, lib/, types/, hooks/
- Create the app shell layout (collapsible sidebar)
- Set up theme: dark theme with cycling-inspired palette
- Configure fonts (DM Mono for data, Instrument Sans for UI)
- Set up routing: / (dashboard), /activities, /activities/[id], /upload, /settings
- Create placeholder pages for each route

### Step 3 ✅ — Go backend scaffolding
- Go project with standard layout (cmd/, internal/, pkg/)
- HTTP router (Chi)
- Health check endpoint
- CORS config for Next.js frontend
- Environment config (.env)

### Step 4 — Database schema & migrations
- PostgreSQL + TimescaleDB setup
- Migration tool (golang-migrate)
- Initial migration with schema above
- Seed script for development

### Step 5 — FIT file upload flow
- Frontend: drag & drop upload component
- Backend: receive file, store raw in object storage
- Backend: parse .FIT file (use go-fit library)
- Backend: compute derived metrics (NP, TSS, IF, power curve, zones)
- Backend: insert parsed data into Postgres
- Frontend: redirect to activity detail page

### Step 6 ✅ — Docker Compose
- Dockerfile for Next.js frontend
- Dockerfile for Go backend
- docker-compose.yml with: frontend, backend, postgres (TimescaleDB), redis
- Works locally, deployable to Hetzner VPS

## Dashboard Features (from prototype)

7 tabs built in prototype:
1. **Overview** — summary cards, elevation profile, zone summaries
2. **Timeline** — toggleable metric overlays (power, HR, speed, cadence, temp) with hover tooltips
3. **Zones** — power and HR zone breakdowns with histograms
4. **Power Lab** — power curve, variability index, cardiac efficiency/decoupling
5. **Laps** — table + bar chart comparison, climb detection
6. **Map** — GPS route colored by any metric
7. **Pedaling** — L/R balance, torque effectiveness, pedal smoothness

## Future Features
- Synced crosshair (hover timeline ↔ map highlight)
- Gradient analysis (real-time gradient from GPS, gradient vs power scatter)
- Match burning analysis (auto-detect hard efforts, recovery visualization)
- Gear usage overlay (from gear change events in .FIT)
- Multi-file support (client-side .FIT parsing)
- Ride vs ride comparison
- Personal records & trend tracking
- Strava import via OAuth
- Social features, segments, leaderboards

## Tech Preferences
- Owner strongest in Kotlin, learning new language (Go chosen for backend)
- Prioritize compatibility and easy migration between technologies
- Hetzner VPS for deployment (simple, cheap)
- Docker everything for portability