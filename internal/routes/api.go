package routes

import (
	"go-attendance-api/internal/handler"
	"go-attendance-api/internal/middleware"
	"go-attendance-api/internal/repository"
	"go-attendance-api/internal/service"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB) {

	// ======================
	// INIT DEPENDENCIES
	// ======================

	attendanceRepo := repository.NewAttendanceRepository(db)
	attendanceService := service.NewAttendanceService(attendanceRepo)
	attendanceHandler := handler.NewAttendanceHandler(attendanceService)

	authRepo := repository.NewAuthRepository(db)
	authService := service.NewAuthService(authRepo)
	authHandler := handler.NewAuthHandler(authService)

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	tenantRepo := repository.NewTenantRepository(db)
	tenantService := service.NewTenantService(tenantRepo)
	tenantHandler := handler.NewTenantHandler(tenantService)

	tenantSettingRepo := repository.NewTenantSettingRepository(db)
	tenantSettingService := service.NewTenantSettingService(tenantSettingRepo)
	tenantSettingHandler := handler.NewTenantSettingHandler(tenantSettingService)

	// ======================
	// SWAGGER
	// ======================

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// ======================
	// API GROUP
	// ======================

	api := r.Group("/api/v1")

	// ======================
	// PUBLIC ROUTES
	// ======================

	auth := api.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	api.GET("/ping", attendanceHandler.HelloTest)

	// ======================
	// PROTECTED ROUTES
	// ======================

	protected := api.Group("/")
	protected.Use(middleware.JWTAuth())

	{
		// Attendance
		protected.POST("/attendance", attendanceHandler.RecordAttendance)
		protected.GET("/attendance", attendanceHandler.GetAllAttendance)

		// Tenant
		protected.GET("/tenants", tenantHandler.GetAllTenant)
		protected.GET("/tenants/:id", tenantHandler.GetTenantByID)
		protected.POST("/tenants", tenantHandler.CreateTenant)

		// Tenant Settings
		protected.GET("/tenant-setting", tenantSettingHandler.GetSetting)
		protected.PUT("/tenant-setting", tenantSettingHandler.UpdateSetting)

		// Users
		protected.GET("/users", userHandler.GetAllUsers)
	}
}
