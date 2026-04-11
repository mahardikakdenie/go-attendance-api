# Go Attendance API

![Go](https://img.shields.io/badge/Go-1.26-blue.svg)
![Gin Gonic](https://img.shields.io/badge/Gin%20Gonic-v1.12.0-red.svg)
![GORM](https://img.shields.io/badge/GORM-v1.31.1-brightgreen.svg)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-blue.svg)
![Redis](https://img.shields.io/badge/Redis-7-red.svg)
![Docker](https://img.shields.io/badge/Docker-enabled-blue.svg)
![Swagger](https://img.shields.io/badge/Swagger-docs-green.svg)

A robust, enterprise-grade attendance management system built with Go. This API is designed for high-availability, featuring multi-tenant support, hierarchical organization structures, and optimized for high-traffic peak hours using advanced concurrency patterns.

## 🏗️ System Architecture & Workflow

### 1. High-Traffic Attendance Processing
To handle thousands of employees clocking in simultaneously during peak hours (e.g., 08:00 AM), the system employs a **Buffered Queue & Background Worker** strategy. This ensures the database never becomes a bottleneck and prevents server crashes.

```mermaid
sequenceDiagram
    participant E as Employee (Frontend)
    participant API as Attendance API
    participant R as Redis (Lock & Cache)
    participant Q as Buffered Queue (Go Channel)
    participant W as Background Workers
    participant DB as PostgreSQL

    E->>API: Clock In Request
    API->>R: Acquire Distributed Lock (SETNX)
    alt Lock Acquired
        R-->>API: Success
        API->>API: Fast Validation (GPS & Time)
        API->>Q: Push Attendance Task
        API->>R: Invalidate Today's Cache
        API-->>E: 200 OK (Instant Response)
        
        loop Asynchronous Processing
            Q->>W: Pick Task
            W->>DB: Save Record & Activity Log
        end
    else Lock Failed
        R-->>API: Conflict (Already Processing)
        API-->>E: 429 Too Many Requests
    end
```

### 2. Organizational Hierarchy & Leave Workflow
The system supports complex N-level organizational trees (CEO → VP → Head → Manager → Employee). It features intelligent leave approval routing with automatic escalation.

```mermaid
graph TD
    CEO[CEO - Level 1] --> VP[VP Operations - Level 2]
    VP --> DH[Dept Head - Level 3]
    DH --> M1[Manager A - Level 4]
    DH --> M2[Manager B - Level 4]
    M1 --> E1[Employee 1]
    M1 --> E2[Employee 2]

    subgraph Leave Approval Logic
    Request[Employee Requests Leave] --> CheckManager{Is Manager on Leave?}
    CheckManager -- No --> NotifyManager[Notify Direct Manager]
    CheckManager -- Yes --> CheckDelegate{Has Delegate?}
    CheckDelegate -- Yes --> NotifyDelegate[Notify Delegate]
    CheckDelegate -- No --> Escalate[Escalate to Skip-Level Manager]
    end
```

---

## 🚀 Key Features

-   **High-Concurrency Performance**: Uses Redis Distributed Locks to prevent race conditions and Go Channels for background DB persistence.
-   **Advanced Organization Tree**: Dynamic Org Chart support with job strata levels and reporting lines.
-   **Smart Leave Management**: 
    - Automatic approval routing based on hierarchy.
    - Temporary task delegation during leave.
    - Professional, responsive email templates for notifications.
-   **Multi-Tenancy**: Complete data isolation between different companies/tenants.
-   **Security**: JWT-based authentication, RBAC (Role-Based Access Control), and GPS-fencing for attendance.

---

## 🛠️ Technologies Used

-   **Backend**: Go (Gin Gonic), GORM (PostgreSQL).
-   **Caching/Concurrency**: Redis (Distributed Locking, Write-Behind Caching).
-   **Worker Pool**: Custom Goroutine-based worker pool for async task execution.
-   **Infrastructure**: Docker & Docker Compose.
-   **Communication**: Resend API for transactional emails with centralized HTML templates.

---

## 📂 Project Structure

```
go-attendance-api/
├── cmd/api/main.go           # Entry point & dependency injection
├── internal/
│   ├── handler/              # Controller layer (HTTP parsing)
│   ├── service/              # Business logic (Concurrency & Queue logic)
│   ├── repository/           # Data access (PostgreSQL & Redis)
│   ├── model/                # GORM entities & Org Tree nodes
│   └── utils/                # Email templates & response helpers
├── docs/                     # Swagger UI documentation
└── ...
```

---

## 🏁 Getting Started

### Prerequisites
-   Go 1.26+
-   Docker & Docker Compose
-   Redis 7.0+

### Setup
1.  **Clone & Configure**:
    ```bash
    cp .env.example .env.local
    # Set your JWT_SECRET and REDIS_ADDR
    ```
2.  **Spin Up Environment**:
    ```bash
    docker-compose up -d --build
    ```
3.  **Access Documentation**:
    Open `http://localhost:8080/swagger/index.html` to explore the API.

---

## 📄 API Rules & Guidelines

1.  **Response Format**: All responses follow a standardized `APIResponse` with a `meta` object containing pagination and status.
2.  **Concurrency**: Never perform heavy DB writes inside the main request context for attendance; always use the `recordQueue`.
3.  **Preloading**: Use the `includes` query parameter to fetch relationships (e.g., `?include=user,position`).

## 🤝 Contributing
Please follow the [Commit Message Strategy](#commit-message-strategy) for all PRs.

## 📜 License
This project is licensed under the MIT License.
