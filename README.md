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
docker run --rm -p 8080:8080 --env-file .env.example desent-api:local
```

Quick check:

```bash
curl http://127.0.0.1:8080/ping
```

## Endpoint

- `GET /ping` -> `{"success":true}`

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
