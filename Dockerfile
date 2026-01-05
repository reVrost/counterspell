# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install git (needed for worktree functionality)
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o app ./cmd/app

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install git for runtime worktree operations
RUN apk add --no-cache git ca-certificates

# Copy binary from builder
COPY --from=builder /app/app .

# Copy static files
COPY --from=builder /app/static ./static

# Copy schema
COPY --from=builder /app/internal/db/schema.sql ./internal/db/schema.sql

# Create data directory for SQLite database
RUN mkdir -p /app/data

# Expose port
EXPOSE 8710

# Run the application
CMD ["./app", "-addr", ":8710", "-db", "./data/counterspell.db"]
