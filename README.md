# Event-Driven Notification System

Scalable notification system that processes and delivers messages through multiple channels (SMS, Email, Push). Built with Go, PostgreSQL, and RabbitMQ.

## Quick Start

```bash
# One-command setup
docker-compose up -d

# API runs at http://localhost:8080
```

**Config:** Use a `.env` file (repo includes `.env.example`; run `cp .env.example .env` on first setup). Docker Compose loads `.env` automatically; set your webhook UUID there.

## Local Development

**Option A – Run API in Docker too (recommended, avoids port conflicts):**
```bash
docker-compose up -d postgres rabbitmq
docker-compose up --build api
# API: http://localhost:8080
```

**Option B – DB + RabbitMQ in Docker, API on host:**
```bash
docker-compose up -d postgres rabbitmq
# Port 5432 must be free (no local Postgres running)
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/notifications?sslmode=disable"
export RABBITMQ_URL="amqp://guest:guest@localhost:5672/"
go run ./cmd/api
```

RabbitMQ Management UI: http://localhost:15672 (guest/guest)

### "database notifications does not exist" error

This usually happens when **the API connects to a local (non-Docker) Postgres on localhost:5432**: the `notifications` database exists only in the Docker Postgres instance.

- **Quick fix:** Run the API inside Docker (so it uses the Docker Postgres):
  ```bash
  docker-compose up -d postgres rabbitmq && docker-compose up --build api
  ```
- **Alternative:** Create the database on the Postgres instance the API actually connects to:
  ```bash
  psql postgres://postgres:postgres@localhost:5432/postgres -c "CREATE DATABASE notifications;"
  ```
  (If your local user/password differ, use your own; or run `createdb notifications`.)

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | /health | Liveness check |
| GET | /ready | Readiness (DB + RabbitMQ) |
| GET | /metrics | Queue depth, success/failure counts |
| POST | /notifications | Create single notification |
| POST | /notifications/batch | Create batch (up to 1000) |
| GET | /notifications | List with filters (status, channel, from, to, limit, offset) |
| GET | /notifications/{id} | Get by ID |
| GET | /notifications/batch/{batchId} | Get by batch ID |
| DELETE | /notifications/{id} | Cancel pending |

## Example Requests

**Create notification:**
```bash
curl -X POST http://localhost:8080/notifications \
  -H "Content-Type: application/json" \
  -d '{"recipient":"+905551234567","channel":"sms","content":"Hello!","priority":"high"}'
```

**Create batch:**
```bash
curl -X POST http://localhost:8080/notifications/batch \
  -H "Content-Type: application/json" \
  -d '{"notifications":[{"recipient":"+905551234567","channel":"sms","content":"Hi"}]}'
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| PORT | 8080 | API port |
| DATABASE_URL | postgres://... | PostgreSQL connection |
| RABBITMQ_URL | amqp://guest:guest@localhost:5672/ | RabbitMQ connection |
| WEBHOOK_BASE_URL | https://webhook.site | Provider webhook base |
| WEBHOOK_UUID | (required for delivery) | Your webhook.site UUID |
| WORKER_COUNT | 5 | Number of worker goroutines |
| RATE_LIMIT_PER_SEC | 100 | Max messages/sec per channel |

## Run Tests

```bash
go test ./...
```

## Test API & Queue (e2e)

**Method 1 – All in Docker (no port conflicts):**
```bash
docker-compose up -d postgres rabbitmq
docker-compose up --build api
# In another terminal: make test-api
```

**Method 2 – API on host (go run):**
```bash
docker-compose up -d postgres rabbitmq
# Only if 5432/5672 are used by Docker:
go run ./cmd/api
```

**Terminal 2 – API test:**
```bash
make test-api
```
The test verifies that created notifications are processed by the worker as **delivered** (or at least not failed). There should be no **failed** notifications; if you see failures, rebuild the API: `docker-compose up -d --build api`

To try endpoints manually:
```bash
curl http://localhost:8080/health
curl http://localhost:8080/ready
curl -X POST http://localhost:8080/notifications -H "Content-Type: application/json" -d '{"recipient":"+905551234567","channel":"sms","content":"Test"}'
```

## Architecture

**Overview:** REST API (Chi) receives notification requests and persists them in PostgreSQL; notification IDs are pushed to RabbitMQ (per-channel priority queues). Worker goroutines consume from the queue, apply per-channel rate limiting, and send to the provider (webhook.site); delivery status is written back to the database. Retries use exponential backoff (max 3 attempts).

**API documentation (Swagger UI):** When the API is running, open [http://localhost:8080/docs](http://localhost:8080/docs). Raw OpenAPI spec: [docs/openapi.yaml](docs/openapi.yaml) or `GET /openapi.yaml`.

## Database Migrations

Migrations are versioned under `internal/storage/migrations/` (e.g. `001_create_notifications.up.sql`, `002_add_source_column.up.sql`). They run automatically on API startup (embed, applied in order). No separate migration command required for normal runs.

---

## Assessment requirements checklist

| Requirement | Status |
|-------------|--------|
| 1. Source code in GitHub/GitLab with clean commit history | Push this repo and keep a linear, meaningful commit history. |
| 2. README: setup, architecture overview, API examples | ✅ Quick Start, Local Development, Architecture (above), Example Requests, Environment Variables. |
| 3. Docker Compose one-command setup | ✅ `docker-compose up -d` (or `docker-compose up --build api` for full stack with API). |
| 4. API documentation: Swagger/OpenAPI | ✅ OpenAPI 3 spec in `docs/openapi.yaml`; Swagger UI at `GET /docs` when API is running. |
| 5. Database migrations: versioned schema | ✅ `internal/storage/migrations/` (001, 002); auto-applied on startup. |
| 6. Test suite: single command | ✅ `go test ./...` or `make test`. |
