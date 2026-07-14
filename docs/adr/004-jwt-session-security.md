# ADR-004: JWT Session Security — Blocklist, Secret Rotation, Enumeration Guard

**Status:** Accepted

## Context

Stateless JWT access tokens cannot be revoked early — a token stays valid until `exp`, even if the user just logged out or an admin blocked the account. Separately: the JWT secret will eventually need to be rotated (compromise, routine hygiene), and rotating it naively would instantly log out every active user. And separately again: the login endpoint is a classic user-enumeration target via timing (a bcrypt comparison for an existing user is significantly slower than an instant "not found").

## Decision

**JTI blocklist (Redis):** on logout / password change / admin block, `blocklist:{jti}` is written to Redis with a TTL equal to the token's remaining lifetime (`SetNX`, `internal/auth/service/helpers.go`). The `AuthenticateUser` middleware checks `EXISTS blocklist:{jti}` on every request and fails closed if Redis is unavailable (see the rate-limiter ADR for the same principle).

**Zero-downtime secret rotation:** `ValidateToken` (`internal/shared/tokens/tokens.go`) first checks the signature against the current `JWT_SECRET`; if that yields `ErrSignatureInvalid`, it retries against `JWT_SECRET_PREV`. Rotating the secret means setting a new `JWT_SECRET`, moving the old one to `JWT_SECRET_PREV`, with no invalidation of live tokens.

**Timing-safe login (enumeration guard):** at service startup, a `dummyPasswordHash` is generated once (`bcrypt.GenerateFromPassword` on a fixed dummy password). If the email isn't found, `loginGetUser` (`internal/auth/service/login.go`) still runs a real `bcrypt.CompareHashAndPassword` against this precomputed hash — response time is identical whether "the user doesn't exist" or "the password is wrong." The hash is generated once at startup, not per request (`GenerateFromPassword` is deliberately slow — generating it per request would invert the timing signature).

## Consequences

- Logout becomes effectively instant for the access token, at the cost of a Redis dependency on the critical authentication path (fail-closed is the chosen trade-off)
- Rotating the JWT secret is a zero-downtime operational action, but requires holding two secrets in configuration simultaneously during the transition window
- `dummyPasswordHash` is generated once at service startup — no per-request bcrypt cost, and it uses the same `BcryptCost` as real user hashes (in tests, `bcrypt.MinCost`, to avoid slowing down the test suite)
