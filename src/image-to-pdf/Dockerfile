# Stage 1: Build the application
FROM golang:1.23-alpine AS builder

# Set environment variables
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Install necessary dependencies
RUN apk --no-cache add git

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum for dependency caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go binary
RUN go build -o server .

# Stage 2: Create the lightweight runtime image
FROM alpine:latest

# Install minimal dependencies (if needed)
RUN apk --no-cache add ca-certificates

# Set the working directory
WORKDIR /root/

# Copy the Go binary from the builder stage
COPY --from=builder /app/server .

# Expose the application port
EXPOSE 8083

# Command to run the executable
CMD ["./server"]
