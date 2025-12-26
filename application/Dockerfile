# Stage 1: Builder
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy Go module files first (better caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build API Server
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/api-server ./cmd/api-server

# Build Analytics Worker
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/analytics-worker ./cmd/analytics-worker

# Stage 2: Runner (Distroless/Alpine for security)
FROM alpine:latest

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /bin/api-server .
COPY --from=builder /bin/analytics-worker .

# Expose API port
EXPOSE 8080

# Default command (can be overridden in docker-compose)
CMD ["./api-server"]