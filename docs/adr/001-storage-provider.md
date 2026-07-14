# ADR-001: Storage Provider

**Status:** Accepted

## Context

Files (video, PDF, images) are not stored in the database — only URL and metadata. An external object storage service is required. Requirements: S3-compatible API (to avoid locking into a single provider), support for presigned URLs for direct browser upload, a reasonable free tier for an MVP.

## Decision

**Production:** Cloudflare R2
**Local dev:** MinIO (Docker, S3-compatible)

Cloudflare R2:
- S3-compatible API — the same Go SDK (`aws-sdk-go-v2`) works unchanged
- No egress fees (unlike AWS S3)
- Free tier: 10 GB storage, 1M Class A operations/month, 10M Class B operations/month
- Supports presigned URLs for direct upload from the browser

MinIO for local dev:
- Fully S3-compatible
- Runs as a single Docker container
- Same code and SDK as R2

Configuration via env:
```
STORAGE_ENDPOINT=
STORAGE_ACCESS_KEY=
STORAGE_SECRET_KEY=
STORAGE_BUCKET=
STORAGE_REGION=auto  # for R2; for MinIO — us-east-1
```

## Consequences

- One Go client (`aws-sdk-go-v2/s3`) for both environments — switching is done purely via env vars
- Presigned URL flow: backend generates a URL → client uploads directly → backend receives a callback and stores the URL in the DB
- Changing providers only requires changing env vars, not code
