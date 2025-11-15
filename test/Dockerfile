# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY c.go ./

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -o whatsapp-service c.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata sqlite-libs

WORKDIR /app

# Create data directory for session persistence
RUN mkdir -p /app/data

# Copy the binary from builder
COPY --from=builder /app/whatsapp-service .

# Expose port
EXPOSE 8083

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8083/ || exit 1

# Run the application
CMD ["./whatsapp-service"]
