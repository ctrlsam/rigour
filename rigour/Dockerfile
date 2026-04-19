# Multi-stage build for the Rigour server
FROM golang:1.24.0-alpine AS builder

WORKDIR /build

RUN apk add --no-cache git gcc musl-dev libpcap-dev

COPY . .

RUN go mod download

RUN go build -a -o server ./cmd/server

# Runtime
FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates libpcap

COPY --from=builder /build/server .

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["/app/server", "--help"]

ENTRYPOINT ["/app/server"]
