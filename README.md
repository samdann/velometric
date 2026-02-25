# Velometric

Cycling performance analytics platform. Upload FIT files, get deep performance insights.

## Quick Start

### Prerequisites
- Node.js 22+
- Go 1.23+
- Docker & Docker Compose
- PostgreSQL (or use Docker)

### Using Docker (Recommended)

```bash
# Start everything
docker-compose up -d

# View logs
docker-compose logs -f

# Stop everything
docker-compose down
```

Services:
- Frontend: http://localhost:3001
- Backend: http://localhost:8081
- Database: localhost:5432
- Redis: localhost:6379

### Local Development

**Backend:**
```bash
cd backend

# Start server
make run

# Or with Go directly
go run cmd/api/main.go
```

**Frontend:**
```bash
cd frontend

# Install dependencies (first time)
npm install

# Start dev server
npm run dev -- -p 3001

# Build for production
npm run build
```

## Useful Commands

### Backend (from `backend/` directory)

| Command | Description |
|---------|-------------|
| `make run` | Start the API server |
| `make build` | Build binary to `bin/api` |
| `make test` | Run tests |
| `make migrate-up` | Run database migrations |
| `make migrate-down` | Rollback last migration |
| `make migrate-reset` | Rollback all migrations |
| `make seed` | Seed database with sample data |
| `make reset-activities` | Delete all activities (keeps users/zones) |

### Frontend (from `frontend/` directory)

| Command | Description |
|---------|-------------|
| `npm run dev -- -p 3001` | Start dev server on port 3001 |
| `npm run build` | Production build |
| `npm run start` | Start production server |
| `npm run lint` | Run ESLint |

### Docker Compose (from root directory)

| Command | Description |
|---------|-------------|
| `docker-compose up -d` | Start all services in background |
| `docker-compose up --build` | Rebuild and start |
| `docker-compose down` | Stop all services |
| `docker-compose logs -f backend` | Follow backend logs |
| `docker-compose exec db psql -U postgres -d velometric` | Connect to database |

### Database Only

| Command | Description |
|---------|-------------|
| `docker-compose up db -d` | Start database only |
| `docker-compose stop db` | Stop database (preserves data) |
| `docker-compose down db` | Stop and remove container (preserves volume) |
| `docker-compose down -v` | Stop and delete all data (destructive) |

### Database

```bash
# Connect to local database
psql -U postgres -d velometric

# Connect via Docker
docker-compose exec db psql -U postgres -d velometric

# Run raw SQL
docker-compose exec db psql -U postgres -d velometric -c "SELECT * FROM activities;"
```

## Project Structure

```
velometric/
├── frontend/          # Next.js app
│   ├── src/
│   │   ├── app/       # Pages (App Router)
│   │   ├── components/
│   │   ├── hooks/
│   │   ├── lib/       # Utilities, API client
│   │   └── types/
│   └── Dockerfile
├── backend/           # Go API
│   ├── cmd/api/       # Entry point
│   ├── internal/
│   │   ├── config/
│   │   ├── database/
│   │   ├── fitparser/ # FIT file parsing
│   │   ├── handler/   # HTTP handlers
│   │   ├── model/
│   │   ├── repository/
│   │   └── service/   # Business logic
│   ├── migrations/
│   ├── scripts/
│   ├── Makefile
│   └── Dockerfile
├── docker-compose.yml
├── BRIEF.md           # Project roadmap
├── BACKLOG.md         # Bugs & improvements
└── CLAUDE.md          # AI assistant context
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/api/activities` | List activities |
| POST | `/api/activities` | Upload FIT file |
| GET | `/api/activities/:id` | Get activity details |
| GET | `/api/activities/:id/records` | Get time-series data |
| GET | `/api/activities/:id/power-curve` | Get power curve |

## Environment Variables

**Backend** (`backend/.env`):
```
PORT=8081
FRONTEND_URL=http://localhost:3001
DATABASE_URL=postgres://postgres:postgres@localhost:5432/velometric?sslmode=disable
REDIS_URL=redis://localhost:6379
```

**Frontend** (`frontend/.env.local`):
```
NEXT_PUBLIC_API_URL=http://localhost:8081
```
