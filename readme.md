# Go Attendance API

REST API for attendance management built with Go, Gin, GORM, and PostgreSQL. The project uses a clean layered structure (`handler -> service -> repository`) and currently covers authentication, attendance recording, tenant management, tenant settings, and user listing.

## Stack

- Go
- Gin
- GORM
- PostgreSQL
- JWT authentication
- Swagger via `swaggo`
- Docker

## Main Features

- Register and login users
- JWT-protected API routes
- Clock in and clock out attendance flow
- Attendance listing with filters and pagination
- Tenant management endpoints
- Tenant setting endpoints
- Swagger UI for API exploration
- Optional migration, reset, and seeding on startup

## Project Structure

```text
go-attendance-api/
├── cmd/api/                 # Application entrypoint
├── docs/                    # Generated Swagger files
├── internal/
│   ├── config/              # Database bootstrap
│   ├── handler/             # HTTP handlers
│   ├── middleware/          # JWT middleware
│   ├── model/               # Entities and request/response models
│   ├── repository/          # Data access layer
│   ├── routes/              # Route registration
│   ├── seeder/              # Seed data
│   ├── service/             # Business logic
│   └── utils/               # Shared helpers
├── Dockerfile
├── go.mod
└── readme.md
```

## Environment Variables

Create a `.env` file in the project root.

```env
DB_HOST=127.0.0.1
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=1234
DB_NAME=attendance-db
APP_PORT=8085
JWT_SECRET=supersecretkey
RUN_MIGRATION=true
RESET_DB=false
RUN_SEEDER=true
```

Notes:

- `RUN_MIGRATION=true` runs `AutoMigrate` on startup.
- `RESET_DB=true` drops tables first, then migrates again.
- `RUN_SEEDER=true` seeds tenants, users, and tenant settings.
- If `APP_PORT` is empty, the app defaults to `8080`.

## Installation

1. Install Go, PostgreSQL, and the Swagger CLI.

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

2. Create a PostgreSQL database for the project.
3. Copy `.env.example` to `.env`.
4. Update `.env` so migration and seeder are enabled:

```env
DB_HOST=127.0.0.1
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=1234
DB_NAME=attendance-db
APP_PORT=8085
JWT_SECRET=supersecretkey
RUN_MIGRATION=true
RESET_DB=false
RUN_SEEDER=true
```

5. Generate Swagger docs from the main entrypoint:

```bash
swag init -g cmd/api/main.go
```

6. Run the API:

```bash
go run cmd/api/main.go
```

7. Open Swagger:

```text
http://localhost:8085/swagger/index.html
```

## Run Locally

Quick start after `.env` is configured:

```bash
swag init -g cmd/api/main.go
go run cmd/api/main.go
```

The API will start on `http://localhost:<APP_PORT>`.

## Docker

The included [Dockerfile](C:\Users\FS-User\Documents\go-attendance-api\Dockerfile) sets:

- `APP_PORT=8080`
- `RUN_MIGRATION=true`
- `RUN_SEEDER=true`
- `RESET_DB=false`

Build and run:

```bash
docker build -t go-attendance-api .
docker run --env-file .env -p 8080:8080 go-attendance-api
```

For Docker, use `APP_PORT=8080` in `.env` so the container port and application port match.

Recommended Docker `.env` values:

```env
APP_PORT=8080
RUN_MIGRATION=true
RUN_SEEDER=true
RESET_DB=false
```

## Swagger

Swagger UI is available at:

```text
http://localhost:<APP_PORT>/swagger/index.html
```

## API Overview

Base path:

```text
/api/v1
```

Public endpoints:

- `POST /auth/register`
- `POST /auth/login`
- `GET /ping`

Protected endpoints:

- `POST /attendance`
- `GET /attendance`
- `GET /users`
- `GET /tenants`
- `GET /tenants/:id`
- `POST /tenants`
- `GET /tenant-setting`
- `PUT /tenant-setting`

## Request Examples

Register:

```json
{
  "name": "Budi Santoso",
  "email": "budi@company.com",
  "password": "123456"
}
```

Login:

```json
{
  "email": "budi@company.com",
  "password": "123456"
}
```

Clock in:

```json
{
  "action": "clock_in",
  "latitude": -6.1339179,
  "longitude": 106.8329504,
  "media_url": "https://example.com/selfie.jpg"
}
```

Clock out:

```json
{
  "action": "clock_out",
  "latitude": -6.1339179,
  "longitude": 106.8329504,
  "media_url": "https://example.com/selfie-out.jpg"
}
```

Sample protected request header:

```text
Authorization: Bearer <jwt-token>
```

## Attendance Query Parameters

`GET /api/v1/attendance` supports:

- `user_id`
- `status`
- `date_from` in `YYYY-MM-DD`
- `date_to` in `YYYY-MM-DD`
- `limit`
- `offset`

## Seeded Data

When `RUN_SEEDER=true`, the app seeds:

- tenants
- users
- tenant settings

The tenant seed currently includes:

- `PT Friendship Logistics`
- `Remote Company Inc`
- `Hybrid Corp`

## Notes

- Swagger files in `docs/` are generated artifacts and should be refreshed when annotations change.
- Regenerate docs with `swag init -g cmd/api/main.go` after updating Swagger annotations.
- JWT is required for protected routes.
- The codebase currently mixes English and Indonesian response messages.
