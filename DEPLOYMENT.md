# Tribunal Deployment Guide

## Quick Start (Docker Compose)

### Prerequisites
- Docker & Docker Compose installed
- `.env` file with required API keys (see `.env.example`)

### Environment Setup

Create a `.env` file in the project root:

```bash
# Required
TRIBUNAL_API_KEY=your_secure_api_key_here

# Optional (defaults provided)
GITHUB_TOKEN=
OPENROUTER_API_KEY=
NEXT_PUBLIC_API_URL=http://localhost:8080
POSTGRES_USER=tribunal
POSTGRES_PASSWORD=tribunal_password_dev
ALLOWED_ORIGINS=http://localhost:3000
```

### Start the Stack

```bash
# Build and run all services (database, backend, frontend)
docker compose up -d

# Check service status
docker compose ps

# View logs
docker compose logs -f interceptor    # Backend logs
docker compose logs -f dashboard      # Frontend logs
docker compose logs -f db             # Database logs
```

### Access the Application

- **Dashboard:** http://localhost:3000
- **API Health:** http://localhost:8080/health
- **API Docs:** http://localhost:8080/api/v1/analyze (POST endpoint)

### Stop Services

```bash
docker compose down
```

### Reset Database

```bash
docker compose down -v  # Remove volumes too
docker compose up -d
```

## Local Development (No Docker)

### Backend Setup

```bash
cd services/go-interceptor
export TRIBUNAL_API_KEY=test_key
export PORT=8080
go run .
```

### Frontend Setup

```bash
cd dashboard
npm install
npm run dev  # Runs on http://localhost:3000
```

### Database Setup (Local PostgreSQL)

```bash
# Create user and database
createuser -P tribunal  # password: tribunal_password_dev
createdb -U tribunal tribunal_db

# Initialize schema
psql -U tribunal -d tribunal_db < schema/postgres.sql
psql -U tribunal -d tribunal_db < schema/seed.sql
```

## Architecture Overview

### Backend (Go)
- **Port:** 8080
- **Health Check:** GET `/health` → Returns goroutines, memory usage, service status
- **Webhook Endpoints:**
  - POST `/webhook/github` – GitHub webhook receiver
  - POST `/webhook/gitlab` – GitLab webhook receiver
  - POST `/webhook/bitbucket` – Bitbucket webhook receiver
- **Analysis:** POST `/api/v1/analyze` – Direct code analysis
- **Security:** Bearer token auth + webhook signature verification

### Frontend (Next.js)
- **Port:** 3000
- **Pages:**
  - `/` – Risk command center (overview)
  - `/analytics` – Enterprise analytics dashboard
- **Features:** Real-time data, demo fallbacks, CSV/JSON/HTML exports

### Database (PostgreSQL)
- **Port:** 5432
- **Tables:** audit_logs, policies, api_keys, webhook_events
- **Initialization:** Auto-loaded from `/schema/postgres.sql` on first run

## Troubleshooting

### Backend won't start
```bash
# Check port 8080 is free
lsof -i :8080

# Verify environment variables
echo $TRIBUNAL_API_KEY
echo $DATABASE_URL
```

### Frontend stuck in restart loop
```bash
cd dashboard
rm -rf .next .turbo node_modules
npm install
npm run dev
```

### Database connection error
```bash
# Verify Postgres is running
docker compose logs db

# Reset database
docker compose down -v
docker compose up -d db
docker compose up -d interceptor
```

### Demo data appearing instead of live data
- Backend API unreachable (check health endpoint)
- API key invalid or expired
- Check frontend logs: `docker compose logs dashboard`

## Performance Monitoring

### Health Endpoint Response (Example)

```json
{
  "status": "ok",
  "service": "go-interceptor",
  "goroutines": 12,
  "memory_mb": 45
}
```

### Key Metrics to Monitor
- **Goroutines:** Should stabilize after startup (watch for leaks)
- **Memory:** Typical usage 40–200 MB depending on load
- **Latency:** `/analyze` endpoint should respond in < 2 seconds for most payloads

## Production Deployment

### Recommended Setup
1. Use managed PostgreSQL (AWS RDS, CloudSQL, etc.)
2. Disable demo fallbacks in production
3. Set `TRIBUNAL_API_KEY` to a strong, randomly-generated value
4. Use environment-specific `.env` files (`.env.prod`)
5. Configure webhook secret verification for all integrations

### Environment Variables (Production)
```bash
DATABASE_URL=postgres://user:password@prod-db.example.com:5432/tribunal
TRIBUNAL_API_KEY=<strong-random-key>
GITHUB_TOKEN=<github-app-token>
OPENROUTER_API_KEY=<openrouter-api-key>
ALLOWED_ORIGINS=https://yourdomain.com
```

### Scaling Considerations
- Backend is stateless (horizontal scaling compatible)
- Use load balancer for multiple backend instances
- PostgreSQL can be upgraded to HA setup with replicas
- Cache frontend static assets with CDN

## Testing

### Run Backend Tests
```bash
cd services/go-interceptor
go test ./...
```

### Run Frontend Build
```bash
cd dashboard
npm run build
```

### Integration Test
```bash
# With stack running, test webhook receiver
curl -X POST http://localhost:8080/webhook/github \
  -H "Authorization: Bearer your_api_key" \
  -H "Content-Type: application/json" \
  -d @sample-analyze-payload.json
```

## Support & Debugging

Enable verbose logging:
```bash
# Backend
LOGLEVEL=debug docker compose up -d interceptor

# View all logs
docker compose logs --tail=100 -f
```

Check service health:
```bash
curl http://localhost:8080/health
curl http://localhost:3000
```

---

**Last Updated:** 19 April 2026  
**Status:** Production-Ready ✅
