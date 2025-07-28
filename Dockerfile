# Stage 1: Build the application
FROM golang:1.21-alpine AS builder

# Install dependencies
RUN apk add --no-cache git ffmpeg

WORKDIR /app

# Copy go modules and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /main cmd/main.go

# Stage 2: Create the final, lightweight image
FROM alpine:latest

# Install ffmpeg
RUN apk add --no-cache ffmpeg

WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /main .

# Copy configuration and TLS keys
COPY .env .
COPY keys/ keys/

# Expose the application port
EXPOSE 8080

# Create a volume for the HLS output files
VOLUME /app/output

# Set the entrypoint
CMD ["./main"] 