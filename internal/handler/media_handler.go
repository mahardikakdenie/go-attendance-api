package handler

import (
	"io"
	"net/http"

	// modelDto "go-attendance-api/internal/dto"
	"go-attendance-api/internal/service"
	"go-attendance-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type MediaHandler struct {
	service service.MediaService
}

func NewMediaHandler(service service.MediaService) *MediaHandler {
	return &MediaHandler{service: service}
}

// Upload godoc
// @Summary Upload image to imgbb
// @Description Upload image file and store metadata to database
// @Tags Media
// @Accept multipart/form-data
// @Produce json
// @Param file formData file false "Upload file (file/image/media)"
// @Param image formData file false "Upload image"
// @Param media formData file false "Upload media"
// @Security BearerAuth
// @Security CookieAuth
// @Success 200 {object} utils.APIResponse{data=model.Media}
// @Failure 400 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /api/v1/media/upload [post]
func (h *MediaHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		file, err = c.FormFile("image")
		if err != nil {
			file, err = c.FormFile("media")
			if err != nil {
				c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("file required", http.StatusBadRequest, "error", nil))
				return
			}
		}
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("failed to open uploaded file", http.StatusBadRequest, "error", err.Error()))
		return
	}
	defer f.Close()

	buffer, err := io.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("failed to read uploaded file", http.StatusBadRequest, "error", err.Error()))
		return
	}

	result, err := h.service.Upload(c.Request.Context(), buffer, file.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse(err.Error(), http.StatusInternalServerError, "error", nil))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("success", http.StatusOK, "success", result))
}
