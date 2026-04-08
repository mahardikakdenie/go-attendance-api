FROM golang:1.26-alpine AS builder

WORKDIR /app

# Install swaggo for documentation generation
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Copy dependency files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code
COPY . .

# Generate Swagger documentation
RUN swag init -g cmd/api/main.go

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api/main.go

FROM alpine:3.20

WORKDIR /app

RUN apk --no-cache add tzdata ca-certificates

# Copy the binary and generated docs
COPY --from=builder /app/main /app/main
COPY --from=builder /app/docs /app/docs

# Default environment variables
ENV APP_PORT=8080
ENV RUN_MIGRATION=true
ENV RUN_SEEDER=true
ENV RESET_DB=false
ENV REDIS_ADDR=redis:6379
ENV DB_HOST=db
ENV DB_PORT=5432
ENV DB_USER=postgres
ENV DB_PASSWORD=password
ENV DB_NAME=attendance_db

EXPOSE 8080

CMD ["/app/main"]
