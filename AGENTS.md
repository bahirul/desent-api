# Go API Standard (Repository Guidelines)

Last updated: 2026-02-28 11:24

Use this document as the default standard for new Go backend projects that follow this architecture.

## 1) Project Structure & Module Organization

```text
cmd/
  api/                 # Main HTTP server entrypoint
  <optional-cli>/      # Standalone utility CLIs (seeders, bootstrap, admin tools)

configs/               # Environment-driven config loading (Load, Getenv)

internal/
  handlers/            # HTTP transport: parse request, auth, call usecase, write response
  usecases/            # Business logic: one file per operation, returns *models.UsecaseResult
  repositories/        # Persistence/data access (MongoDB in this standard)
  middlewares/         # HTTP middlewares (recover, logger, CORS, etc.)
  responses/           # Response helper utilities
  models/              # Domain, DTO/view, request/response payload models
  utils/               # Shared helpers (JWT parsing, hashing, result handler)

docs/
  api/                 # API contracts grouped by resource
  architecture.md      # Architecture, flow, dependency boundaries
  collections.md       # MongoDB schema and indexes
  business-rules.md    # Rule catalog + owner file mapping + threshold values
  README.md            # Documentation index
```

Example pairing:
- `internal/usecases/create_vehicle_usecase.go`
- `internal/handlers/vehicle_handler.go`

## 2) Architecture & Boundaries

Core flow: **handlers -> usecases -> repositories**.

- **Handlers**
  - Thin transport layer only.
  - Decode/validate request shape, enforce endpoint auth, call one usecase, return via `utils.HandleUsecaseResult`.
  - Must not contain business rules.
- **Usecases**
  - Own business logic, orchestration, validation, and policy decisions.
  - Return `*models.UsecaseResult`.
- **Repositories**
  - Only persistence logic and query construction.
  - No business policy.

Dependency direction:
- `handlers` can import: `usecases`, `responses`, `utils`, `configs`, `models`.
- `usecases` can import: `repositories`, `models`, `utils`, `configs`.
- `repositories` can import: `models` + DB driver packages.

Boundary enforcement rules:
- Enforce layer boundaries with tooling (for example: `depguard`, custom lint rule, or architecture tests).
- CI must fail if forbidden imports violate `handlers -> usecases -> repositories` direction.

## 3) Response & Error Contract

Domain error taxonomy (default mapping):
- `400 Bad Request`: malformed request shape or invalid primitive parsing at handler boundary.
- `401 Unauthorized`: missing/invalid access token.
- `403 Forbidden`: authenticated caller lacks required role/policy.
- `404 Not Found`: resource does not exist for requested scope.
- `409 Conflict`: duplicate key/state transition conflict/business uniqueness violation.
- `422 Unprocessable Entity`: domain/business validation errors with field-level details.
- `500 Internal Server Error`: unexpected/unclassified server errors.

Rules:
- Usecases should return stable, predictable `HttpCode` + `Message` pairs for equivalent failures.
- Handlers must avoid remapping usecase domain failures inconsistently across endpoints.
- For non-validation domain failures, use a canonical response payload with stable `error_code` and `message` fields.

## 4) Build, Test, and Development Commands

```bash
make air-install
make dev
go tool air -c .air.toml
go run ./cmd/api
make build
make test
make test-cover
make vet
make fmt
make tidy
go test -cover ./...
go vet ./...
go fmt ./...
```

### CI Enforcement (Required)

- Every PR must pass CI checks for:
  - `go test ./...`
  - `go vet ./...`
  - `go test -cover ./...` (or project-defined coverage target)
  - `golangci-lint run` (default required lint baseline)
- If `golangci-lint` is intentionally not used, `staticcheck ./...` is required and the exception must be documented in `docs/architecture.md`.
- Standards in this file are enforceable only when mapped to CI checks; keep CI config aligned with this document.

## 5) Documentation Rules

