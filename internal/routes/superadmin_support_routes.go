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
		superadmin.GET("/analytics/dashboard", superadminH.GetAnalyticsDashboard)
		superadmin.GET("/owners-stats", superadminH.GetOwnersWithStats)
		superadmin.GET("/subscription-features", subscriptionH.GetSubscriptionFeatures)
		superadmin.GET("/tenants/:id/full-details", superadminH.GetTenantDetails)
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
			roles.PATCH("/:id", superadminH.PatchSystemRole)
			roles.DELETE("/:id", superadminH.DeleteSystemRole)
		}
		superadmin.GET("/permissions", superadminH.ListAllPermissions)
		superadmin.GET("/tenant-modules", superadminH.ListTenantModules)

		subscriptions := superadmin.Group("/subscriptions")
		{
			subscriptions.GET("", subscriptionH.GetSubscriptions)
			subscriptions.PUT("/:id", subscriptionH.UpdateTenantSubscription)
			subscriptions.POST("/:id/suspend", subscriptionH.SuspendTenant)
			subscriptions.POST("/:id/reactivate", subscriptionH.ReactivateSubscription)
			subscriptions.POST("/:id/remind", subscriptionH.RemindTenant)
		}

		plans := superadmin.Group("/plans")
		{
			plans.GET("", subscriptionH.GetAllPlans)
			plans.GET("/:id", subscriptionH.GetPlanByID)
			plans.POST("", subscriptionH.CreatePlan)
			plans.PUT("/:id", subscriptionH.UpdatePlan)
			plans.DELETE("/:id", subscriptionH.DeletePlan)
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
	rg.POST("/support/message/:id/reply", h.CreateReply)
	rg.GET("/support/message/:id/replies", h.GetReplies)
}

func RegisterSubscriptionRoutes(rg *gin.RouterGroup, h handler.SubscriptionHandler) {
	subs := rg.Group("/subscriptions")
	subs.Use(middleware.RequireRole("superadmin", "admin"))
	{
		subs.GET("/me", h.GetMySubscription)
		subs.POST("/upgrade", h.UpgradeSubscription)
		subs.GET("/plans", h.GetAllPlans) // Added for non-superadmin access to plan list
	}
}
