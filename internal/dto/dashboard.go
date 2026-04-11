package modelDto

type MappedUser struct {
	ID           uint    `json:"id"`
	Name         string  `json:"name"`
	Avatar       string  `json:"avatar"`
	Department   string  `json:"department,omitempty"`
	Score        int     `json:"score,omitempty"`
	RequestCount int     `json:"request_count,omitempty"`
	TotalDays    int     `json:"total_days,omitempty"`
	Note         string  `json:"note,omitempty"` // Context like "Annual Leave" or "Sick Leave"
}

// Admin Dashboard
type AdminDashboardStats struct {
	TotalTenants  int64   `json:"total_tenants"`
	TotalUsers    int64   `json:"total_users"`
	ActiveSubs    int64   `json:"active_subs"`
	MonthlyGrowth float64 `json:"monthly_growth"`
}

type TenantGrowthItem struct {
	Month string `json:"month"`
	Count int64  `json:"count"`
}

type PlanDistributionItem struct {
	Label string       `json:"label"`
	Value int64        `json:"value"`
	Users []MappedUser `json:"users,omitempty"`
}

type AdminDashboardResponse struct {
	User             interface{}            `json:"user,omitempty"`
	Stats            AdminDashboardStats    `json:"stats"`
	TenantGrowth     []TenantGrowthItem     `json:"tenant_growth"`
	PlanDistribution []PlanDistributionItem `json:"plan_distribution"`
}

// HR Dashboard
type HrDashboardStats struct {
	PresenceRate float64      `json:"presence_rate"`
	AvgOvertime  float64      `json:"avg_overtime"`
	PendingLeave int64        `json:"pending_leave"`
	AtRiskStaff  int64        `json:"at_risk_staff"`
	AtRiskUsers  []MappedUser `json:"at_risk_users,omitempty"`
}

type EmployeePerformanceItem struct {
	ID            uint    `json:"id"`
	Name          string  `json:"name"`
	Avatar        string  `json:"avatar"`
	Department    string  `json:"department"`
	Score         int     `json:"score"`
	TotalLate     int     `json:"total_late"`
	AvgClockIn    string  `json:"avg_clock_in"`
	Status        string  `json:"status"`
	OvertimeHours float64 `json:"overtime_hours"`
	LeaveBalance  int     `json:"leave_balance"`
}

type HeatmapItem struct {
	Day   string       `json:"day"`
	Time  string       `json:"time"`
	Value int          `json:"value"`
	Users []MappedUser `json:"users,omitempty"`
}

type LeaveTrendSeries struct {
	Name string `json:"name"`
	Data []int  `json:"data"`
}

type HrDashboardResponse struct {
	User              interface{}               `json:"user,omitempty"`
	Stats             HrDashboardStats          `json:"stats"`
	TopPerformers     []EmployeePerformanceItem `json:"top_performers"`
	NeedAttention     []EmployeePerformanceItem `json:"need_attention"`
	PerformanceMatrix []EmployeePerformanceItem `json:"performance_matrix"`
	LeaveDistribution []PlanDistributionItem    `json:"leave_distribution"`
	LeaveTrends       []LeaveTrendSeries        `json:"leave_trends"`
}

type HeatmapQuery struct {
	UserID   uint   `form:"user_id"`
	Type     string `form:"type" binding:"omitempty,oneof=clockin clockout leave"` // clockin, clockout, leave
	DateFrom string `form:"date_from"`
	DateTo   string `form:"date_to"`
}

// Finance Dashboard
type FinanceDashboardStats struct {
	TotalPayroll      float64 `json:"total_payroll"`
	OvertimeCosts     float64 `json:"overtime_costs"`
	PendingDisbursals int64   `json:"pending_disbursals"`
	CostReduction     float64 `json:"cost_reduction"`
}

type PayrollTrendItem struct {
	Month         string  `json:"month"`
	BaseSalary    float64 `json:"base_salary"`
	OvertimeCosts float64 `json:"overtime_costs"`
}

type FinanceDashboardResponse struct {
	User          interface{}            `json:"user,omitempty"`
	Stats         FinanceDashboardStats  `json:"stats"`
	PayrollTrends []PayrollTrendItem     `json:"payroll_trends"`
	CostBreakdown []PlanDistributionItem `json:"cost_breakdown"`
}
