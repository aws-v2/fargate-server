# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies if needed
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mini-fargate main.go

# Final stage
FROM alpine:latest

# Add non-root user for security
RUN adduser -D -u 1000 appuser

WORKDIR /app

# Install certificates
RUN apk --no-cache add ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/mini-fargate .

# Use non-root user
USER appuser

# Expose the port
EXPOSE 8086

# Run the binary
CMD ["./mini-fargate"]
