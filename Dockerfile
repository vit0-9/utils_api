# ---- Builder Stage ----
# Use an official Go image. Alpine is small.
# Ensure the Go version matches your project's go.mod or is compatible.
FROM golang:1.24-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Install build tools if necessary (e.g., git for private modules, though not needed here)
# RUN apk add --no-cache git

RUN go install github.com/swaggo/swag/cmd/swag@latest
# Copy go.mod and go.sum files to download dependencies first
# This leverages Docker's layer caching.
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy the rest of the application's source code
# This includes your main.go, app.go, handlers/, models/, pkg/, etc.
COPY . .

RUN /go/bin/swag init -g app.go
# Build the Go application
# -o /app/utils_api: output the binary named 'utils_api' to /app/
# -ldflags="-w -s": strip debugging information to reduce binary size
# CGO_ENABLED=0: build a statically-linked binary (important for Alpine/distroless)
# GOOS=linux GOARCH=amd64: ensure it's built for a typical Linux Docker environment
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/utils_api .

# ---- Final Stage ----
# Use a minimal base image for the final stage. Alpine is a good choice.
# For even smaller/more secure, consider gcr.io/distroless/static-debian11
# if your binary is fully static and has no external dependencies like libc calls (CGO_ENABLED=0 helps here).
FROM alpine:latest

# It's good practice to run as a non-root user for security.
# Create a group and user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set the working directory
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/utils_api .

# Copy the data directory containing MMDB files
# Ensure your 'data' directory is in the same context as your Dockerfile when building
COPY data ./data

# Expose the port the app runs on (as defined by PORT env var or default 8080)
# This doesn't publish the port, it's documentation for docker run -p
EXPOSE 8080

# Set environment variables
# These paths should match where you copied the files within the container
ENV MMDB_CITY_PATH=/app/data/GeoLite2-City.mmdb
ENV MMDB_ASN_PATH=/app/data/GeoLite2-ASN.mmdb
ENV PORT=8080
ENV GIN_MODE=release

# Switch to the non-root user
USER appuser

# Command to run the application
# ENTRYPOINT makes the container behave like an executable.
ENTRYPOINT ["/app/utils_api"]
