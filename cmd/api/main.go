package main

import (
	"fmt"
	"log"
	"os"

	"go-attendance-api/docs"
	"go-attendance-api/internal/config"
	"go-attendance-api/internal/middleware"
	"go-attendance-api/internal/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// @title Go Attendance API
// @version 1.0.0
// @description A robust and scalable attendance management API built with Go, Gin Gonic, and GORM.
// @description Featuring role-based access control, tenant management, and integration with PostgreSQL and Redis.
// @description This project includes comprehensive user features, attendance tracking, overtime management, and user activity logs.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /
// @query.collection.format multi

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

// @securityDefinitions.apikey CookieAuth
// @in cookie
// @name access_token
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Peringatan: File .env tidak ditemukan")
	}

	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = "8080"
	}

	docs.SwaggerInfo.Host = fmt.Sprintf("localhost:%s", appPort)

	db := config.InitDB()

	rdb := config.NewRedis()

	pong, err := rdb.Ping(config.Ctx).Result()
	if err != nil {
		panic(err)
	}

	println("Redis connected:", pong)

	r := gin.Default()
	r.Use(middleware.RateLimiter())

	routes.SetupRoutes(r, db)

	if err := r.Run(":" + appPort); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
