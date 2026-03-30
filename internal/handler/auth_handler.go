package handler

import (
	"net/http"

	"go-attendance-api/internal/model"
	"go-attendance-api/internal/service"

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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data"})
		return
	}

	user, err := h.service.Register(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Registration successful",
		"data":    user,
	})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data"})
		return
	}

	res, err := h.service.Login(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"data":    res,
	})
}
