# Build stage
FROM golang:1.25.5-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gido-guardian-bot .

# Final stage
FROM alpine:3.22

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/gido-guardian-bot .

# Create a non-root user for security
RUN addgroup -g 1000 -S appuser && adduser -u 1000 -S appuser -G appuser
USER appuser

ENTRYPOINT ["./gido-guardian-bot"]
