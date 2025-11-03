# Build stage
FROM golang:1.24-alpine AS builder

# Install git (needed for some Go modules)
RUN apk add --no-cache git

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
# CGO_ENABLED=0 creates a statically linked binary
# -ldflags="-w -s" reduces binary size by removing debug info
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o delivery-route-system \
    ./cmd/main.go

# Final stage
FROM alpine:latest

# Install CA certificates and wget for HTTPS requests and health checks
RUN apk --no-cache add ca-certificates tzdata wget

# Create non-root user for security
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/delivery-route-system .

# Change ownership to non-root user
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose the port the app runs on
EXPOSE 8000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8000/health || exit 1

# Run the binary
CMD ["./delivery-route-system"]

