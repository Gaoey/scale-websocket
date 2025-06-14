Build stage
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install necessary build tools
RUN apk add --no-cache git gcc musl-dev

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o websocket-server ./cmd/server

# Final stage
FROM alpine:3.18

# Add ca certificates and timezone data
RUN apk --no-cache add ca-certificates tzdata

# Create a non-root user
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/websocket-server .

# Use non-root user
USER appuser

# Expose the port
EXPOSE 8080

# Set the entrypoint
ENTRYPOINT ["./websocket-server"]