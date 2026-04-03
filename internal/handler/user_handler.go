package handler

import (
	"net/http"
	"strconv"

	"go-attendance-api/internal/model"
	"go-attendance-api/internal/service"
	"go-attendance-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type UserHandler interface {
	GetAllUsers(c *gin.Context)
	GetUserByID(c *gin.Context)
	GetMe(c *gin.Context)
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
// @Description Get list of users with filter, sorting, pagination, and dynamic includes
// @Tags Users
// @Produce json
// @Param name query string false "Filter by Name"
// @Param email query string false "Filter by Email"
// @Param role query string false "Filter by Role (admin, manager, employee)"
// @Param limit query int false "Limit (default 10)"
// @Param offset query int false "Offset (default 0)"
// @Param order_by query string false "Order by field"
// @Param sort query string false "Sort direction (asc/desc)"
// @Param includes query string false "Relations (comma separated: tenant,attendances,attendances.user)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/users [get]
func (h *userHandler) GetAllUsers(c *gin.Context) {
	var filter model.UserFilter

	ctx := c.Request.Context()

	filter.Name = c.Query("name")
	filter.Email = c.Query("email")
	filter.OrderBy = c.Query("order_by")
	filter.Sort = c.Query("sort")

	if role := c.Query("role"); role != "" {
		filter.Role = model.UserRole(role)
	}

	if l := c.Query("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil {
			filter.Limit = val
		}
	}

	if o := c.Query("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil {
			filter.Offset = val
		}
	}

	if tenantIDVal, exists := c.Get("tenant_id"); exists {
		if tenantID, ok := tenantIDVal.(uint); ok {
			filter.TenantID = tenantID
		}
	}

	includes := utils.ParseIncludes(c.Query("includes"))

	data, total, err := h.service.GetAllUsers(ctx, filter, includes)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			utils.BuildErrorResponse("Failed to fetch users", 500, "error", err.Error()),
		)
		return
	}

	meta := map[string]interface{}{
		"total":    total,
		"limit":    filter.Limit,
		"offset":   filter.Offset,
		"includes": includes,
	}

	c.JSON(http.StatusOK,
		utils.BuildResponse("Users fetched successfully", 200, "success", gin.H{
			"data": data,
			"meta": meta,
		}),
	)
}

// @Summary Get User By ID
// @Description Get single user detail with optional includes
// @Tags Users
// @Produce json
// @Param id path int true "User ID"
// @Param includes query string false "Relations (comma separated: tenant,attendances)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/users/{id} [get]
func (h *userHandler) GetUserByID(c *gin.Context) {
	ctx := c.Request.Context()

	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(400, utils.BuildErrorResponse("Invalid ID", 400, "error", err.Error()))
		return
	}

	includes := utils.ParseIncludes(c.Query("includes"))

	user, err := h.service.GetByID(ctx, uint(id), includes)
	if err != nil {
		c.JSON(404, utils.BuildErrorResponse("User not found", 404, "error", err.Error()))
		return
	}

	c.JSON(200, utils.BuildResponse("Success", 200, "success", gin.H{
		"data":     user,
		"includes": includes,
	}))
}

// @Summary Get current user
// @Description Get authenticated user profile from token (httpOnly cookie) with optional includes
// @Tags Users
// @Produce json
// @Param includes query string false "Relations (comma separated: tenant,attendances)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/users/me [get]
func (h *userHandler) GetMe(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{
			"message": "Unauthorized",
		})
		return
	}

	userID := userIDVal.(uint)

	includes := utils.ParseIncludes(c.Query("includes"))

	user, err := h.service.GetMe(c.Request.Context(), userID, includes)
	if err != nil {
		c.JSON(404, gin.H{
			"message": "User not found",
		})
		return
	}

	c.JSON(200, gin.H{
		"data":     user,
		"includes": includes,
	})
}
