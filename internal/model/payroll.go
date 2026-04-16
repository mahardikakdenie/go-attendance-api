package model

import (
	"time"

	"gorm.io/gorm"
)

type PayrollStatus string

const (
	PayrollStatusDraft     PayrollStatus = "Draft"
	PayrollStatusPublished PayrollStatus = "Published"
)

type Payroll struct {
	ID                   uint           `gorm:"primaryKey" json:"id"`
	TenantID             uint           `gorm:"index;not null" json:"tenant_id"`
	UserID               uint           `gorm:"index;not null" json:"user_id"`
	Period               string         `gorm:"type:varchar(7);not null;index" json:"period"` // Format: YYYY-MM
	
	// Snapshot Context
	EmployeeFullName     string         `json:"employee_full_name"`
	EmployeeID           string         `json:"employee_id_snapshot"`
	EmployeePosition     string         `json:"employee_position"`
	EmployeeDepartment   string         `json:"employee_department"`
	EmployeePtkpStatus   string         `json:"employee_ptkp_status"`
	
	// Earnings
	BasicSalary          float64        `json:"basic_salary"`
	FixedAllowances      float64        `json:"fixed_allowances"` // Basis for BPJS & Overtime
	VariableAllowances   float64        `json:"variable_allowances"` // Meal/Transport based on attendance
	Incentives           float64        `json:"incentives"`
	Bonus                float64        `json:"bonus"`
	OvertimePay          float64        `json:"overtime_pay"`
	
	// Working Info
	AttendanceDays       int            `json:"attendance_days"`
	WorkingDays          int            `json:"working_days"`
	UnpaidLeaveDays      int            `json:"unpaid_leave_days"`
	UnpaidLeaveDeduction float64        `json:"unpaid_leave_deduction"`
	
	// Deductions (Employee Share)
	GrossIncome          float64        `json:"gross_income"`
	Pph21Amount          float64        `json:"pph21_amount"`
	BpjsHealthEmployee   float64        `json:"bpjs_health_employee"`
	BpjsJhtEmployee      float64        `json:"bpjs_jht_employee"`
	BpjsJpEmployee       float64        `json:"bpjs_jp_employee"`
	
	// Employer Contributions (Company Share - for Total Employer Cost)
	BpjsHealthCompany    float64        `json:"bpjs_health_company"`
	BpjsJhtCompany       float64        `json:"bpjs_jht_company"`
	BpjsJpCompany        float64        `json:"bpjs_jp_company"`
	BpjsJkk              float64        `json:"bpjs_jkk"`
	BpjsJkm              float64        `json:"bpjs_jkm"`
	
	TotalDeductions      float64        `json:"total_deductions"`
	NetSalary            float64        `json:"net_salary"`
	Status               PayrollStatus  `gorm:"type:varchar(20);default:'Draft'" json:"status"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"-"`

	User   User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Tenant Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
}

type PayrollSummary struct {
	TotalNetPayout       float64 `json:"total_net_payout"`
	TotalTaxLiability    float64 `json:"total_tax_liability"`
	TotalBpjsProvision   float64 `json:"total_bpjs_provision"`
	AttendanceSyncRate   float64 `json:"attendance_sync_rate"`
	PayoutDiffPercentage float64 `json:"payout_diff_percentage"`
}
