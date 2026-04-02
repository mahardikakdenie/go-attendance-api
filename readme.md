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
RUN_MIGRATION=false
RESET_DB=false
RUN_SEEDER=false
```

Notes:

- `RUN_MIGRATION=true` runs `AutoMigrate` on startup.
- `RESET_DB=true` drops tables first, then migrates again.
- `RUN_SEEDER=true` seeds tenants, users, and tenant settings.
- If `APP_PORT` is empty, the app defaults to `8080`.

## Run Locally

1. Install Go and PostgreSQL.
2. Create the database defined in `DB_NAME`.
3. Copy `.env.example` to `.env` and adjust the values.
4. Run the API:

```bash
go run cmd/api/main.go
```

The API will start on `http://localhost:<APP_PORT>`.

## Docker

Build and run:

```bash
docker build -t go-attendance-api .
docker run --env-file .env -p 8080:8080 go-attendance-api
```

Note: the container exposes port `8080`, so align `APP_PORT` with that if you want the app inside the container to listen on `8080`.

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
- JWT is required for protected routes.
- The codebase currently mixes English and Indonesian response messages.
