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
	TenantStatus    string    `json:"tenant_status"`
	IsSuspended     bool      `json:"is_suspended"`
	PlanID          uint      `json:"plan_id"`
	EmployeeCount   int64     `json:"employee_count"`
	AttendanceCount int64     `json:"attendance_count"`
	LeaveCount      int64     `json:"leave_count"`
	OvertimeCount   int64     `json:"overtime_count"`
	PayrollCount    int64     `json:"payroll_count"`
	ExpenseCount    int64     `json:"expense_count"`
	CreatedAt       time.Time `json:"created_at"`
}

type CreateSystemRoleRequest struct {
	TenantID      *uint    `json:"tenant_id"`
	Name          string   `json:"name" binding:"required"`
	Description   string   `json:"description"`
	BaseRole      string   `json:"base_role"`
	PermissionIDs []string `json:"permission_ids"`
	Permissions   []string `json:"permissions"` // Alias: frontend sends "permissions"
}

type UpdateSystemRoleRequest struct {
	TenantID         *uint    `json:"tenant_id"`
	Name             *string  `json:"name"`
	Description      *string  `json:"description"`
	BaseRole         *string  `json:"base_role"`
	PermissionIDs    []string `json:"permissions"`    // User's payload uses "permissions"
	PermissionIDsAlt []string `json:"permission_ids"` // Keep support for original tag
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

type TenantFullDetailsResponse struct {
	Tenant struct {
		ID              uint      `json:"id"`
		Name            string    `json:"name"`
		Code            string    `json:"code"`
		CreatedAt       time.Time `json:"created_at"`
		IsSuspended     bool      `json:"is_suspended"`
		SuspendedReason string    `json:"suspended_reason"`
	} `json:"tenant"`
	Subscription struct {
		PlanName        string    `json:"plan_name"`
		Status          string    `json:"status"`
		Amount          float64   `json:"amount"`
		BillingCycle    string    `json:"billing_cycle"`
		NextBillingDate time.Time `json:"next_billing_date"`
	} `json:"subscription"`
	UsageStats struct {
		TotalEmployees   int64 `json:"total_employees"`
		TotalAttendances int64 `json:"total_attendances"`
		TotalLeaves      int64 `json:"total_leaves"`
		TotalPayrolls    int64 `json:"total_payrolls"`
		TotalExpenses    int64 `json:"total_expenses"`
	} `json:"usage_stats"`
	Employees []struct {
		ID         uint      `json:"id"`
		Name       string    `json:"name"`
		Email      string    `json:"email"`
		Role       string    `json:"role"`
		Position   string    `json:"position"`
		Department string    `json:"department"`
		CreatedAt  time.Time `json:"created_at"`
	} `json:"employees"`
}

type UpdateTenantRequest struct {
	Name            string `json:"name"`
	PlanID          uint   `json:"plan_id"`
	PlanName        string `json:"plan_name"` // Fallback for name-based update
	IsSuspended     *bool  `json:"is_suspended"`
	SuspendedReason string `json:"suspended_reason"`
}
