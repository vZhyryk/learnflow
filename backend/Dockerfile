# syntax=docker/dockerfile:1.7

FROM golang:1.26-alpine AS base

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

FROM base AS dev

RUN go install github.com/air-verse/air@latest

COPY . .

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]

FROM base AS builder

COPY . .

## Runtime binaries
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -buildvcs=false -o /out/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -buildvcs=false -o /out/worker ./cmd/worker
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -buildvcs=false -o /out/migrate ./cmd/migrate

## API runtime
FROM alpine:3.23.3 AS api
RUN apk add --no-cache ca-certificates \
    && addgroup -g 1000 app \
    && adduser -D -u 1000 -G app app
WORKDIR /app
COPY --from=builder /out/api /app/api
RUN chown app:app /app/api
USER app:app
EXPOSE 8080
CMD ["/app/api"]

## Worker runtime
FROM alpine:3.23.3 AS worker
RUN apk add --no-cache ca-certificates \
    && addgroup -g 1000 app \
    && adduser -D -u 1000 -G app app
WORKDIR /app
COPY --from=builder /out/worker /app/worker
RUN chown app:app /app/worker
USER app:app
CMD ["/app/worker"]

## Migration runtime
FROM alpine:3.23.3 AS migrate
RUN apk add --no-cache ca-certificates \
    && addgroup -g 1000 app \
    && adduser -D -u 1000 -G app app
WORKDIR /app
COPY --from=builder /out/migrate /app/migrate
COPY migrations /app/migrations
RUN chown -R app:app /app
USER app:app
CMD ["/app/migrate", "-direction", "up"]
