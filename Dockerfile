# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build binaries
RUN CGO_ENABLED=1 GOOS=linux go build -o /bin/a2a-server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/a2a ./cmd/a2a

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy binaries from builder
COPY --from=builder /bin/a2a-server /bin/a2a-server
COPY --from=builder /bin/a2a /bin/a2a

# Create data directory
RUN mkdir -p /data

# Expose ports
EXPOSE 50051 8080

# Default command
ENTRYPOINT ["/bin/a2a-server"]
CMD ["--db", "/data/a2a.db"]
