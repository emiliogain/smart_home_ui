# Multi-stage build for Smart Home Backend
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/smart-home-backend ./cmd/smart-home-backend

# Production image
FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Copy binary from builder stage
COPY --from=builder /app/bin/smart-home-backend .

# Copy configuration
COPY --from=builder /app/config ./config

EXPOSE 8080

CMD ["./smart-home-backend"]