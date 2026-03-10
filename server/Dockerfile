# Multi-stage build for edge-metrics-server
# Stage 1: Build stage
FROM golang:1.23.4-alpine AS builder

# Install build dependencies for CGO (SQLite3 requires CGO)
RUN apk add --no-cache \
    gcc \
    musl-dev \
    sqlite-dev

WORKDIR /build

# Copy go.mod and go.sum first for dependency caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary with CGO enabled (required for SQLite)
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o edge-metrics-server .

# Stage 2: Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add \
    ca-certificates \
    sqlite-libs \
    tzdata

WORKDIR /app

# Create non-root user for security
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    mkdir -p /data && \
    chown -R appuser:appuser /app /data

# Copy binary from builder stage
COPY --from=builder --chown=appuser:appuser /build/edge-metrics-server .

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8081

# Environment variables (can be overridden in deployment)
ENV PORT=8081
ENV DB_PATH=/data/config.db

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8081/health || exit 1

# Run the server
CMD ["./edge-metrics-server"]
