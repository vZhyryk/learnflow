# ADR-006: Rate Limiter Fails Closed on Redis Error

**Status:** Accepted

## Context

Rate limiting (token bucket, Redis Lua script) protects `/auth/*` from brute-force and abuse. Design question: what happens to a request when Redis itself is unavailable and the counter can't be checked? The typical "convenient" choice is fail-open (let the request through, so an outage in a supporting service doesn't take down the whole API). For endpoints like login/register/password-reset, that opens a window of unlimited brute-force at exactly the worst possible moment — during a Redis incident.

## Decision

`redisRateLimit` (`cmd/api/router/rateLimit.go`) returns `(false, err)` on any Redis error (network, timeout, script failure). The middleware (`cmd/api/router/middleware.go`) responds with an error immediately on a non-nil `err` (`respondRateLimiterError`, 500) and never reaches `next.ServeHTTP` — the request isn't processed at all, rather than being "let through without a limit."

## Consequences

- A Redis outage temporarily makes `/auth/*` unavailable to users — worse for availability than fail-open
- But it eliminates an entire class of incident: "Redis goes down → login is wide open to unlimited brute-force" — a deliberate choice favoring security over availability specifically for the auth surface
- Consistent with the same principle in the JWT blocklist check (ADR-004) — Redis-dependent security checks in this project are systematically fail-closed, not just the rate limiter
