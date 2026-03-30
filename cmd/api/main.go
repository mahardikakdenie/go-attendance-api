package main

import (
	"fmt"
	"log"
	"os"

	"go-attendance-api/docs"
	"go-attendance-api/internal/handler"
	"go-attendance-api/internal/middleware"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
	"go-attendance-api/internal/seeder"
	"go-attendance-api/internal/service"

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

	if os.Getenv("RESET_DB") == "true" {
		db.Migrator().DropTable(&model.Attendance{}, &model.User{})
		log.Println("Tabel berhasil direset (Drop Table)")
	}

	if os.Getenv("RUN_MIGRATION") == "true" || os.Getenv("RESET_DB") == "true" {
		err = db.AutoMigrate(&model.User{}, &model.Attendance{})
		if err != nil {
			log.Fatalf("Gagal melakukan migrasi database: %v", err)
		}
		log.Println("Migrasi database berhasil dieksekusi")
	}

	if os.Getenv("RUN_SEEDER") == "true" {
		seeder.SeedUsers(db)
	}

	return db
}

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

	db := InitDB()

	attendanceRepo := repository.NewAttendanceRepository(db)
	attendanceService := service.NewAttendanceService(attendanceRepo)
	attendanceHandler := handler.NewAttendanceHandler(attendanceService)

	authRepo := repository.NewAuthRepository(db)
	authService := service.NewAuthService(authRepo)
	authHandler := handler.NewAuthHandler(authService)

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	r := gin.Default()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		api.GET("/ping", attendanceHandler.HelloTest)

		protected := api.Group("")
		protected.Use(middleware.JWTAuth())
		{
			protected.POST("/attendance", attendanceHandler.RecordAttendance)
			protected.GET("/users", userHandler.GetAllUsers)
		}
	}

	if err := r.Run(":" + appPort); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
