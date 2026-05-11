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
	}

	superadmin := rg.Group("/superadmin/menus")
	superadmin.Use(middleware.RequireBaseRole(model.BaseRoleSuperAdmin))
	{
		superadmin.GET("", h.GetAllMenus)
	}
}
