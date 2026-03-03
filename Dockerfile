# ── Stage 1: build ────────────────────────────────────────────────────────────
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Download dependencies first so Docker layer caching skips this on code-only changes.
COPY go.mod go.sum* ./
RUN go mod download

COPY . .

# Build a statically-linked binary. -s -w strips debug info (~30% smaller).
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -trimpath -o /app/server ./cmd/server

# ── Stage 2: run ──────────────────────────────────────────────────────────────
# distroless/static is scratch + TLS certs + timezone data.
# nonroot variant runs as uid 65532 — no shell, minimal attack surface.
FROM gcr.io/distroless/static:nonroot

COPY --from=builder /app/server /server

EXPOSE 8080

ENTRYPOINT ["/server"]
