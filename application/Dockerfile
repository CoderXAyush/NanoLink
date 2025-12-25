# ---------- Build stage ----------
FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

# Copy go.mod first
COPY go.mod ./

# Download deps (safe even if none exist)
RUN go mod download

# Copy rest of the source
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o api-server ./cmd/api-server

# ---------- Runtime stage ----------
FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=builder /app/api-server .

EXPOSE 8080

USER nonroot:nonroot

CMD ["./api-server"]

