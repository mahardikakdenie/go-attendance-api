# Build stage
FROM golang:1.26-alpine AS builder

# Set Go proxy for faster downloads
ENV GOPROXY=https://proxy.golang.org,direct

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Install swaggo for documentation generation
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Copy dependency files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code
COPY . .

# Generate Swagger documentation
RUN swag init -g cmd/api/main.go

# Build the binary with optimization flags
# -s: Omit the symbol table and debug information.
# -w: Omit the DWARF symbol table.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o main ./cmd/api/main.go

# Final stage
FROM alpine:3.20

WORKDIR /app

# Install runtime dependencies and create a non-root user
RUN apk --no-cache add tzdata ca-certificates \
    && adduser -D -u 1000 appuser

# Copy the binary and generated docs from the builder stage
COPY --from=builder /app/main /app/main
COPY --from=builder /app/docs /app/docs

# Set ownership to the non-root user
RUN chown -R appuser:appuser /app

# Use the non-root user
USER appuser

# Default environment variables (can be overridden)
ENV APP_PORT=8080 \
    GIN_MODE=release \
    RUN_MIGRATION=true \
    RUN_SEEDER=true \
    RESET_DB=false

EXPOSE 8080

CMD ["/app/main"]
