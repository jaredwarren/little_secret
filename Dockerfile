# Build stage
FROM golang:1.23-alpine AS builder

# Set workspace
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy module manifests
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and data files
COPY . .

# Build the Go application binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o littlesecrets main.go

# Production stage
FROM alpine:3.20

WORKDIR /app

# Copy the binary and assets from builder
COPY --from=builder /app/littlesecrets .
COPY --from=builder /app/static ./static
COPY --from=builder /app/data ./data

# Expose the application port
EXPOSE 8080

# Run the app
ENV PORT=8080
CMD ["./littlesecrets"]
