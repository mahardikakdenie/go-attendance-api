package modelDto

import (
	"go-attendance-api/internal/model"
	"time"
)

type SubscriptionStats struct {
	MRR                 float64 `json:"mrr"`
	MRRGrowth           string  `json:"mrr_growth"`
	ActiveTenants       int64   `json:"active_tenants"`
	ActiveTenantsGrowth string  `json:"active_tenants_growth"`
	PastDueAmount       float64 `json:"past_due_amount"`
	PastDueGrowth       string  `json:"past_due_growth"`
}

type SubscriptionItem struct {
	ID              uint                     `json:"id"`
	TenantID        uint                     `json:"tenant_id"`
	TenantName      string                   `json:"tenant_name"`
	TenantCode      string                   `json:"tenant_code"`
	TenantLogo      string                   `json:"tenant_logo"`
	Plan            string                   `json:"plan"`
	BillingCycle    model.BillingCycle       `json:"billing_cycle"`
	Amount          float64                  `json:"amount"`
	Status          model.SubscriptionStatus `json:"status"`
	NextBillingDate time.Time                `json:"next_billing_date"`
	ActiveEmployees int64                    `json:"active_employees"`
	CreatedAt       time.Time                `json:"created_at"`
}

type SubscriptionsResponse struct {
	Stats []SubscriptionStats `json:"stats"` // Note: Requested format had it as an object, but usually it's one stats object. I'll use a single object in the actual response if needed.
	Items []SubscriptionItem  `json:"items"`
	Total int64               `json:"total"`
}

// Actually, the issue description shows:
// "stats": { ... }
type SubscriptionsDataResponse struct {
	Stats SubscriptionStats  `json:"stats"`
	Items []SubscriptionItem `json:"items"`
	Total int64              `json:"total"`
}

type SuspendRequest struct {
	Reason string `json:"reason" binding:"required"`
}

type UpgradeRequest struct {
	Plan string `json:"plan" binding:"required"`
}

type CreatePlanRequest struct {
	Name         string   `json:"name" binding:"required"`
	MaxEmployees int      `json:"max_employees"`
	Features     []string `json:"features" binding:"required"`
}

type UpdatePlanRequest struct {
	Name         string   `json:"name"`
	MaxEmployees int      `json:"max_employees"`
	Features     []string `json:"features"`
	IsActive     *bool    `json:"is_active"`
}

type UpdateTenantSubscriptionRequest struct {
	PlanID          uint      `json:"plan_id"`
	Status          string    `json:"status"` // Active, Past Due, Canceled, Trial
	Amount          float64   `json:"amount"`
	NextBillingDate time.Time `json:"next_billing_date"`
}
