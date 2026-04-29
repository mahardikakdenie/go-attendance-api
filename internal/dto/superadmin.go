package modelDto

import "time"

type OwnerWithStatsResponse struct {
	ID              uint      `json:"id"`
	Name            string    `json:"name"`
	Email           string    `json:"email"`
	TenantID        uint      `json:"tenant_id"`
	TenantName      string    `json:"tenant_name"`
	TenantCode      string    `json:"tenant_code"`
	TenantPlan      string    `json:"tenant_plan"`
	EmployeeCount   int64     `json:"employee_count"`
	AttendanceCount int64     `json:"attendance_count"`
	LeaveCount      int64     `json:"leave_count"`
	OvertimeCount   int64     `json:"overtime_count"`
	PayrollCount    int64     `json:"payroll_count"`
	ExpenseCount    int64     `json:"expense_count"`
	CreatedAt       time.Time `json:"created_at"`
}

type CreateSystemRoleRequest struct {
	Name          string   `json:"name" binding:"required"`
	Description   string   `json:"description"`
	BaseRole      string   `json:"base_role"`
	PermissionIDs []string `json:"permission_ids"`
}

type PermissionModule struct {
	Name        string               `json:"name"`
	Key         string               `json:"key"`
	Permissions []PermissionResponse `json:"permissions"`
}

type PermissionResponse struct {
	ID          string `json:"id"`
	Module      string `json:"module"`
	Action      string `json:"action"`
	Description string `json:"description"`
}

type AnalyticsDashboardResponse struct {
	KPIs         AnalyticsKPIs      `json:"kpis"`
	GrowthChart  GrowthChartData    `json:"growth_chart"`
	TenantStatus []TenantStatusData `json:"tenant_status"`
}

type AnalyticsKPIs struct {
	TotalTenants        KPIData `json:"total_tenants"`
	TotalUsers          KPIData `json:"total_users"`
	ActiveSubscriptions KPIData `json:"active_subscriptions"`
	MonthlyGrowth       KPIData `json:"monthly_growth"`
}

type KPIData struct {
	Value     int64   `json:"value"`
	GrowthPct float64 `json:"growth_pct"`
}

type GrowthChartData struct {
	Labels []string `json:"labels"`
	Data   []int64  `json:"data"`
}

type TenantStatusData struct {
	Label string `json:"label"`
	Value int64  `json:"value"`
	Color string `json:"color"`
}
