package routes

import (
	"go-attendance-api/internal/handler"
	"go-attendance-api/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterOrgRoutes(rg *gin.RouterGroup, orgH handler.OrganizationHandler, tenantRoleH handler.TenantRoleHandler) {
	org := rg.Group("/organization")
	{
		org.GET("/chart", orgH.GetOrgTree)
		org.GET("/positions", orgH.GetPositions)
		org.POST("/positions", middleware.RequireRole("superadmin", "admin"), orgH.CreatePosition)
	}

	tenantRoles := rg.Group("/tenant-roles")
	{
		tenantRoles.GET("", middleware.RequireRole("superadmin", "admin"), tenantRoleH.ListRoles)
		tenantRoles.POST("", middleware.RequireRole("superadmin", "admin"), tenantRoleH.CreateRole)
		tenantRoles.PATCH("/:id", middleware.RequireRole("superadmin", "admin"), tenantRoleH.UpdateRole)
		tenantRoles.DELETE("/:id", middleware.RequireRole("superadmin", "admin"), tenantRoleH.DeleteRole)
		tenantRoles.GET("/:id/hierarchy", middleware.RequireRole("superadmin", "admin"), tenantRoleH.GetHierarchy)
		tenantRoles.POST("/hierarchy", middleware.RequireRole("superadmin", "admin"), tenantRoleH.SaveHierarchy)
	}
}

func RegisterTenantRoutes(rg *gin.RouterGroup, tenantH handler.TenantHandler, tenantSettingH handler.TenantSettingHandler) {
	tenants := rg.Group("/tenants")
	{
		tenants.GET("", middleware.RequireRole("superadmin"), tenantH.GetAllTenant)
		tenants.POST("", middleware.RequireRole("superadmin"), tenantH.CreateTenant)
		tenants.GET("/:id", tenantH.GetTenantByID)
	}

	tenantSetting := rg.Group("/tenant-setting")
	{
		tenantSetting.GET("", tenantSettingH.GetSetting)
		tenantSetting.PUT("", tenantSettingH.UpdateSetting)
	}
}
