# Stage 1: Build
FROM golang:1.23.1-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /main cmd/main.go

# Stage 2: Run
FROM alpine:latest

# Install certificates and ffmpeg
RUN apk --no-cache add ca-certificates ffmpeg

# Set working directory
WORKDIR /root/

# Copy built binary from builder stage
COPY --from=builder /main .

COPY --from=builder /app/templates /root/templates

# Optional: Expose port if your app uses it
EXPOSE 8080

# Run the binary
CMD ["./main"]