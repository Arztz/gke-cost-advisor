# =============================================================================
# Build Stage - Compile the Go application
# =============================================================================
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache ca-certificates git

# Set working directory
WORKDIR /build

# Copy go mod files first for dependency caching
COPY go.mod go.sum ./

# Download dependencies
# If go.sum doesn't exist, this will still work but won't cache
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the binary with optimizations
# CGO_ENABLED=0 for static binary (required for alpine distroless)
# GOOS=linux to ensure Linux binary
# -ldflags for stripping debug info and setting version
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o server ./cmd/server

# =============================================================================
# Runtime Stage - Minimal production image
# =============================================================================
# Using alpine for better compatibility while maintaining small size
FROM alpine:3.20

# Install CA certificates for HTTPS connections and busybox for health check
RUN apk add --no-cache ca-certificates busybox

# Create non-root user and group
RUN addgroup -g 1000 appgroup && \
    adduser -u 1000 -G appgroup -s /bin/sh -D appuser

# Set working directory
WORKDIR /app

# Copy the compiled binary from builder stage
COPY --from=builder /build/server .

# Copy static assets if any (uncomment if needed)
# COPY --from=builder /build/assets ./assets

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose the server port (configurable via SERVER_PORT env var)
EXPOSE 8080

# Health check: verify the server is responding
# Using wget to check the /health endpoint
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Set environment variables for runtime configuration
ENV SERVER_PORT=8080
ENV PROMETHEUS_ENDPOINT=http://localhost:9090
ENV GCP_REGION=us-central1

# Run the server
CMD ["./server"]