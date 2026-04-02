FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /go-attendance-api ./cmd/api/main.go

FROM alpine:3.20

WORKDIR /app

RUN apk --no-cache add tzdata ca-certificates

COPY --from=builder /go-attendance-api /app/go-attendance-api

ENV APP_PORT=8080
ENV RUN_MIGRATION=true
ENV RUN_SEEDER=true
ENV RESET_DB=false

EXPOSE 8080

CMD ["/app/go-attendance-api"]
