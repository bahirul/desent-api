# desent-api

Go API service scaffold for the desent.io coding quest (interview test), using `chi`, environment-driven config, structured logging, and Air hot reload for local development.

## Prerequisites

- Go `1.25+`

## Development

Install Air once (optional if `go tool air` is already available in your environment cache):

```bash
make air-install
```

Run with hot reload:

```bash
make dev
```

Run without hot reload:

```bash
go run ./cmd/api
```

## Commands

```bash
make build
make test
make test-cover
make vet
make fmt
make tidy
```

## Docker

Build image:

```bash
docker build -t desent-api:local .
```

Run container:

```bash
mkdir -p ./data
docker run --rm -p 18080:8080 --env-file .env.example desent-api:local
```

Run container with persistent DB on host:

```bash
mkdir -p ./data
docker run --rm -p 18080:8080 --env-file .env.example -e DB_DSN=file:/app/data/books.db -v "$(pwd)/data:/app/data" desent-api:local
```

Quick check:

```bash
curl http://127.0.0.1:18080/ping
```

## Endpoints

- `GET /ping` -> `{"success":true}`
- `POST /echo` -> echoes the exact JSON body
- `POST /auth/token` -> returns JWT token for `{ "username":"admin", "password":"password" }`
- `POST /books` -> creates a book
- `GET /books` -> returns all books (requires `Authorization: Bearer <token>`)
- `GET /books/:id` -> returns one book
- `PUT /books/:id` -> updates one book
- `DELETE /books/:id` -> deletes one book

Auth example:

```bash
TOKEN=$(curl -sS -X POST http://127.0.0.1:18080/auth/token \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"password"}' | jq -r .token)

curl -sS http://127.0.0.1:18080/books -H "Authorization: Bearer $TOKEN"
```

## Environment Variables

Server:
- `HTTP_ADDR` (default: `:8080`)
- `HTTP_READ_TIMEOUT_SECONDS` (default: `10`)
- `HTTP_READ_HEADER_TIMEOUT_SECONDS` (default: `5`)
- `HTTP_WRITE_TIMEOUT_SECONDS` (default: `15`)
- `HTTP_IDLE_TIMEOUT_SECONDS` (default: `60`)

Logging:
- `LOGS_DIR` (default: `logs`)
- `LOG_RETENTION_DAYS` (default: `7`)
- `HTTP_LOG_CONSOLE_ENABLED` (default: `true`)
- `HTTP_LOG_FILE_ENABLED` (default: `true`)

Database:
- `DB_DRIVER` (default: `sqlite`)
- `DB_DSN` (default: `file:/tmp/books.db`)

Auth:
- `JWT_SECRET` (default: `dev-secret-change-me`)
- `JWT_TTL_SECONDS` (default: `3600`)
