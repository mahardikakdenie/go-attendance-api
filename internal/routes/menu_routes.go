package routes

import (
	"go-attendance-api/internal/handler"
	"go-attendance-api/internal/middleware"
	"go-attendance-api/internal/model"

	"github.com/gin-gonic/gin"
)

func RegisterMenuRoutes(rg *gin.RouterGroup, h handler.MenuHandler) {
	menus := rg.Group("/menus")
	{
		menus.GET("/me", h.GetMyMenus)
		menus.GET("/overview", middleware.RequireBaseRole(model.BaseRoleSuperAdmin, model.BaseRoleAdmin), h.GetRoleMenuOverview)
	}

	superadmin := rg.Group("/superadmin/menus")
	superadmin.Use(middleware.RequireBaseRole(model.BaseRoleSuperAdmin))
	{
		superadmin.GET("", h.GetAllMenus)
		superadmin.POST("", h.CreateMenu)
		superadmin.PUT("/:id", h.UpdateMenu)
	}
}
