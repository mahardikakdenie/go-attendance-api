package handler

import (
	"net/http"

	"go-attendance-api/internal/model"
	"go-attendance-api/internal/service"
	"go-attendance-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type AuthHandler interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
	Logout(c *gin.Context)
	GetSessions(c *gin.Context)
}

type authHandler struct {
	service service.AuthService
}

func NewAuthHandler(service service.AuthService) AuthHandler {
	return &authHandler{
		service: service,
	}
}

// @Summary Register new employee
// @Description Register a new user with employee role by default
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body model.RegisterRequest true "Register Data"
// @Success 200 {object} utils.APIResponse{data=model.User}
// @Failure 400 {object} utils.APIResponse
// @Router /api/v1/auth/register [post]
func (h *authHandler) Register(c *gin.Context) {
	var req model.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response := utils.BuildErrorResponse("Invalid input data", http.StatusBadRequest, "error", err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	user, err := h.service.Register(req)
	if err != nil {
		response := utils.BuildErrorResponse("Failed to register user", http.StatusBadRequest, "error", err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	response := utils.BuildResponse("Registration successful", http.StatusOK, "success", user)
	c.JSON(http.StatusOK, response)
}

// @Summary Login employee
// @Description Authenticate user and get JWT via session cookie
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body model.LoginRequest true "Login Data"
// @Success 200 {object} utils.APIResponse{data=model.UserResponse}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Router /api/v1/auth/login [post]
func (h *authHandler) Login(c *gin.Context) {
	var req struct {
		model.LoginRequest
		DeviceInfo string `json:"device_info"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response := utils.BuildErrorResponse("Invalid request", http.StatusBadRequest, "error", err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	ip := c.ClientIP()
	ua := c.Request.UserAgent()

	token, user, err := h.service.Login(req.LoginRequest, ip, ua, req.DeviceInfo)
	if err != nil {
		response := utils.BuildErrorResponse("Login failed", http.StatusUnauthorized, "error", err.Error())
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "access_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true, // Requirement: Secure
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400,
	})

	response := utils.BuildResponse("Login successful", http.StatusOK, "success", user)
	c.JSON(http.StatusOK, response)
}

// @Summary Logout user
// @Description Invalidate current session and clear cookie
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Success 200 {object} utils.APIResponse
// @Router /api/v1/auth/logout [post]
func (h *authHandler) Logout(c *gin.Context) {
	token, _ := c.Cookie("access_token")

	_ = h.service.Logout(token)

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})

	response := utils.BuildResponse("Logout successful", http.StatusOK, "success", nil)
	c.JSON(http.StatusOK, response)
}

// @Summary Get active sessions
// @Description Get all active login sessions for the current user
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Success 200 {object} utils.APIResponse{data=[]model.SessionResponse}
// @Router /api/v1/auth/sessions [get]
func (h *authHandler) GetSessions(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	token, _ := c.Cookie("access_token")

	sessions, err := h.service.GetSessions(userID, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch sessions", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Sessions fetched successfully", 200, "success", sessions))
}
