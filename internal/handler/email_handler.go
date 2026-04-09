package handler

import (
	"net/http"

	// modelDto "go-attendance-api/internal/dto"
	"go-attendance-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type EmailRequest struct {
	To      []string `json:"to" binding:"required"`
	Subject string   `json:"subject" binding:"required"`
	Html    string   `json:"html" binding:"required"`
}

// @Summary Send Test Email
// @Description Send a test email
// @Tags Email
// @Accept json
// @Produce json
// @Param request body EmailRequest true "Email Data"
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /api/v1/email/test [post]
func SendEmailTest(c *gin.Context) {
	var req EmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request", http.StatusBadRequest, "error", err.Error()))
		return
	}

	err := utils.SendEmail(req.To, req.Subject, req.Html)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to send email", http.StatusInternalServerError, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Email sent successfully", http.StatusOK, "success", nil))
}
