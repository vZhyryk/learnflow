# ADR-003: Payment Providers

**Status:** Accepted

## Context

The platform sells courses and consultations. It needs to support international card payments and Polish local payment methods (BLIK, bank transfer). The DB schema (`payments`, `payment_line_items`) is designed for multiple providers via the `provider_reference` field.

## Decision

**Stripe** — international cards
**Przelewy24** — Polish payment methods (BLIK, bank transfers, installments)

### Stripe
- Industry standard for SaaS payments
- Go SDK: `github.com/stripe/stripe-go/v78`
- Webhooks for async confirmation (`payment_intent.succeeded`)
- Handles 3D Secure automatically
- `provider_reference` = Stripe PaymentIntent ID

### Przelewy24
- The dominant provider in Poland — BLIK, 160+ banks, installments
- REST API, confirmation via webhook (`transaction.verified`)
- `provider_reference` = P24 transaction ID

### Architecture

```
payments.payment_method = 'card'          → Stripe
payments.payment_method = 'blik'          → Przelewy24
payments.payment_method = 'bank_transfer' → Przelewy24
```

The backend implements a `PaymentProvider` interface:
```go
type PaymentProvider interface {
    CreateIntent(ctx context.Context, req PaymentRequest) (PaymentIntent, error)
    HandleWebhook(ctx context.Context, payload []byte, sig string) (WebhookEvent, error)
}
```

Provider selection happens at the service level, based on `payment_method`.

Configuration via env:
```
STRIPE_SECRET_KEY=
STRIPE_WEBHOOK_SECRET=
P24_MERCHANT_ID=
P24_API_KEY=
P24_WEBHOOK_SECRET=
```

## Consequences

- Two separate webhook endpoints: `POST /webhooks/stripe`, `POST /webhooks/p24`
- Both insert an event into `event_outbox` within the same transaction — a single, unified processing flow
- Partial refunds are only implemented via Stripe (P24 supports full refund only for the MVP)
- Testing: Stripe test mode + P24 sandbox
