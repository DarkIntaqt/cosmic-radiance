# Use the official Golang image to build the application
FROM golang:1.25 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go modules manifests
COPY go.mod go.sum ./

# Download the Go module dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o cosmic-radiance ./cmd/cosmic-radiance/main.go

# Use a minimal base image for the final container
FROM gcr.io/distroless/base-debian11

# Set the working directory inside the container
WORKDIR /

# Copy the built binary from the builder stage
COPY --from=builder /app/cosmic-radiance .

# Expose the application's port
# TODO: this probably doesn't work in all cases
EXPOSE 8001

# Command to run the application
CMD ["./cosmic-radiance"]