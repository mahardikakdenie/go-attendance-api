# Go Attendance API

![Go](https://img.shields.io/badge/Go-1.26-blue.svg)
![Gin Gonic](https://img.shields.io/badge/Gin%20Gonic-v1.12.0-red.svg)
![GORM](https://img.shields.io/badge/GORM-v1.31.1-brightgreen.svg)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-blue.svg)
![Redis](https://img.shields.io/badge/Redis-7-red.svg)
![Docker](https://img.shields.io/badge/Docker-enabled-blue.svg)
![Swagger](https://img.shields.io/badge/Swagger-docs-green.svg)

A robust and scalable attendance management API built with Go, Gin Gonic, and GORM, featuring role-based access control, tenant management, and integration with PostgreSQL and Redis. This project includes comprehensive user features, attendance tracking, overtime management, and a new user activity log.

## Table of Contents

- [Features](#features)
- [Technologies Used](#technologies-used)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Local Development Setup (using Docker Compose)](#local-development-setup-using-docker-compose)
  - [Environment Variables](#environment-variables)
- [API Endpoints](#api-endpoints)
  - [Authentication](#authentication)
  - [Users](#users)
  - [Attendance](#attendance)
  - [Overtime](#overtime)
  - [Tenants & Tenant Settings](#tenants--tenant-settings)
  - [Media Upload](#media-upload)
  - [User Change Requests](#user-change-requests)
- [Swagger Documentation](#swagger-documentation)
- [Commit Message Strategy](#commit-message-strategy)
- [Contributing](#contributing)
- [License](#license)

## Features

-   **User Authentication**: Secure registration, login, and logout with JWT.
-   **Role-Based Access Control (RBAC)**: Fine-grained permissions based on user roles (SuperAdmin, Admin, HR, Employee).
-   **User Management**: CRUD operations for users, including profile updates and photo uploads.
-   **Recent Activity Log**: Track and view recent user activities (e.g., logins, clock-ins, profile changes).
-   **Tenant Management**: Support for multi-tenant architecture.
-   **Attendance Tracking**: Clock-in/out functionality with location, media, and status (on-time, late).
-   **Overtime Management**: Request, approve, and reject overtime.
-   **User Change Requests**: System for users to request changes to their data, requiring admin approval.
-   **Media Upload**: API for uploading images (e.g., profile photos, attendance media).
-   **Email Integration**: Placeholder for email notifications (e.g., password reset, notifications).
-   **API Documentation**: Auto-generated Swagger UI for easy API exploration.

## Technologies Used

-   **Go**: Primary programming language.
    -   [**Gin Gonic**](https://github.com/gin-gonic/gin): HTTP web framework for Go.
    -   [**GORM**](https://gorm.io/): ORM library for Go.
    -   [**golang-jwt/jwt/v5**](https://github.com/golang-jwt/jwt/v5): For JWT authentication.
    -   [**joho/godotenv**](https://github.com/joho/godotenv): For loading environment variables.
    -   [**google/uuid**](https://github.com/google/uuid): For UUID generation.
    -   [**go-playground/validator/v10**](https://github.com/go-playground/validator/v10): For request validation.
-   **Database**:
    -   [**PostgreSQL**](https://www.postgresql.org/): Relational database.
-   **Caching/Message Broker**:
    -   [**Redis**](https://redis.io/): In-memory data store, used for session management and anti-replay attacks.
-   **Containerization**:
    -   [**Docker**](https://www.docker.com/): For packaging the application.
    -   [**Docker Compose**](https://docs.docker.com/compose/): For orchestrating multi-container Docker applications.
-   **API Documentation**:
    -   [**Swag**](https://github.com/swaggo/swag): Converts Go annotations to Swagger UI documentation.

## Project Structure

```
go-attendance-api/
├── cmd/
│   └── api/
│       └── main.go           # Main application entry point
├── docs/                     # Auto-generated Swagger documentation
├── internal/
│   ├── config/               # Database, Redis, and other configurations
│   ├── dto/                  # Data Transfer Objects (request/response models)
│   ├── handler/              # HTTP handlers (controller layer)
│   ├── middleware/           # Gin middleware (JWT, RBAC)
│   ├── model/                # Database models (GORM entities)
│   ├── repository/           # Data access layer (GORM operations)
│   ├── routes/               # API route definitions
│   ├── seeder/               # Database seeders for initial data
│   ├── service/              # Business logic layer
│   └── utils/                # Utility functions (response, email, includes parsing)
├── .air.toml                 # Air hot-reloading configuration (for local dev)
├── .dockerignore             # Specifies intentionally untracked files to ignore by Docker
├── .env.example              # Example environment variables
├── .env.local                # Local environment variables (NOT committed to Git)
├── Dockerfile                # Docker build instructions
├── go.mod                    # Go modules file
├── go.sum                    # Go modules checksums
├── readme.md                 # Project README file
└── ...
```

## Getting Started

Follow these instructions to set up and run the project locally.

### Prerequisites

-   [Go (1.26.x or later)](https://golang.org/doc/install)
-   [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/)
-   A `.env.local` file (see [Environment Variables](#environment-variables) section below)

### Local Development Setup (using Docker Compose)

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/your-repo/go-attendance-api.git
    cd go-attendance-api
    ```

2.  **Create your `.env.local` file:**
    Copy the `.env.example` file and rename it to `.env.local`. Adjust the values as needed. This file will be used by `docker-compose`.

    ```bash
    cp .env.example .env.local
    # Edit .env.local with your preferred values or keep defaults for local setup
    ```

3.  **Build and run the Docker containers:**
    This command will build the `app` image, pull `postgres` and `redis` images, and start all services. It also performs database migrations and seeding automatically based on `RUN_MIGRATION` and `RUN_SEEDER` environment variables in `.env.local`.

    ```bash
    docker-compose up -d --build
    ```
    *   The `db` (PostgreSQL) container will be accessible on port `5433` (host) mapped to `5432` (container).
    *   The `redis` container will be accessible on port `6380` (host) mapped to `6379` (container).
    *   The `app` container (Go API) will be accessible on port `8086` (host) mapped to `8080` (container).

4.  **Verify services are running:**
    ```bash
    docker-compose ps
    docker-compose logs app # Check application logs
    ```
    You should see output similar to:
    ```
    db          | ✅ Database connected
    db          | ✅ Migrasi database berhasil
    db          | 🌱 Running seeder...
    db          | ✅ Seeder selesai
    app         | Redis connected: PONG
    app         | [GIN-debug] Listening and serving HTTP on :8080
    ```

### Environment Variables

The application relies on environment variables for configuration, especially for database and Redis connections.

**`.env.local`**: This file is used by `docker-compose` to supply environment variables to your services. It **should not be committed to version control**.

```properties
# Database Configuration
DB_HOST=db                 # Service name in docker-compose for PostgreSQL
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=attendance_db

# Redis Configuration
REDIS_ADDR=redis:6379      # Service name in docker-compose for Redis
REDIS_PASSWORD=            # Leave empty if no password

# Application Configuration
APP_PORT=8080              # Internal port for the Go application
RESET_DB=false             # Set to 'true' to drop all tables before migration
RUN_MIGRATION=true         # Set to 'true' to run database migrations on startup
RUN_SEEDER=true            # Set to 'true' to run database seeders on startup

# JWT Secret
JWT_SECRET=your_jwt_secret_key_here # **CHANGE THIS IN PRODUCTION**

# Resend Email Configuration (if applicable)
RESEND_API_KEY=
RESEND_FROM_EMAIL=

# Image Upload Configuration (if applicable)
IMGBB_API_KEY=
```

## API Endpoints

The API documentation is available via Swagger UI once the application is running. You can access it at `http://localhost:8086/swagger/index.html`.

Here's a brief overview of key endpoints:

### Authentication

| Method | Endpoint               | Description            |
| :----- | :--------------------- | :--------------------- |
| `POST` | `/api/v1/auth/register`| Register a new user    |
| `POST` | `/api/v1/auth/login`   | Authenticate and get JWT |
| `POST` | `/api/v1/auth/logout`  | Invalidate current session |

### Users

| Method | Endpoint                   | Description                      | Authentication |
| :----- | :------------------------- | :------------------------------- | :------------- |
| `GET`  | `/api/v1/users`            | Get all users (with filters)     | `Bearer`       |
| `GET`  | `/api/v1/users/{id}`       | Get user by ID                   | `Bearer`       |
| `GET`  | `/api/v1/users/me`         | Get current user's profile (preloads tenant, roles, attendance, recent activities) | `Bearer`       |
| `GET`  | `/api/v1/users/me/activities`| Get current user's recent activities | `Bearer`       |
| `POST` | `/api/v1/users`            | Create a new user (admin roles)  | `Bearer`       |
| `PUT`  | `/api/v1/users/profile-photo`| Update user's profile photo      | `Bearer`       |

### Attendance

| Method | Endpoint                   | Description                      | Authentication |
| :----- | :------------------------- | :------------------------------- | :------------- |
| `POST` | `/api/v1/attendance`       | Record clock-in/clock-out        | `Bearer`       |
| `GET`  | `/api/v1/attendance`       | Get all attendance records       | `Bearer`       |
| `GET`  | `/api/v1/attendance/summary`| Get attendance summary statistics| `Bearer`       |

### Overtime

| Method | Endpoint                   | Description                      | Authentication |
| :----- | :------------------------- | :------------------------------- | :------------- |
| `POST` | `/api/v1/overtime`         | Create an overtime request       | `Bearer`       |
| `GET`  | `/api/v1/overtime`         | Get all overtime requests        | `Bearer`       |
| `GET`  | `/api/v1/overtime/{id}`    | Get overtime request by ID       | `Bearer`       |
| `POST` | `/api/v1/overtime/approve/{id}`| Approve overtime request (Admin/HR)| `Bearer`    |
| `POST` | `/api/v1/overtime/reject/{id}`| Reject overtime request (Admin/HR)| `Bearer`     |

### Tenants & Tenant Settings

| Method | Endpoint                  | Description                     | Authentication |
| :----- | :------------------------ | :------------------------------ | :------------- |
| `GET`  | `/api/v1/tenants`         | Get all tenants (SuperAdmin only)| `Bearer`       |
| `POST` | `/api/v1/tenants`         | Create a new tenant (SuperAdmin only)| `Bearer`   |
| `GET`  | `/api/v1/tenants/{id}`    | Get tenant by ID                | `Bearer`       |
| `GET`  | `/api/v1/tenant-setting`  | Get current tenant settings     | `Bearer`       |
| `PUT`  | `/api/v1/tenant-setting`  | Update current tenant settings  | `Bearer`       |

### Media Upload

| Method | Endpoint                  | Description                     | Authentication |
| :----- | :------------------------ | :------------------------------ | :------------- |
| `POST` | `/api/v1/media/upload`    | Upload a file (e.g., image)     | `Bearer`       |

### User Change Requests

| Method | Endpoint                       | Description                            | Authentication |
| :----- | :----------------------------- | :------------------------------------- | :------------- |
| `POST` | `/api/v1/users/request-change` | Create a user data change request      | `Bearer`       |
| `GET`  | `/api/v1/users/pending-changes`| Get all pending change requests (Admin/HR) | `Bearer`   |
| `POST` | `/api/v1/users/approve-change/{id}`| Approve a change request (Admin/HR) | `Bearer`       |
| `POST` | `/api/v1/users/reject-change/{id}`| Reject a change request (Admin/HR)  | `Bearer`       |

## Swagger Documentation

Once the application is running via Docker Compose, you can access the interactive API documentation at:

[http://localhost:8086/swagger/index.html](http://localhost:8086/swagger/index.html)

This documentation is automatically generated from the code annotations.

## Commit Message Strategy

This project adheres to a structured commit message format to ensure clarity and facilitate code reviews. The format is as follows:

```
<type>(<scope>): <subject>

<body>

<footer>
```

Please refer to the [Commit Message Strategy](#commit-message-strategy) section in the documentation for detailed guidelines and examples.

## Contributing

Contributions are welcome! Please feel free to open issues or submit pull requests.
1.  Fork the repository.
2.  Create a new branch (`git checkout -b feature/your-feature`).
3.  Make your changes.
4.  Write clear, concise commit messages following the [Commit Message Strategy](#commit-message-strategy).
5.  Push your branch (`git push origin feature/your-feature`).
6.  Open a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
