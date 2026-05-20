# Backend Structure

The backend is a Go modular monolith.

- `cmd/api` runs the HTTP API
- `cmd/worker` runs async consumers and background jobs
- `cmd/migrate` is reserved for database migrations and seed logic
- `internal/<domain>` contains bounded contexts such as `auth`, `content`, and `consultations`
- `internal/infrastructure` holds infrastructure wiring like config, database, redis, metrics, and logging
- `internal/events` and `internal/queue` support event-driven flows
- `migrations` stores SQL migration files
- `tests/integration` is for end-to-end and DB-backed tests
