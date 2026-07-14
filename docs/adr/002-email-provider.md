# ADR-002: Email Provider

**Status:** Accepted

## Context

Email is needed for: email confirmation after registration, password reset, notifications (booking_confirmed, payment_confirmed, refund_processed, reminder). The backend sends via SMTP or an HTTP API.

## Decision

**Resend** (resend.com) — HTTP API + SMTP relay.

Rationale:
- Free tier: 3,000 emails/month, 100/day — sufficient for an MVP
- Simple Go SDK (`github.com/resend/resend-go/v2`) or plain HTTP — minimal dependency footprint
- SMTP relay — usable with plain `net/smtp` if we want to avoid the SDK
- Transactional email with good deliverability
- Dashboard with per-email logs — convenient for debugging

Configuration via env:
```
EMAIL_PROVIDER=resend
RESEND_API_KEY=
EMAIL_FROM=noreply@learnflow.com
```

## Consequences

- For local dev — `delivery_status = 'stubbed'` in the `notifications` table; the real sender is never invoked
- The notification worker checks the `EMAIL_PROVIDER` env var — if `stub`, it logs the send and marks it `stubbed`
- Changing providers only requires swapping the implementation behind the `EmailSender` interface; the rest of the code is untouched
