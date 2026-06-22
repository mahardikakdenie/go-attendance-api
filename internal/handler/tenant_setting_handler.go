package handler

import (
	"time"

	"go-attendance-api/internal/model"
	"go-attendance-api/internal/service"
	"go-attendance-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type TenantSettingHandler interface {
	GetSetting(c *gin.Context)
	UpdateSetting(c *gin.Context)
}

type tenantSettingHandler struct {
	service service.TenantSettingService
}

func NewTenantSettingHandler(service service.TenantSettingService) TenantSettingHandler {
	return &tenantSettingHandler{service: service}
}

// TenantSettingResponse is the DTO for sending settings with populated plan name
type TenantSettingResponse struct {
	ID                 uint                 `json:"id"`
	TenantID           uint                 `json:"tenant_id"`
	Tenant             model.TenantResponse `json:"tenant"`
	OfficeLatitude     float64              `json:"office_latitude"`
	OfficeLongitude    float64              `json:"office_longitude"`
	MaxRadiusMeter     float64              `json:"max_radius_meter"`
	AllowRemote        bool                 `json:"allow_remote"`
	RequireLocation    bool                 `json:"require_location"`
	ClockInStartTime   string               `json:"clock_in_start_time"`
	ClockInEndTime     string               `json:"clock_in_end_time"`
	LateAfterMinute    int                  `json:"late_after_minute"`
	ClockOutStartTime  string               `json:"clock_out_start_time"`
	ClockOutEndTime    string               `json:"clock_out_end_time"`
	RequireSelfie      bool                 `json:"require_selfie"`
	AllowMultipleCheck bool                 `json:"allow_multiple_check"`
	TenantLogo         string               `json:"tenant_logo"`
	BpjsHealthMaxBasis float64              `json:"bpjs_health_max_basis"`
	BpjsJpMaxBasis     float64              `json:"bpjs_jp_max_basis"`
	CreatedAt          time.Time            `json:"created_at"`
	UpdatedAt          time.Time            `json:"updated_at"`
}

func toTenantSettingResponse(s *model.TenantSetting) TenantSettingResponse {
	planName := "Trial" // Default fallback
	if s.Tenant.Subscription != nil && s.Tenant.Subscription.Plan != nil {
		planName = s.Tenant.Subscription.Plan.Name
	}

	return TenantSettingResponse{
		ID:              s.ID,
		TenantID:        s.TenantID,
		OfficeLatitude:  s.OfficeLatitude,
		OfficeLongitude: s.OfficeLongitude,
		MaxRadiusMeter:  s.MaxRadiusMeter,
		AllowRemote:     s.AllowRemote,
		RequireLocation: s.RequireLocation,
		ClockInStartTime:  s.ClockInStartTime,
		ClockInEndTime:    s.ClockInEndTime,
		LateAfterMinute:   s.LateAfterMinute,
		ClockOutStartTime: s.ClockOutStartTime,
		ClockOutEndTime:   s.ClockOutEndTime,
		RequireSelfie:     s.RequireSelfie,
		AllowMultipleCheck: s.AllowMultipleCheck,
		TenantLogo:         s.TenantLogo,
		BpjsHealthMaxBasis: s.BpjsHealthMaxBasis,
		BpjsJpMaxBasis:     s.BpjsJpMaxBasis,
		CreatedAt:          s.CreatedAt,
		UpdatedAt:          s.UpdatedAt,
		Tenant: model.TenantResponse{
			ID:              s.Tenant.ID,
			Name:            s.Tenant.Name,
			Plan:            planName,
			IsSuspended:     s.Tenant.IsSuspended,
			SuspendedReason: s.Tenant.SuspendedReason,
			Subscription:    s.Tenant.Subscription,
		},
	}
}

// @Summary Get Tenant Setting
// @Description Get settings for the current tenant
// @Tags Tenant Setting
// @Security BearerAuth
// @Security CookieAuth
// @Produce json
// @Success 200 {object} utils.APIResponse{data=TenantSettingResponse}
// @Failure 404 {object} utils.APIResponse
// @Router /api/v1/tenant-setting [get]
func (h *tenantSettingHandler) GetSetting(c *gin.Context) {
	tenantIDVal, _ := c.Get("tenant_id")
	tenantID := tenantIDVal.(uint)

	data, err := h.service.GetByTenant(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(404, utils.BuildErrorResponse("Not found", 404, "error", err.Error()))
		return
	}

	resp := toTenantSettingResponse(data)
	c.JSON(200, utils.BuildResponse("Success", 200, "success", resp))
}

// @Summary Update Tenant Setting
// @Description Update settings for the current tenant
// @Tags Tenant Setting
// @Security BearerAuth
// @Security CookieAuth
// @Accept json
// @Produce json
// @Param request body model.TenantSetting true "Tenant Setting"
// @Success 200 {object} utils.APIResponse{data=TenantSettingResponse}
// @Failure 400 {object} utils.APIResponse
// @Router /api/v1/tenant-setting [put]
func (h *tenantSettingHandler) UpdateSetting(c *gin.Context) {
	var req model.TenantSetting

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	tenantIDVal, _ := c.Get("tenant_id")
	tenantID := tenantIDVal.(uint)

	data, err := h.service.UpdateSetting(c.Request.Context(), tenantID, req)
	if err != nil {
		c.JSON(400, utils.BuildErrorResponse("Failed", 400, "error", err.Error()))
		return
	}

	resp := toTenantSettingResponse(data)
	c.JSON(200, utils.BuildResponse("Updated", 200, "success", resp))
}
