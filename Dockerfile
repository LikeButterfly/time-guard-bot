# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o time-guard-bot ./cmd/bot

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/time-guard-bot .

# Copy swagger docs if they exist
COPY --from=builder /app/docs ./docs

# Expose API port
EXPOSE 8080

# Run the bot
CMD ["./time-guard-bot"]
