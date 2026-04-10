package service

import (
	"context"
)

type PayrollService interface {
	Calculate(ctx context.Context, req PayrollRequest) (PayrollResponse, error)
}

type PayrollRequest struct {
	BasicSalary         float64 `json:"basicSalary"`
	Allowances          float64 `json:"allowances"`
	AttendanceDays      int     `json:"attendanceDays"`
	WorkingDaysInMonth  int     `json:"workingDaysInMonth"`
	OvertimeHours       float64 `json:"overtimeHours"`
	UnpaidLeaveDays     int     `json:"unpaidLeaveDays"`
	PTKPStatus          string  `json:"ptkpStatus"`
}

type PayrollResponse struct {
	Breakdown struct {
		ProratedBasic        float64 `json:"proratedBasic"`
		UnpaidLeaveDeduction float64 `json:"unpaidLeaveDeduction"`
		GrossIncome          float64 `json:"grossIncome"`
		Pph21Amount          float64 `json:"pph21Amount"`
		BPJS                struct {
			Health struct {
				Employee float64 `json:"employee"`
				Company  float64 `json:"company"`
			} `json:"health"`
			JHT    struct {
				Employee float64 `json:"employee"`
				Company  float64 `json:"company"`
			} `json:"jht"`
			JKK    float64 `json:"jkk"`
			JKM    float64 `json:"jkm"`
		} `json:"bpjs"`
	} `json:"breakdown"`
	NetSalary       float64 `json:"netSalary"`
	TotalDeductions float64 `json:"totalDeductions"`
}

type payrollService struct{}

func NewPayrollService() PayrollService {
	return &payrollService{}
}

func (s *payrollService) Calculate(ctx context.Context, req PayrollRequest) (PayrollResponse, error) {
	// Prorated Basic Calculation
	proratedBasic := req.BasicSalary
	unpaidLeaveDeduction := 0.0
	if req.WorkingDaysInMonth > 0 {
		oneDaySalary := req.BasicSalary / float64(req.WorkingDaysInMonth)
		unpaidLeaveDeduction = float64(req.UnpaidLeaveDays) * oneDaySalary
		proratedBasic = req.BasicSalary - unpaidLeaveDeduction
	}

	// Gross Income
	// Simple assumption: Basic + Allowances + Overtime (simplified for this mock/requirement)
	// In real logic, overtime has complex multiplier. Let's assume 1.5x hourly rate.
	hourlyRate := req.BasicSalary / 173.0
	overtimePay := req.OvertimeHours * hourlyRate * 1.5
	
	grossIncome := proratedBasic + req.Allowances + overtimePay

	// BPJS Calculation (Standard Indonesia)
	// Health: 1% employee, 4% company (Max cap apply usually, but simplified here)
	// JHT: 2% employee, 3.7% company
	// JKK: 0.24% - 1.74% company (assume 0.24%)
	// JKM: 0.3% company
	
	res := PayrollResponse{}
	res.Breakdown.ProratedBasic = proratedBasic
	res.Breakdown.UnpaidLeaveDeduction = unpaidLeaveDeduction
	res.Breakdown.GrossIncome = grossIncome

	// BPJS Health
	res.Breakdown.BPJS.Health.Employee = grossIncome * 0.01
	res.Breakdown.BPJS.Health.Company = grossIncome * 0.04
	
	// BPJS JHT
	res.Breakdown.BPJS.JHT.Employee = grossIncome * 0.02
	res.Breakdown.BPJS.JHT.Company = grossIncome * 0.037
	
	// BPJS JKK/JKM
	res.Breakdown.BPJS.JKK = grossIncome * 0.0024
	res.Breakdown.BPJS.JKM = grossIncome * 0.003

	// PPh 21 Calculation (Simplified TER approach)
	// In reality this depends on PTKP status and Category A/B/C
	pphRate := 0.05 // Default 5% for simplified demo
	if grossIncome > 10000000 {
		pphRate = 0.09
	} else if grossIncome > 5000000 {
		pphRate = 0.025
	}
	
	res.Breakdown.Pph21Amount = grossIncome * pphRate

	// Net Salary & Deductions
	res.TotalDeductions = res.Breakdown.Pph21Amount + res.Breakdown.BPJS.Health.Employee + res.Breakdown.BPJS.JHT.Employee
	res.NetSalary = grossIncome - res.TotalDeductions

	return res, nil
}
