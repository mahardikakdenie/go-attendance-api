package handler

import (
	"net/http"

	"go-attendance-api/internal/model"
	"go-attendance-api/internal/service"
	"go-attendance-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type UserHandler interface {
	GetAllUsers(c *gin.Context)
}

type userHandler struct {
	service service.UserService
}

func NewUserHandler(service service.UserService) UserHandler {
	return &userHandler{
		service: service,
	}
}

// @Summary Get All Users
// @Description Get list of users with dynamic filter and sorting
// @Tags Users
// @Accept json
// @Produce json
// @Param name query string false "Filter by Name"
// @Param email query string false "Filter by Email"
// @Param order_by query string false "Order by field (e.g., name, created_at)"
// @Param sort query string false "Sort direction (asc or desc)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/users [get]
func (h *userHandler) GetAllUsers(c *gin.Context) {
	filter := model.UserFilter{
		Name:    c.Query("name"),
		Email:   c.Query("email"),
		OrderBy: c.Query("order_by"),
		Sort:    c.Query("sort"),
	}

	users, err := h.service.GetAllUsers(filter)
	if err != nil {
		response := utils.BuildErrorResponse("Gagal mengambil data pengguna", http.StatusInternalServerError, "error", err.Error())
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := utils.BuildResponse("Berhasil mengambil data pengguna", http.StatusOK, "success", users)
	c.JSON(http.StatusOK, response)
}
