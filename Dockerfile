# Stage 1: Build the Go binary
FROM golang:1.25.0-alpine AS builder

# Set the working directory
WORKDIR /app

# Install git for downloading dependencies (if needed) and ca-certificates for secure HTTPS
RUN apk add --no-cache git ca-certificates

# Copy dependency files first to leverage Docker layer caching
COPY services/go-interceptor/go.mod services/go-interceptor/go.sum ./services/go-interceptor/

# Download all dependencies
WORKDIR /app/services/go-interceptor
RUN go mod download

# Copy the actual application source code
COPY services/go-interceptor/ .

# Build the statically linked binary executable for absolute portability
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go-interceptor .

# Stage 2: Create an ultra-lightweight, secure runtime image
FROM scratch

# Copy only the compiled binary and essential root certificates from the builder stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go-interceptor /bin/go-interceptor

# Expose the default HTTP port
EXPOSE 8080

# Command to execute when the container boots
ENTRYPOINT ["/bin/go-interceptor"]
