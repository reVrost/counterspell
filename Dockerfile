# Stage 1: Build Svelte UI
FROM node:lts AS ui-builder
WORKDIR /app/ui
COPY ui/package.json ui/package-lock.json ./
RUN npm ci
COPY ui/ ./
RUN npm run build

# Stage 2: Build Go binary
FROM golang:1.25-alpine AS go-builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=ui-builder /app/ui/dist ./ui/dist
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o app ./cmd/app

# Stage 3: Runtime
FROM ubuntu:25.10

WORKDIR /app

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    git \
    bash \
    bubblewrap \
    ffmpeg \
    curl \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Install jj (prebuilt binary)
ARG TARGETARCH
RUN if [ "$TARGETARCH" = "amd64" ]; then \
    JJ_URL="https://github.com/jj-vcs/jj/releases/download/v0.37.0/jj-v0.37.0-x86_64-unknown-linux-musl.tar.gz"; \
    else \
    JJ_URL="https://github.com/jj-vcs/jj/releases/download/v0.37.0/jj-v0.37.0-aarch64-unknown-linux-musl.tar.gz"; \
    fi \
    && curl -fsSL "$JJ_URL" -o jj.tar.gz \
    && tar -xzf jj.tar.gz \
    && mv jj /usr/local/bin/jj \
    && chmod +x /usr/local/bin/jj \
    && rm -rf jj.tar.gz LICENSE README.md

# Install Claude Code via official installer
RUN curl -fsSL https://claude.ai/install.sh | bash

# Copy binary from builder
COPY --from=go-builder /app/app .

# Copy schema
COPY internal/db/schema.sql ./internal/db/schema.sql

# Create data directory for SQLite database
RUN mkdir -p /app/data

# Expose port
EXPOSE 8710

# Run the application
CMD ["./app", "-addr", ":8710"]
