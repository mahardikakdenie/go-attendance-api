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
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body model.RegisterRequest true "Register Data"
// @Success 200 {object} map[string]interface{}
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
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body model.LoginRequest true "Login Data"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/auth/login [post]
func (h *authHandler) Login(c *gin.Context) {
	var req model.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response := utils.BuildErrorResponse("Invalid request", http.StatusBadRequest, "error", err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	token, user, err := h.service.Login(req)
	if err != nil {
		response := utils.BuildErrorResponse("Login failed", http.StatusUnauthorized, "error", err.Error())
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400,
	})

	response := utils.BuildResponse("Login successful", http.StatusOK, "success", user)
	c.JSON(http.StatusOK, response)
}
