# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o carflow ./cmd

# Final stage
FROM alpine:3.19

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/carflow .
COPY --from=builder /app/docs /app/docs
COPY --from=builder /app/public /app/public

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./carflow"] 