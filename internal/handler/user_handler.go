package handler

import (
	"math"
	"net/http"
	"strconv"

	// modelDto "go-attendance-api/internal/dto"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/service"
	"go-attendance-api/internal/utils"

	"github.com/gin-gonic/gin"
)

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func parseIncludeParams(c *gin.Context) []string {
	if include := c.Query("includes"); include != "" {
		return utils.ParseIncludes(include)
	}

	return utils.ParseIncludes(c.Query("include"))
}

type UserHandler interface {
	GetAllUsers(c *gin.Context)
	GetUserByID(c *gin.Context)
	GetMe(c *gin.Context)
	GetRecentActivities(c *gin.Context)
	UpdateProfilePhoto(c *gin.Context)
	CreateUser(c *gin.Context)
}

type userHandler struct {
	service service.UserService
}

func NewUserHandler(service service.UserService) UserHandler {
	return &userHandler{
		service: service,
	}
}

//////////////////////////////////////////////////////////////
// DTO
//////////////////////////////////////////////////////////////

type UpdateProfilePhotoRequest struct {
	MediaURL string `json:"media_url" binding:"required"`
}

//////////////////////////////////////////////////////////////
// HANDLERS
//////////////////////////////////////////////////////////////

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
// @Security CookieAuth
// @Success 200 {object} utils.APIResponse{data=[]model.UserResponse}
// @Failure 500 {object} utils.APIResponse
// @Router /api/v1/users [get]
func (h *userHandler) GetAllUsers(c *gin.Context) {
	var filter model.UserFilter

	ctx := c.Request.Context()

	filter.Name = c.Query("name")
	filter.Email = c.Query("email")
	filter.OrderBy = c.Query("order_by")
	filter.Sort = c.Query("sort")

	if roleIDStr := c.Query("role_id"); roleIDStr != "" {
		if val, err := strconv.Atoi(roleIDStr); err == nil {
			filter.RoleID = uint(val)
		}
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

	filter.RequesterID = c.MustGet("user_id").(uint)

	includes := parseIncludeParams(c)

	data, total, err := h.service.GetAllUsers(ctx, filter, includes)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			utils.BuildErrorResponse("Failed to fetch users", 500, "error", err.Error()),
		)
		return
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}
	offset := filter.Offset

	pagination := utils.Pagination{
		Total:       total,
		PerPage:     limit,
		CurrentPage: (offset / limit) + 1,
		LastPage:    int(math.Ceil(float64(total) / float64(limit))),
	}
	if pagination.LastPage == 0 {
		pagination.LastPage = 1
	}

	c.JSON(http.StatusOK,
		utils.BuildResponseWithPagination("Users fetched successfully", 200, "success", data, pagination),
	)
}

// @Summary Get User By ID
// @Description Get single user detail with optional includes
// @Tags Users
// @Produce json
// @Param id path int true "User ID"
// @Param includes query string false "Relations (comma separated: tenant,attendances)"
// @Security BearerAuth
// @Security CookieAuth
// @Success 200 {object} utils.APIResponse{data=model.UserResponse}
// @Failure 400 {object} utils.APIResponse
// @Failure 404 {object} utils.APIResponse
// @Router /api/v1/users/{id} [get]
func (h *userHandler) GetUserByID(c *gin.Context) {
	ctx := c.Request.Context()

	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(400, utils.BuildErrorResponse("Invalid ID", 400, "error", err.Error()))
		return
	}

	includes := parseIncludeParams(c)

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
// @Description Get authenticated user profile from token with optional includes
// @Tags Users
// @Produce json
// @Param includes query string false "Relations (comma separated: tenant,attendances)"
// @Security BearerAuth
// @Security CookieAuth
// @Success 200 {object} utils.APIResponse{data=model.UserResponse}
// @Failure 401 {object} utils.APIResponse
// @Failure 404 {object} utils.APIResponse
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

	includes := parseIncludeParams(c)

	// Always include required relations for GetMe as requested for efficiency
	requiredIncludes := []string{"tenant", "tenant.tenant_settings", "attendances", "role", "recent_activities", "role.permissions"}
	for _, inc := range requiredIncludes {
		if !contains(includes, inc) {
			includes = append(includes, inc)
		}
	}

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

// @Summary Get user's recent activities
// @Description Get recent activities for the logged-in user
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Success 200 {object} utils.APIResponse{data=[]model.RecentActivityResponse}
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /api/v1/users/me/activities [get]
func (h *userHandler) GetRecentActivities(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{
			"message": "Unauthorized",
		})
		return
	}

	userID := userIDVal.(uint)

	activities, err := h.service.GetRecentActivities(c.Request.Context(), userID)
	if err != nil {
		c.JSON(500, gin.H{
			"message": "Failed to fetch activities",
		})
		return
	}

	c.JSON(200, gin.H{
		"data": activities,
	})
}

//////////////////////////////////////////////////////////////
// ✅ NEW ENDPOINT: UPDATE PROFILE PHOTO
//////////////////////////////////////////////////////////////

// @Summary Update Profile Photo
// @Description Update logged-in user's profile photo
// @Tags Users
// @Accept json
// @Produce json
// @Param body body UpdateProfilePhotoRequest true "Profile Photo Payload"
// @Security BearerAuth
// @Security CookieAuth
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /api/v1/users/profile-photo [put]
func (h *userHandler) UpdateProfilePhoto(c *gin.Context) {
	var req UpdateProfilePhotoRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest,
			utils.BuildErrorResponse("Invalid request body", 400, "error", err.Error()),
		)
		return
	}

	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized,
			utils.BuildErrorResponse("Unauthorized", 401, "error", "user not logged in"),
		)
		return
	}

	userID := userIDVal.(uint)

	err := h.service.UpdateProfilePhoto(userID, req.MediaURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			utils.BuildErrorResponse("Failed to update profile photo", 500, "error", err.Error()),
		)
		return
	}

	c.JSON(http.StatusOK,
		utils.BuildResponse("Profile photo updated successfully", 200, "success", nil),
	)
}

// @Summary Create User
// @Description Create new user with role hierarchy (SuperAdmin can create any, Admin can create HR/Employee)
// @Tags Users
// @Accept json
// @Produce json
// @Param body body model.CreateUserRequest true "User Payload"
// @Security BearerAuth
// @Security CookieAuth
// @Success 201 {object} utils.APIResponse{data=model.UserResponse}
// @Failure 400 {object} utils.APIResponse
// @Router /api/v1/users [post]
func (h *userHandler) CreateUser(c *gin.Context) {
	var req model.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	adminID := c.MustGet("user_id").(uint)

	res, err := h.service.CreateUser(c.Request.Context(), adminID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Failed to create user", 400, "error", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, utils.BuildResponse("User created successfully", 201, "success", res))
}
