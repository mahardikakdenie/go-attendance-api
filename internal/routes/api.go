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

	attendanceRepo := repository.NewAttendanceRepository(db)
	attendanceService := service.NewAttendanceService(attendanceRepo, userRepo, tenantSettingRepo, tenantRepo)
	attendanceHandler := handler.NewAttendanceHandler(attendanceService)

	// Swagger (only dev)
	if gin.Mode() != gin.ReleaseMode {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	api := r.Group("/api/v1")

	// PUBLIC
	auth := api.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		// auth.POST("/logout", authHandler.Logout)
	}

	api.GET("/ping", attendanceHandler.HelloTest)

	// PROTECTED
	protected := api.Group("")
	protected.Use(middleware.JWTAuth())
	{
		protected.POST("/attendance", attendanceHandler.RecordAttendance)
		protected.GET("/attendance", attendanceHandler.GetAllAttendance)

		protected.GET("/tenants", tenantHandler.GetAllTenant)
		protected.GET("/tenants/:id", tenantHandler.GetTenantByID)
		protected.POST("/tenants", tenantHandler.CreateTenant)

		protected.GET("/tenant-setting", tenantSettingHandler.GetSetting)
		protected.PUT("/tenant-setting", tenantSettingHandler.UpdateSetting)

		protected.GET("/users", userHandler.GetAllUsers)
		protected.GET("/users/me", userHandler.GetMe)

	}
}
