# HRD Attendance API 🏢⏱️

A robust, scalable RESTful API built with Go (Golang) for managing employee attendances. This project is structured using **Clean Architecture** principles (Handler -> Service -> Repository), making it highly modular, testable, and maintainable.

## 🚀 Tech Stack

* **Language:** Go (Golang)
* **Web Framework:** [Gin HTTP Framework](https://github.com/gin-gonic/gin)
* **ORM:** [GORM](https://gorm.io/)
* **Database:** PostgreSQL
* **API Documentation:** [Swagger (swaggo)](https://github.com/swaggo/swag)

---

## 🏗️ Architecture & Folder Structure

This project follows a strict separation of concerns using the Dependency Injection pattern:

```text
go-attendance-api/
├── cmd/
│   └── api/
│       └── main.go          # Application entry point & Dependency Injection setup
├── docs/                    # Auto-generated Swagger documentation
├── internal/
│   ├── handler/             # Layer 1: HTTP Delivery (Gin Controllers & Routing)
│   ├── service/             # Layer 2: Business Logic (Attendance rules, time checking)
│   ├── repository/          # Layer 3: Data Access (PostgreSQL queries via GORM)
│   └── model/               # Layer 4: Entities & DTOs (Database structs & JSON binding)
├── go.mod                   # Go module dependencies
└── go.sum                   # Go module checksums
