# Multi-stage build for ebcctl CLI tool
FROM golang:1.25.1-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /workspace

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o ebcctl ./cmd/ebcctl

# Final stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S ebcctl && \
    adduser -u 1001 -S ebcctl -G ebcctl

# Copy binary from builder stage
COPY --from=builder /workspace/ebcctl /usr/local/bin/ebcctl

# Set ownership and permissions
RUN chown ebcctl:ebcctl /usr/local/bin/ebcctl && \
    chmod +x /usr/local/bin/ebcctl

# Switch to non-root user
USER ebcctl

# Set working directory
WORKDIR /home/ebcctl

# Expose port (for potential future web interface)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ebcctl version || exit 1

# Default command
ENTRYPOINT ["ebcctl"]
CMD ["--help"]
