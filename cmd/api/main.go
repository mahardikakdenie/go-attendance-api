package main

import (
	"fmt"
	"log"
	"os"

	"go-attendance-api/internal/handler"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
	"go-attendance-api/internal/service"

	_ "go-attendance-api/docs"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		host, user, password, dbName, port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Gagal koneksi ke database: %v", err)
	}

	if os.Getenv("RUN_MIGRATION") == "true" {
		err = db.AutoMigrate(&model.User{}, &model.Attendance{})
		if err != nil {
			log.Fatalf("Gagal melakukan migrasi database: %v", err)
		}
		log.Println("Migrasi database berhasil dieksekusi")
	}

	return db
}

// @title HRD Attendance API
// @version 1.0
// @description API documentation for HRD Attendance System
// @host localhost:8085
// @BasePath /
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Peringatan: File .env tidak ditemukan")
	}

	db := InitDB()

	attendanceRepo := repository.NewAttendanceRepository(db)
	attendanceService := service.NewAttendanceService(attendanceRepo)
	attendanceHandler := handler.NewAttendanceHandler(attendanceService)

	r := gin.Default()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1")
	{
		api.GET("/ping", attendanceHandler.HelloTest)
		api.POST("/attendance", attendanceHandler.RecordAttendance)
	}

	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = "8085"
	}

	if err := r.Run(":" + appPort); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
