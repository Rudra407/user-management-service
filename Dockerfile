FROM golang:1.23-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o user-management-service ./cmd/server

# Use a smaller image for the final application
FROM alpine:latest

# Add ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/user-management-service .

# Create directory for the environment file
RUN mkdir -p /app/configs

# Expose the port
EXPOSE 8080

# Set the entry point
CMD ["./user-management-service"] 