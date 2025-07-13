# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Set the working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/server ./cmd/server

# Final stage
FROM alpine:3.18

# Install required packages
RUN apk --no-cache add ca-certificates tzdata

# Set timezone
ENV TZ=UTC

# Create app directory
WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/bin/server .

# Copy configuration files
COPY --from=builder /app/configs ./configs

# Create logs directory
RUN mkdir -p /app/logs

# Expose the application port
EXPOSE 8080

# Command to run the application
CMD ["./server"]
