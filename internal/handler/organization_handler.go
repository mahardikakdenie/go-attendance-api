package handler

import (
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/service"
	"go-attendance-api/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type OrganizationHandler interface {
	GetOrgTree(c *gin.Context)
	CreatePosition(c *gin.Context)
	GetPositions(c *gin.Context)
}

type organizationHandler struct {
	service service.OrganizationService
}

func NewOrganizationHandler(service service.OrganizationService) OrganizationHandler {
	return &organizationHandler{service: service}
}

// @Summary Get Organization Chart
// @Description Get the full hierarchical tree of the organization
// @Tags Organization
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Success 200 {object} utils.APIResponse{data=[]model.OrgNode}
// @Router /api/v1/organization/chart [get]
func (h *organizationHandler) GetOrgTree(c *gin.Context) {
	tenantID := c.MustGet("tenant_id").(uint)

	tree, err := h.service.GetOrgTree(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to build org tree", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Org tree fetched", 200, "success", tree))
}

// @Summary Create Position
// @Description Create a new job level/position
// @Tags Organization
// @Accept json
// @Produce json
// @Param body body model.Position true "Position Data"
// @Security BearerAuth
// @Security CookieAuth
// @Router /api/v1/organization/positions [post]
func (h *organizationHandler) CreatePosition(c *gin.Context) {
	var req model.Position
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	req.TenantID = c.MustGet("tenant_id").(uint)

	if err := h.service.CreatePosition(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, utils.BuildResponse("Position created", 201, "success", req))
}

// @Summary Get Positions
// @Tags Organization
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Router /api/v1/organization/positions [get]
func (h *organizationHandler) GetPositions(c *gin.Context) {
	tenantID := c.MustGet("tenant_id").(uint)

	res, err := h.service.GetPositions(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Success", 200, "success", res))
}
