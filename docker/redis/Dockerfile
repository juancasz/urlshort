# Stage 1: Build stage
FROM golang:1.21-alpine AS builder
LABEL stage=builder

# Set the working directory
WORKDIR /app

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/redis

# Stage 2: Final stage
FROM alpine:latest AS launch

# Set the working directory
WORKDIR /app

# Copy html files
COPY --from=builder /app/html /app/html

# Copy the binary from the build stage
COPY --from=builder /app/main .

EXPOSE 8080

# Set the entrypoint command
ENTRYPOINT ["/app/main"]