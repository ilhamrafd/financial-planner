# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary (enable CGO for PostgreSQL driver)
RUN CGO_ENABLED=1 GOOS=linux go build -o main .

# Runtime stage
FROM alpine:3.20

# Install CA certificates and timezone data
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/main .

# Expose port used by API
EXPOSE 8080

CMD ["./main"]