- Keep `docs/architecture.md` aligned with real routes/modules.
- Keep `docs/README.md` as the single index of all docs.
- Keep one API file per resource in `docs/api/<resource>.md`.
- If present, keep `docs/collections.md` aligned with models/repositories (fields, optionality, indexes, TTL).
- If present, keep `docs/business-rules.md` aligned with implemented behavior in usecases/repositories/handlers.
- If `docs/business-rules.md` exists, any PR that changes business behavior (validation, thresholds, transition logic, authorization policy, or default filtering) must update it in the same change.

## 6) Coding Style & Naming

- Always run `go fmt ./...` before commit.
- Prefer small files per usecase/handler operation.
- Package names are short/lowercase/plural when appropriate.
- Exported identifiers: `PascalCase`; internal: `camelCase`.
- Keep comments concise and intent-focused.
- For handler and usecase operations, add ordered flow comments for non-trivial paths:
  - Use numeric markers like `// 1)`, `// 2)`, `// 3)` to show execution order.
  - Keep each flow comment short (one line) and explain intent, not syntax.
  - Handler flow comments should describe: auth/guard -> request parsing -> usecase call -> response mapping.
  - Usecase flow comments should describe: validation -> dependency/repository setup -> business rules/orchestration -> persistence -> result mapping.
  - Only use flow comments where they improve readability; avoid comment noise in simple code.

## 7) Testing Standard

- Tests live alongside code in `*_test.go`.
- Prefer table-driven tests for validators and mapping logic.
- Handler tests should cover:
  - auth guard behavior (authorized vs unauthorized roles),
  - validation path behavior (`400`/`422`),
  - deterministic branch behavior.
- Usecase tests should cover:
  - success path,
  - domain validation failures,
  - repository error mapping to HTTP codes/messages.
- Repository tests should cover:
  - query/filter correctness for key read paths,
  - index assumptions that affect behavior/performance (including uniqueness),
  - error mapping behavior for duplicate key/not found/write failures where applicable.
- API contract tests should verify endpoint responses stay aligned with `docs/api/<resource>.md`.
- Minimum gate before merge: all required CI checks in `CI Enforcement (Required)` pass.

## 8) API & Data Design Rules

- Keep request parsing and business validation separated (handler vs usecase).
- For optional request fields, define and test explicit default behavior.
- Do not expose sensitive fields in views/responses (example: password hash, refresh token hash).
- Preserve backward compatibility unless intentionally versioned or announced.

## 9) Commit & PR Guidelines

Use Conventional Commits:
- `feat: add developer-only user detail endpoint`
- `fix: default vehicle active status to false on create`
- `chore: sync docs for jwt sessions endpoint`

PR checklist:
- Code + tests + docs updated together.
- No secret values committed.
- Any behavior change explicitly called out.
- After large refactors, explicitly clean dead code and stale assets before merge:
  - remove unused functions/types,
  - remove unused folders/files,
  - remove unreachable routes or legacy wiring no longer referenced.
- If this `AGENTS.md` file is changed by an agent, update the `Last updated` date and time at the top of the file in the same change.

## 10) Security & Configuration

- Never commit secrets; keep `.env.example` updated.
- Validate all external inputs at boundaries.
- Use strong `JWT_SECRET` and explicit token TTLs.
- Require explicit HTTP server timeouts (`ReadTimeout`, `WriteTimeout`, `IdleTimeout`, `ReadHeaderTimeout`).
- Define request body size limits for write endpoints.
- Apply explicit CORS policy (origins, methods, headers); never rely on permissive wildcard defaults for authenticated APIs.
- Add rate limiting on public/auth-sensitive endpoints (login, token refresh, password reset, OTP, etc.).

### Logging Configuration

- Log files are stored in the directory specified by `LOGS_DIR` (default: `logs`).
- Two log files are created daily with date stamps: `error.YYYY-MM-DD.log` and `http.YYYY-MM-DD.log`.
- Auto-delete old log files based on `LOG_RETENTION_DAYS` (default: 7 days).
- Enable/disable logging via `HTTP_LOG_CONSOLE_ENABLED` and `HTTP_LOG_FILE_ENABLED`.
- Use structured logs with a request/correlation ID for traceability.
- Never log sensitive values (access/refresh tokens, passwords, password hashes, auth headers, private keys, raw secret env values, high-risk PII).
