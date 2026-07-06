# ── Build Stage ─────────────────────────────────────────────────────────────
FROM golang:1.23-alpine AS builder

# Install git and ca-certificates (needed for go get)
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy dependency manifests first (layer-cache friendly)
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /bin/api \
    ./cmd/api

# ── Run Stage ────────────────────────────────────────────────────────────────
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

# Non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy binary from builder
COPY --from=builder /bin/api .

# Copy migrations so the app can run them at startup
COPY --from=builder /app/migrations ./migrations

USER appuser

EXPOSE 8080

ENTRYPOINT ["./api"]
