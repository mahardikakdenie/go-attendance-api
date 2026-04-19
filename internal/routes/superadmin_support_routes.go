package routes

import (
	"go-attendance-api/internal/handler"
	"go-attendance-api/internal/middleware"
	"go-attendance-api/internal/model"
	"github.com/gin-gonic/gin"
)

func RegisterSuperadminRoutes(rg *gin.RouterGroup, superadminH handler.SuperadminHandler, tenantH handler.TenantHandler, subscriptionH handler.SubscriptionHandler) {
	superadmin := rg.Group("/superadmin")
	superadmin.Use(middleware.RequireBaseRole(model.BaseRoleSuperAdmin))
	{
		superadmin.GET("/owners-stats", superadminH.GetOwnersWithStats)
		superadmin.PUT("/tenants/:id", tenantH.UpdateTenant)

		platform := superadmin.Group("/platform-accounts")
		{
			platform.GET("", superadminH.GetPlatformAccounts)
			platform.POST("", superadminH.CreatePlatformAccount)
			platform.PUT("/:id", superadminH.UpdatePlatformAccount)
			platform.PATCH("/:id/status", superadminH.TogglePlatformAccountStatus)
		}

		roles := superadmin.Group("/system-roles")
		{
			roles.GET("", superadminH.ListSystemRoles)
			roles.POST("", superadminH.CreateSystemRole)
			roles.PUT("/:id", superadminH.UpdateSystemRole)
			roles.DELETE("/:id", superadminH.DeleteSystemRole)
		}
		superadmin.GET("/permissions", superadminH.ListAllPermissions)

		subscriptions := superadmin.Group("/subscriptions")
		{
			subscriptions.GET("", subscriptionH.GetSubscriptions)
			subscriptions.POST("/:id/remind", subscriptionH.RemindTenant)
			subscriptions.POST("/:id/suspend", subscriptionH.SuspendTenant)
		}
	}
}

func RegisterSupportRoutes(rg *gin.RouterGroup, h handler.SupportHandler) {
	support := rg.Group("/admin/support")
	support.Use(middleware.RequireTenant(1)) // HQ only
	{
		support.GET("/inbox", middleware.HasPermission("support.manage"), h.GetAllSupportMessages)
		support.PATCH("/inbox/:id", middleware.HasPermission("support.manage"), h.UpdateSupportStatus)
		support.GET("/trials", middleware.HasPermission("support.manage"), h.GetAllTrialRequests)
		support.PATCH("/trials/:id", middleware.HasPermission("support.manage"), h.UpdateTrialStatus)

		support.GET("/provisioning", middleware.RequireBaseRole(model.BaseRoleSuperAdmin), h.GetAllProvisioningTickets)
		support.POST("/provisioning/:id/execute", middleware.RequireBaseRole(model.BaseRoleSuperAdmin), h.ExecuteProvisioning)
	}

	rg.POST("/support/message", h.CreateSupportMessage)
}

func RegisterSubscriptionRoutes(rg *gin.RouterGroup, h handler.SubscriptionHandler) {
	subs := rg.Group("/subscriptions")
	subs.Use(middleware.RequireRole("superadmin", "admin"))
	{
		subs.GET("/me", h.GetMySubscription)
		subs.POST("/upgrade", h.UpgradeSubscription)
	}
}
