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
