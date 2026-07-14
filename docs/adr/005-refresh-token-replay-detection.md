# ADR-005: Refresh Token Rotation with Replay Detection

**Status:** Accepted

## Context

A long-lived refresh token that never changes is an attractive target: if it's stolen (XSS, log leak, MITM), the attacker gets access for the token's entire TTL, unnoticed by the legitimate user. Simple rotation (a new refresh token on every `/auth/refresh`) solves only part of the problem — without an additional check, a stolen token and the legitimate one are equally "valid" the moment they're used.

## Decision

Every `refresh_hash` in `user_sessions` is replaced with a new one after use, and the old one is moved to `previous_refresh_hash` (not deleted immediately). `Refresh` (`internal/auth/service/refresh.go`) checks for two independent attack scenarios:

1. **Stale replay** — a token that was already rotated out shows up again (someone is using a stale copy) → `RevokeAllUserSessions`.
2. **Concurrent duplication** — the same refresh token surfaces simultaneously as another session's `previous_refresh_hash` (a race between the legitimate user and an attacker rotating "at the same time") → also `RevokeAllUserSessions`.

Both cases are recorded with `revoke_reason = 'suspicious_activity'`, which gives the user a visible signal in their session list rather than a silent logout.

## Consequences

- A stolen refresh token is useful to an attacker exactly once — the first use by the legitimate user after the theft (or vice versa) invalidates the entire session pool, including the attacker's
- The cost: one legitimate race (two parallel refresh requests from the same client, e.g. a duplicate network retry) also triggers mass revocation — an acceptable trade-off, since the client gets an explicit `401` and simply logs in again instead of silently losing access
- Requires keeping `previous_refresh_hash` instead of deleting the old hash immediately after rotation — an extra column and index (`idx_user_sessions_previous_refresh_hash`, migration 000003/000004)
