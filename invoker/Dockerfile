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
RUN apk add --no-cache git make
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=ui-builder /app/ui/dist ./ui/dist
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o app ./cmd/invoker/main.go

# Stage 3: Runtime
# FROM ubuntu:25.10
FROM alpine:3.21

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates
# RUN apt-get update && apt-get install -y

# Copy binary from builder
COPY --from=go-builder /app/app .

# Expose port
ENV PORT=8080
EXPOSE 8080

# Run the application
CMD ["./app"]
