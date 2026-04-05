package main

import (
	"fmt"
	"log"
	"os"

	"go-attendance-api/docs"
	"go-attendance-api/internal/config"
	"go-attendance-api/internal/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// @title HRD Attendance API
// @version 1.0
// @description API documentation for HRD Attendance System
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
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

	routes.SetupRoutes(r, db)

	if err := r.Run(":" + appPort); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
