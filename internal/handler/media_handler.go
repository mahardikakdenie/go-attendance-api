package handler

import (
	"io"
	"net/http"

	"go-attendance-api/internal/service"

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
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/media/upload [post]
func (h *MediaHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		file, err = c.FormFile("image")
		if err != nil {
			file, err = c.FormFile("media")
			if err != nil {
				c.JSON(400, gin.H{"message": "file required"})
				return
			}
		}
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "failed to open uploaded file"})
		return
	}
	defer f.Close()

	buffer, err := io.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "failed to read uploaded file"})
		return
	}

	result, err := h.service.Upload(c.Request.Context(), buffer, file.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data":    result,
	})
}
