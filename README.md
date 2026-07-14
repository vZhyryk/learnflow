# LearnFlow

![Go](https://img.shields.io/badge/Go-1.26.3-00ADD8?style=flat&logo=go&logoColor=white)
![TypeScript](https://img.shields.io/badge/TypeScript-6.0-3178C6?style=flat&logo=typescript&logoColor=white)
![Vue 3](https://img.shields.io/badge/Vue-3.5-4FC08D?style=flat&logo=vue.js&logoColor=white)
![Nuxt](https://img.shields.io/badge/Nuxt-4.4-00DC82?style=flat&logo=nuxt.js&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-17-4169E1?style=flat&logo=postgresql&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-7-DC382D?style=flat&logo=redis&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=flat&logo=docker&logoColor=white)
![Status](https://img.shields.io/badge/Status-In%20Development-orange?style=flat)

> A production-like learning SaaS platform — built to practice clean architecture, event-driven systems, and full-stack engineering at production depth.

---

## Overview

LearnFlow is a **personal engineering project** simulating a real SaaS platform for online learning content and 1:1 consultations. It is deliberately scoped and architected as a production system — not a tutorial app — to exercise patterns that matter in professional backend and full-stack work.

The platform covers the full lifecycle of a learning product: course catalog, video/book/presentation content delivery, consultation booking with availability management, payment processing (Stripe + Przelewy24), notifications, analytics, admin tooling, and P&L reporting.

**What this project is for:**

- Practicing clean architecture and strict dependency rules in Go
- Building event-driven workflows with a transactional outbox pattern
- Working with a real relational schema (32 tables) designed for production constraints
- Integrating real external services (Stripe, Resend, Cloudflare R2, Prometheus/Grafana)
- SSR frontend with Nuxt 4, Pinia, and a strict layered architecture

**Current status:** Phase 2 of 9 — Auth & Users complete (JWT sessions, email flows, event-driven workers, full unit + real-Postgres integration test coverage).

---

## Architecture

LearnFlow is a **modular monolith** following Clean Architecture with a strict inward dependency rule. Each domain module is self-contained and follows the same four-layer structure:

```
HTTP Request
      │
      ▼
┌─────────────────────────────────────────────────┐
│  Transport Layer  (net/http, no frameworks)      │
│  Handlers · Middleware · Route registration      │
└────────────────────────┬────────────────────────┘
                         │  calls
                         ▼
┌─────────────────────────────────────────────────┐
│  Service Layer                                   │
│  Business logic · Validation · Event emission   │
└────────────────────────┬────────────────────────┘
                         │  calls
                         ▼
┌─────────────────────────────────────────────────┐
│  Repository Layer                                │
│  Raw SQL (database/sql) · No ORM                │
└────────────────────────┬────────────────────────┘
                         │  reads/writes
                         ▼
┌─────────────────────────────────────────────────┐
│  Domain Layer                                    │
│  Types · Interfaces · Constants · No imports     │
└────────────────────────┬────────────────────────┘
                         │
              ┌──────────┴──────────┐
              ▼                     ▼
        PostgreSQL 17           Redis 7
```

**Dependency rule:** each layer only imports inward. `domain/` imports nothing from the project. `repository/` imports only `domain/`. `service/` imports `domain/` + `repository/`. `transport/` imports `service/`.

**Entrypoints:**

| Binary    | Path           | Role                       |
| --------- | -------------- | -------------------------- |
| `api`     | `cmd/api/`     | HTTP API server            |
| `worker`  | `cmd/worker/`  | Async event processor      |
| `migrate` | `cmd/migrate/` | Database migrations runner |

---

## Tech Stack

| Layer                | Technology                                    | Notes                                                 |
| -------------------- | --------------------------------------------- | ----------------------------------------------------- |
| **Language**         | Go 1.26.3                                     | Stdlib only — `net/http`, `database/sql`, `context`   |
| **Frontend**         | Nuxt 4 + Vue 3.5 + TypeScript 6               | SSR, Pinia, Tailwind CSS 4                            |
| **Database**         | PostgreSQL 17                                 | UUID PKs, soft deletes, bigint cents, partial indexes |
| **Cache / Queue**    | Redis 7                                       | Event queue (BLPop), outbox relay                     |
| **Migrations**       | golang-migrate                                | Sequential `.up.sql` / `.down.sql`                    |
| **Object Storage**   | Cloudflare R2 (prod) · MinIO (local)          | S3-compatible, swappable via env                      |
| **Email**            | Resend (prod) · stub (local)                  | Interface-driven, no lock-in                          |
| **Payments**         | Stripe + Przelewy24                           | Card + BLIK/bank transfer (PL market)                 |
| **Observability**    | Prometheus + Grafana                          | `pg_stat_statements`, custom metrics                  |
| **Testing**          | GoConvey · Vitest · Playwright                | Unit + real-Postgres/Redis integration + e2e          |
| **Linting**          | golangci-lint (15+ linters) · oxlint + ESLint | Strict configs, CI-enforced                           |
| **Containerization** | Docker Compose                                | Dev, observability, and full-stack profiles           |

---

## Domain Modules

The backend is organized into 16 bounded contexts under `backend/internal/`:

| Module            | Responsibility                                                                                    |
| ----------------- | ------------------------------------------------------------------------------------------------- |
| `auth/`           | JWT sessions (refresh rotation, reuse detection, brute-force lock), email verification, password reset, email change, account recovery |
| `users/`          | User accounts, profiles (avatar, phone), notification preferences                                 |
| `courses/`        | Course catalog (draft/published/archived), SEO metadata, landing pages, reviews & ratings         |
| `content/`        | Learning content items (video, book, presentation), completion tracking, individual access grants |
| `consultations/`  | 1:1 session booking, availability slot management, rescheduling                                   |
| `payments/`       | Stripe + Przelewy24 integration, course/content access, gift coupons, admin refunds               |
| `notifications/`  | Delivery pipeline, reminder scheduling (24h + 3h), re-engagement (14-day inactivity)              |
| `analytics/`      | UTM attribution, user activity tracking                                                           |
| `pnl/`            | P&L reporting — monthly + per-course, expenses, campaign costs, ROAS                              |
| `admin/`          | User management (deactivate/delete/role), announcements, full audit trail                         |
| `notes/`          | Personal notes linked to course or content item, CRUD, full-text search                           |
| `articles/`       | Admin-managed articles (draft/published), SEO metadata, soft delete, public SSR pages             |
| `history/`        | Activity log — user action history (`activity_log` table)                                         |
| `app/`            | Shared app-level logic and bootstrap                                                              |
| `shared/`         | Cross-cutting: JWT middleware, error types (`ErrNotFound`, `ErrForbidden`), pagination            |
| `infrastructure/` | DB, Redis, logger, config, storage, email, sanitizer, metrics                                     |

---

## Key Engineering Decisions

### No ORM — Raw SQL Only

All database access uses `database/sql` with prepared statements and named query constants. This is a deliberate constraint: it forces explicit understanding of query plans, index usage, and transaction boundaries — skills that ORM abstractions obscure.

### No HTTP Framework

Only Go's stdlib `net/http`. Middleware is written as `http.Handler` wrappers. Route matching uses the standard mux. This keeps the transport layer thin, testable, and free of framework coupling.

### Strict Logging Policy

Logging only happens in handlers and workers — never in service or repository layers. This keeps business logic pure, avoids log noise from internal retries, and makes log context (request ID, user, endpoint) always available at the boundary where it belongs.

### Architecture Decision Records

Key infrastructure and security choices are documented as ADRs in `docs/adr/`:

| ADR                                                   | Decision                                      | Rationale                                                                                                            |
| ------------------------------------------------------ | ---------------------------------------------- | ------------------------------------------------------------------------------------------------------------------- |
| [001](docs/adr/001-storage-provider.md)                | Cloudflare R2 + MinIO                          | S3-compatible API — no vendor lock-in; local dev parity                                                              |
| [002](docs/adr/002-email-provider.md)                  | Resend + stub provider                         | Interface-driven — swap provider without touching business logic                                                    |
| [003](docs/adr/003-payment-providers.md)               | Stripe + Przelewy24                            | Card + BLIK/bank transfer for the Polish market; outbox for webhook reliability                                      |
| [004](docs/adr/004-jwt-session-security.md)            | JWT blocklist + zero-downtime secret rotation  | Redis JTI blocklist for instant logout; dual-secret validation rotates `JWT_SECRET` without invalidating live sessions |
| [005](docs/adr/005-refresh-token-replay-detection.md)  | Refresh token rotation + replay detection      | Reuse of a rotated-out refresh token (theft or race) revokes all sessions for that user                              |
| [006](docs/adr/006-fail-closed-rate-limiting.md)       | Rate limiter fails closed on Redis error       | A Redis outage blocks `/auth/*` rather than silently disabling brute-force protection                               |

### Security & Reliability Mechanisms

Beyond the ADRs above, a few deliberate, non-obvious decisions worth calling out:

- **Timing-safe login** — a precomputed dummy bcrypt hash is compared even when the email doesn't exist, so response time can't reveal whether an account is registered ([ADR-004](docs/adr/004-jwt-session-security.md))
- **`SELECT ... FOR UPDATE SKIP LOCKED`** on the outbox poller — multiple worker instances can poll `event_outbox` concurrently without claiming the same row twice
- **Compile-time repository contract enforcement** (`make lint_repo_interfaces`) — every repository package must declare `var _ domain.XRepository = (*Repository)(nil)`, since no linter catches a silently-missing interface assertion
- **Non-root containers** — `api`, `worker`, and `migrate` all run as `app:app` (uid 1000), not root
- **Secrets never hit `argv`** — `SONAR_TOKEN` is written to a `chmod 600` temp file and passed via `docker run --env-file`, not `-e KEY=value` (which leaks through `ps`/`/proc/<pid>/cmdline`)

---

## Event System

LearnFlow uses the **Transactional Outbox Pattern** to guarantee reliable event delivery even if Redis is temporarily unavailable.

```
Business Operation (e.g. POST /briefs)
        │
        ▼
  Database Transaction
  ┌──────────────────────────────┐
  │  INSERT into domain table    │
  │  INSERT into event_outbox    │  ← same transaction
  └──────────────────────────────┘
        │
        ▼
  Outbox Worker (polls status = 'pending')
        │
        ▼
  Redis Queue  ──►  Domain Worker (BLPop, 5s timeout)
                          │
               ┌──────────┴──────────┐
               ▼                     ▼
         Downstream action     Send Notification
         (e.g. booking)        (e.g. confirmation email)
```

**Worker resilience:** 3× retry with exponential backoff → Dead Letter Queue (`failed_jobs` table) on exhaustion. Unresolved entries tracked with `resolved_at IS NULL` and visible in the admin panel.

Event type constants live in `internal/events/types.go`; payload structs in `internal/events/payloads.go`.

---

## Database Design

The schema spans 39 tables across six domains, designed for PostgreSQL 17+ with production-grade constraints.

**Design principles applied throughout:**

| Principle            | Implementation                                                                              |
| -------------------- | ------------------------------------------------------------------------------------------- |
| UUID primary keys    | `DEFAULT gen_random_uuid()` — generated in DB, not application                              |
| Immutable money      | `bigint` in cents everywhere — no `float64` for financial amounts                           |
| Soft deletes         | `deleted_at timestamptz` on 8 tables; every query filters `WHERE deleted_at IS NULL`        |
| Enum-like fields     | `text NOT NULL` + named `CHECK` constraint — readable without an enum type                  |
| Partial indexes      | Sparse columns, pending-status queues, active-record lookups — avoids scanning deleted rows |
| Deferred constraints | `DEFERRABLE INITIALLY DEFERRED` on position UNIQUE — allows in-transaction position swaps   |
| Transactional Outbox | `event_outbox` written in the same transaction as domain operations                         |

The annotated schema is in [`infrastructure/postgres/init/explain_init.sql`](infrastructure/postgres/init/explain_init.sql) with inline rationale for every design decision. A visual DBML model is at [`docs/DATABASE_SCHEMA.dbml`](docs/DATABASE_SCHEMA.dbml).

---

## Observability

The observability stack runs as a separate Docker Compose profile:

- **Prometheus** — scrapes Go runtime metrics + custom application metrics + `pg_stat_statements` via `postgres_exporter`
- **Grafana** — pre-configured dashboards for API latency, DB query performance, worker throughput, and error rates

```bash
make run_obs      # start observability stack only
make run_full     # full dev stack + observability
```

---

## Getting Started

### Prerequisites

- Docker + Docker Compose
- Go 1.26.2+
- Node.js 22+
- Make

### Quick Start

```bash
# Clone
git clone https://github.com/VZhyryk/learnflow.git
cd learnflow

# Copy environment config
cp .env.example .env

# Start the full stack (PostgreSQL, Redis, MinIO, backend, frontend)
make run

# Apply database migrations
make migrate
```

The API is available at `http://localhost:8080`, the frontend at `http://localhost:3000`, and MinIO console at `http://localhost:9001`.

### Environment Variables

Key variables from `.env.example`:

| Variable            | Default          | Description                       |
| ------------------- | ---------------- | --------------------------------- |
| `API_PORT`          | `8080`           | Backend HTTP port                 |
| `FRONTEND_PORT`     | `3000`           | Nuxt SSR port                     |
| `DATABASE_DSN`      | `postgres://...` | PostgreSQL connection string      |
| `REDIS_ADDR`        | `localhost:6379` | Redis address                     |
| `JWT_SECRET`        | `dev-secret`     | Change in production              |
| `EMAIL_PROVIDER`    | `stub`           | `stub` (local) or `resend` (prod) |
| `STORAGE_PROVIDER`  | `minio`          | `minio` (local) or `r2` (prod)    |
| `STRIPE_SECRET_KEY` | —                | Stripe secret API key             |
| `P24_MERCHANT_ID`   | —                | Przelewy24 merchant ID            |

---

## Development Commands

```bash
# Dev stack
make run                      # full stack via Docker Compose
make run_air                  # backend with Air hot-reload
make run_no_docker            # infra in Docker, backend locally

# Database
make migrate                  # apply all pending migrations
make migrate_down             # roll back last migration batch

# Testing
make run_test_coverage_backend   # Go tests with coverage report
make run_test_coverage_frontend  # Vue/TS tests with coverage report
make run_test_goconvey        # Go tests with GoConvey UI
make test_integration         # real-Postgres/Redis integration tests (-tags=integration)
make test_integration_down    # stop the integration test stack

# Linting
make lint_backend             # golangci-lint (15+ linters)
make lint_frontend            # oxlint + ESLint

# Observability
make run_obs                  # Prometheus + Grafana
make run_full                 # dev + observability

# Database backup
make backup                   # create timestamped backup
make restore FILE=backup.sql  # restore from file
```

**Single package test:**

```bash
cd backend && go test ./internal/auth/service/... -run TestName -v
```

**Frontend type check:**

```bash
cd frontend && npm run type-check
```

---

## Project Roadmap

| Phase | Focus                                                           | Status         |
| ----- | --------------------------------------------------------------- | -------------- |
| 1     | Backend foundation — project structure, infrastructure, CI      | ✅ Complete    |
| 1.1   | Spike — physical data model and database schema                 | ✅ Complete    |
| **2** | **Auth & Users — JWT, registration, verification, profiles**    | ✅ Complete    |
| 3     | Courses & Content — catalog, content items, completion tracking | ⬜ Planned     |
| 4     | Consultations — booking, availability, rescheduling             | ⬜ Planned     |
| 4.1   | Spike — real-time support chat (WebSocket)                      | ⬜ Planned     |
| 5     | Payments — Stripe, Przelewy24, access grants, coupons           | ⬜ Planned     |
| 6     | Notifications — delivery pipeline, reminders, re-engagement     | ⬜ Planned     |
| 6.1   | Spike — discount and coupon engine                              | ⬜ Planned     |
| 7     | Admin — user management, announcements, audit trail             | ⬜ Planned     |
| 8     | Analytics & P&L — UTM attribution, revenue reporting, ROAS      | ⬜ Planned     |
| 9     | Frontend — Nuxt 4 SSR, full UI, SEO, i18n                       | ⬜ Planned     |

---

## Project Structure

```
learnflow/
├── backend/
│   ├── cmd/
│   │   ├── api/            # HTTP server entrypoint
│   │   ├── worker/         # Event worker entrypoint
│   │   └── migrate/        # Migration runner
│   ├── internal/
│   │   ├── auth/           # }
│   │   ├── users/          # } Domain modules
│   │   ├── courses/        # } (transport → service → repository → domain)
│   │   ├── ...             # }
│   │   ├── shared/         # Cross-cutting concerns
│   │   ├── infrastructure/ # DB, Redis, logger, config, storage, email
│   │   ├── events/         # Event types and payloads
│   │   ├── queue/          # Redis BLPop worker infrastructure
│   │   └── worker/         # Domain workers
│   └── migrations/         # Sequential .up.sql / .down.sql files
├── frontend/
│   └── app/
│       ├── pages/          # Nuxt file-based routing
│       ├── components/     # Vue 3 SFC components
│       ├── composables/    # Data-fetching bridge layer
│       └── stores/         # Pinia state management
├── infrastructure/
│   ├── docker/             # Docker Compose configs (dev, obs, full)
│   └── postgres/init/      # Annotated reference schema
├── docs/
│   ├── adr/                # Architecture Decision Records
│   ├── DATABASE_SCHEMA.dbml
│   └── DATABASE_SCHEMA_GUIDE.md
├── prometheus/             # Prometheus scrape config
├── scripts/                # Utility scripts
└── makefile
```

---

## Skills Demonstrated

- **Clean Architecture** in Go — strict layer boundaries, interface-driven design, no global state
- **Event-driven systems** — transactional outbox, Redis-backed async workers, DLQ handling
- **Relational database design** — normalization, constraint design, partial indexes, soft deletes at scale
- **Full-stack SSR** — Nuxt 4 server-side rendering, composable data-fetching layer, Pinia
- **Payment integration** — Stripe webhooks, Przelewy24, idempotency via outbox pattern
- **Observability** — Prometheus instrumentation, Grafana dashboards, structured JSON logging
- **Production hygiene** — golangci-lint (15+ linters), oxlint + ESLint strict, migration safety rules, ADRs
- **Test discipline** — GoConvey unit suite (97.6%+ service coverage) plus real-Postgres/Redis integration tests exercising full BLPop → process → retry → DLQ worker flows

---

_LearnFlow is a personal learning project. It is not a live product._
