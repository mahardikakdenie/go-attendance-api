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

	mediaRepo := repository.NewMediaRepository(db)
	mediaService := service.NewMediaService(mediaRepo)
	mediaHandler := handler.NewMediaHandler(mediaService)

	attendanceRepo := repository.NewAttendanceRepository(db)
	attendanceService := service.NewAttendanceService(attendanceRepo, userRepo, tenantSettingRepo, tenantRepo)
	attendanceHandler := handler.NewAttendanceHandler(attendanceService)

	if gin.Mode() != gin.ReleaseMode {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	api := r.Group("/api/v1")

	auth := api.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	api.GET("/ping", attendanceHandler.HelloTest)

	protected := api.Group("")
	protected.Use(middleware.CookieAuth(authService))
	{
		attendance := protected.Group("/attendance")
		{
			attendance.POST("", attendanceHandler.RecordAttendance)
			attendance.GET("", attendanceHandler.GetAllAttendance)
		}

		tenants := protected.Group("/tenants")
		{
			tenants.GET("", tenantHandler.GetAllTenant)
			tenants.GET("/:id", tenantHandler.GetTenantByID)
			tenants.POST("", tenantHandler.CreateTenant)
		}

		tenantSetting := protected.Group("/tenant-setting")
		{
			tenantSetting.GET("", tenantSettingHandler.GetSetting)
			tenantSetting.PUT("", tenantSettingHandler.UpdateSetting)
		}

		users := protected.Group("/users")
		{
			users.GET("", userHandler.GetAllUsers)
			users.GET("/me", userHandler.GetMe)
		}

		protected.POST("/media/upload", mediaHandler.Upload)
		protected.POST("/auth/logout", authHandler.Logout)
	}
}
