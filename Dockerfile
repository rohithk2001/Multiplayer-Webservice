# Stage 1: Build the application binary
FROM golang:1.23 AS builder

WORKDIR /app

# Copy Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the application binary (statically linked for Alpine compatibility)
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

# Stage 2: Create a minimal image for running the application
FROM alpine:latest

WORKDIR /root/

# Install necessary runtime dependencies
RUN apk add --no-cache ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Copy the .env file from the builder stage
COPY --from=builder /app/.env ./.env

# Expose ports for HTTP and gRPC
EXPOSE 8080 50051

# Set the entrypoint to the application binary
ENTRYPOINT ["./main"]