# Multi-stage build for optimal image size
# Because nobody wants a 1GB container for a simple API hehe

# Build stage
FROM golang:1.23-alpine AS builder

# Set up build environment
WORKDIR /app

# Install git for go mod download (if needed)
RUN apk add --no-cache git

# Copy go mod files first (for better Docker layer caching)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=0 ensures a static binary
# -ldflags="-w -s" strips debug info to reduce binary size
# TARGETARCH is provided by Docker BuildKit for multi-arch support (amd64, arm64, etc.)
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build \
    -ldflags="-w -s" \
    -o go-mail-checker .

FROM alpine:3.21

# Add ca-certificates for HTTPS requests (even though we don't use them, better safe than sorry)
RUN apk --no-cache add ca-certificates

# Create a non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/go-mail-checker .

# Change ownership to non-root user
RUN chown appuser:appgroup go-mail-checker

# Switch to non-root user
USER appuser

# Expose port 8080
EXPOSE 8080

# Health check to ensure container is running properly
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the binary
CMD ["./go-mail-checker"]