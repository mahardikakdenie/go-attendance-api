package service

import (
	"context"
	"errors"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
	"math"
	"strings"
	"time"
)

type PayrollService interface {
	Calculate(ctx context.Context, req PayrollRequest) (PayrollResponse, error)
	GeneratePayroll(ctx context.Context, tenantID uint, period string, runType model.PayrollRunType, method model.CalculationMethod) error
	GetAllPayroll(ctx context.Context, tenantID uint, period string, search string, limit, offset int) ([]PayrollResponse, int64, error)
	GetSummary(ctx context.Context, tenantID uint, period string) (model.PayrollSummary, error)
	PublishPayroll(ctx context.Context, tenantID uint, id uint) error
	GetMyPayrolls(ctx context.Context, userID uint) ([]PayrollResponse, error)

	// User Payroll Profile
	GetUserPayrollProfile(ctx context.Context, userID uint) (*model.UserPayrollProfile, error)
	UpdateUserPayrollProfile(ctx context.Context, userID uint, req UpdateUserPayrollProfileRequest) error
	GetMyPayrollProfile(ctx context.Context, userID uint) (*model.UserPayrollProfile, error)
	GetMySlip(ctx context.Context, userID uint, period string) (*PayrollResponse, error)

	// Individual Extensions
	GetEmployeeBaseline(ctx context.Context, userID uint) (EmployeeBaselineResponse, error)
	SyncEmployeeAttendance(ctx context.Context, userID uint, period string) (AttendanceSyncResponse, error)
	SaveIndividualPayroll(ctx context.Context, tenantID uint, userID uint, req SaveIndividualPayrollRequest) error
}

type PayrollRequest struct {
	UserID                  uint                    `json:"userId"`
	RunType                 model.PayrollRunType    `json:"runType"`
	Method                  model.CalculationMethod `json:"method"`
	BasicSalary             float64                 `json:"basicSalary"`
	FixedAllowances         float64                 `json:"fixedAllowances"`
	DailyMealAllowance      float64                 `json:"dailyMealAllowance"`
	DailyTransportAllowance float64                 `json:"dailyTransportAllowance"`
	Incentives              float64                 `json:"incentives"`
	Bonus                   float64                 `json:"bonus"`
	THR                     float64                 `json:"thr"`
	CalculateTHR            bool                    `json:"calculateThr"`
	AttendanceDays          int                     `json:"attendanceDays"`
	WorkingDaysInMonth      int                     `json:"workingDaysInMonth"`
	OvertimeHours           float64                 `json:"overtimeHours"`
	UnpaidLeaveDays         int                     `json:"unpaidLeaveDays"`
	PTKPStatus              string                  `json:"ptkpStatus"`
}

type PayrollResponse struct {
	ID                   uint                     `json:"id"`
	RunType              model.PayrollRunType     `json:"run_type"`
	Method               model.CalculationMethod  `json:"method"`
	CompanyContext       CompanyContext           `json:"company_context"`
	User                 EmployeeContext          `json:"user"`
	Breakdown            DetailedPayrollBreakdown `json:"breakdown"`
	NetSalary            float64                  `json:"net_salary"`
	PeriodText           string                   `json:"period_text"`
	AttendanceDays       int                      `json:"attendance_days"`
	WorkingDays          int                      `json:"working_days"`
	UnpaidLeaveDays      int                      `json:"unpaid_leave_days"`
	UnpaidLeaveDeduction float64                  `json:"unpaid_leave_deduction"`
	Status               model.PayrollStatus      `json:"status"`
}

type CompanyContext struct {
	Name    string `json:"name"`
	LogoURL string `json:"logo_url"`
	Address string `json:"address"`
}

type EmployeeContext struct {
	FullName          string `json:"full_name"`
	EmployeeID        string `json:"employee_id"`
	Position          string `json:"position"`
	Department        string `json:"department"`
	PTKPStatus        string `json:"ptkp_status"`
	BankName          string `json:"bank_name"`
	BankAccountNumber string `json:"bank_account_number"`
}

type DetailedPayrollBreakdown struct {
	Earnings struct {
		BasicSalary        float64 `json:"basic_salary"`
		FixedAllowances    float64 `json:"fixed_allowances"`
		VariableAllowances float64 `json:"variable_allowances"`
		OvertimePay        float64 `json:"overtime_pay"`
		Incentives         float64 `json:"incentives"`
		Bonus              float64 `json:"bonus"`
		THR                float64 `json:"thr"`
		TaxAllowance       float64 `json:"tax_allowance,omitempty"`  // For Net/Gross-up
		BpjsAllowance      float64 `json:"bpjs_allowance,omitempty"` // For Net/Gross-up
		GrossIncome        float64 `json:"gross_income"`
	} `json:"earnings"`
	Deductions struct {
		Pph21Amount          float64 `json:"pph21_amount"`
		UnpaidLeaveDeduction float64 `json:"unpaid_leave_deduction"`
		BpjsHealthEmployee   float64 `json:"bpjs_health_employee"`
		BpjsJhtEmployee      float64 `json:"bpjs_jht_employee"`
		BpjsJpEmployee       float64 `json:"bpjs_jp_employee"`
		TotalDeductions      float64 `json:"total_deductions"`
	} `json:"deductions"`
	EmployerContributions struct {
		BpjsHealthCompany float64 `json:"bpjs_health_company"`
		BpjsJhtCompany    float64 `json:"bpjs_jht_company"`
		BpjsJpCompany     float64 `json:"bpjs_jp_company"`
		BpjsJkk           float64 `json:"bpjs_jkk"`
		BpjsJkm           float64 `json:"bpjs_jkm"`
		TotalEmployerCost float64 `json:"total_employer_cost"`
	} `json:"employer_contributions"`
}

type EmployeeBaselineResponse struct {
	UserID          uint    `json:"user_id"`
	EmployeeName    string  `json:"employee_name"`
	Department      string  `json:"department"`
	PTKPStatus      string  `json:"ptkp_status"`
	BasicSalary     float64 `json:"basic_salary"`
	FixedAllowances float64 `json:"fixed_allowances"`
}

type AttendanceSyncResponse struct {
	Period             string  `json:"period"`
	WorkingDaysInMonth int     `json:"working_days_in_month"`
	AttendanceDays     int     `json:"attendance_days"`
	UnpaidLeaveDays    int     `json:"unpaid_leave_days"`
	OvertimeHours      float64 `json:"overtime_hours"`
}

type SaveIndividualPayrollRequest struct {
	Period                  string                  `json:"period" binding:"required"`
	RunType                 model.PayrollRunType    `json:"run_type"`
	Method                  model.CalculationMethod `json:"method"`
	BasicSalary             float64                 `json:"basic_salary"`
	FixedAllowances         float64                 `json:"fixed_allowances"`
	DailyMealAllowance      float64                 `json:"daily_meal_allowance"`
	DailyTransportAllowance float64                 `json:"daily_transport_allowance"`
	Incentives              float64                 `json:"incentives"`
	Bonus                   float64                 `json:"bonus"`
	THR                     float64                 `json:"thr"`
	AttendanceDays          int                     `json:"attendance_days"`
	WorkingDaysInMonth      int                     `json:"working_days_in_month"`
	OvertimeHours           float64                 `json:"overtime_hours"`
	UnpaidLeaveDays         int                     `json:"unpaid_leave_days"`
	PTKPStatus              string                  `json:"ptkp_status"`
	Status                  model.PayrollStatus     `json:"status"`
}

type UpdateUserPayrollProfileRequest struct {
	BankName                string           `json:"bank_name"`
	BankAccountNumber       string           `json:"bank_account_number"`
	BankAccountHolder       string           `json:"bank_account_holder"`
	BpjsHealthNumber        string           `json:"bpjs_health_number"`
	BpjsEmploymentNumber    string           `json:"bpjs_employment_number"`
	NpwpNumber              string           `json:"npwp_number"`
	PtkpStatus              model.PtkpStatus `json:"ptkp_status"`
	BasicSalary             float64          `json:"basic_salary"`
	FixedAllowance          float64          `json:"fixed_allowance"`
	DailyMealAllowance      float64          `json:"daily_meal_allowance"`
	DailyTransportAllowance float64          `json:"daily_transport_allowance"`
}

type payrollService struct {
	repo           repository.PayrollRepository
	userRepo       repository.UserRepository
	tenantRepo     repository.TenantRepository
	settingRepo    repository.TenantSettingRepository
	attendanceRepo repository.AttendanceRepository
	leaveRepo      repository.LeaveRepository
	profileRepo    repository.UserPayrollProfileRepository
	overtimeRepo   repository.OvertimeRepository
	hrOpsRepo      repository.HrOpsRepository
}

func NewPayrollService(
	repo repository.PayrollRepository,
	userRepo repository.UserRepository,
	tenantRepo repository.TenantRepository,
	settingRepo repository.TenantSettingRepository,
	attendanceRepo repository.AttendanceRepository,
	leaveRepo repository.LeaveRepository,
	profileRepo repository.UserPayrollProfileRepository,
	overtimeRepo repository.OvertimeRepository,
	hrOpsRepo repository.HrOpsRepository,
) PayrollService {
	return &payrollService{
		repo:           repo,
		userRepo:       userRepo,
		tenantRepo:     tenantRepo,
		settingRepo:    settingRepo,
		attendanceRepo: attendanceRepo,
		leaveRepo:      leaveRepo,
		profileRepo:    profileRepo,
		overtimeRepo:   overtimeRepo,
		hrOpsRepo:      hrOpsRepo,
	}
}

// Indonesian Payroll Constants
const (
	MaxHealthBasis = 12000000.0
	MaxJPBasis     = 10042300.0 // 2024 approximation

	// Overtime constants
	OvertimeHourlyDivider = 173.0
	OvertimeRate          = 1.5 // Simplified standard

	// BPJS Employee Rates
	BpjsHealthEmployeeRate = 0.01
	BpjsJhtEmployeeRate    = 0.02
	BpjsJpEmployeeRate     = 0.01

	// BPJS Employer Rates
	BpjsHealthCompanyRate = 0.04
	BpjsJhtCompanyRate    = 0.037
	BpjsJpCompanyRate     = 0.02
	BpjsJkkRate           = 0.0024
	BpjsJkmRate           = 0.003
)

func (s *payrollService) Calculate(ctx context.Context, req PayrollRequest) (PayrollResponse, error) {
	// Set defaults
	if req.RunType == "" {
		req.RunType = model.RunTypeRegular
	}
	if req.Method == "" {
		req.Method = model.MethodGross
	}

	// If profile exists, use it as baseline
	if req.UserID != 0 {
		profile, _ := s.profileRepo.FindByUserID(ctx, req.UserID)
		if profile != nil {
			if req.BasicSalary == 0 {
				req.BasicSalary = profile.BasicSalary
			}
			if req.FixedAllowances == 0 {
				req.FixedAllowances = profile.FixedAllowance
			}
			if req.PTKPStatus == "" {
				req.PTKPStatus = string(profile.PtkpStatus)
			}
			if req.DailyMealAllowance == 0 {
				req.DailyMealAllowance = profile.DailyMealAllowance
			}
			if req.DailyTransportAllowance == 0 {
				req.DailyTransportAllowance = profile.DailyTransportAllowance
			}
		} else {
			// Fallback to user base salary
			user, err := s.userRepo.FindByID(ctx, req.UserID, []string{})
			if err == nil && req.BasicSalary == 0 {
				req.BasicSalary = user.BaseSalary
			}
		}
	}

	// 1. Determine base components based on RunType
	var baseSalary, fixedAllowance, meal, transport, overtime, incentives, bonus, thr float64

	switch req.RunType {
	case model.RunTypeRegular:
		baseSalary = req.BasicSalary
		fixedAllowance = req.FixedAllowances
		meal = req.DailyMealAllowance
		transport = req.DailyTransportAllowance
		overtime = req.OvertimeHours
		incentives = req.Incentives
	case model.RunTypeTHR:
		if req.CalculateTHR && req.THR == 0 {
			thr = req.BasicSalary + req.FixedAllowances
		} else {
			thr = req.THR
		}
	case model.RunTypeBonus:
		bonus = req.Bonus
	case model.RunTypeAll:
		baseSalary = req.BasicSalary
		fixedAllowance = req.FixedAllowances
		meal = req.DailyMealAllowance
		transport = req.DailyTransportAllowance
		overtime = req.OvertimeHours
		incentives = req.Incentives
		bonus = req.Bonus
		if req.CalculateTHR && req.THR == 0 {
			thr = req.BasicSalary + req.FixedAllowances
		} else {
			thr = req.THR
		}
	}

	// 2. Prorate Calculation
	proratedBasic := baseSalary
	proratedFixedAllowance := fixedAllowance
	unpaidLeaveDeduction := 0.0

	if req.WorkingDaysInMonth > 0 && req.RunType != model.RunTypeTHR && req.RunType != model.RunTypeBonus {
		attendanceRatio := float64(req.AttendanceDays) / float64(req.WorkingDaysInMonth)
		proratedBasic = baseSalary * attendanceRatio
		proratedFixedAllowance = fixedAllowance * attendanceRatio

		if req.UnpaidLeaveDays > 0 {
			oneDayBasis := (baseSalary + fixedAllowance) / float64(req.WorkingDaysInMonth)
			unpaidLeaveDeduction = float64(req.UnpaidLeaveDays) * oneDayBasis
		}
	}

	// 3. Variable Allowances (Based on Attendance)
	variableAllowances := (meal + transport) * float64(req.AttendanceDays)

	// 4. Overtime (Basis: Basic + Fixed Allowances)
	hourlyRate := (baseSalary + fixedAllowance) / OvertimeHourlyDivider
	overtimePay := overtime * hourlyRate * OvertimeRate

	// 5. BPJS Calculation (Only for Regular components)
	bpjsBasis := proratedBasic + proratedFixedAllowance
	healthBasis := math.Min(bpjsBasis, MaxHealthBasis)
	jpBasis := math.Min(bpjsBasis, MaxJPBasis)

	res := PayrollResponse{
		RunType: req.RunType,
		Method:  req.Method,
	}
	res.Breakdown.Earnings.BasicSalary = proratedBasic
	res.Breakdown.Earnings.FixedAllowances = proratedFixedAllowance
	res.Breakdown.Earnings.VariableAllowances = variableAllowances
	res.Breakdown.Earnings.OvertimePay = overtimePay
	res.Breakdown.Earnings.Incentives = incentives
	res.Breakdown.Earnings.Bonus = bonus
	res.Breakdown.Earnings.THR = thr

	grossIncome := proratedBasic + proratedFixedAllowance + variableAllowances + overtimePay + incentives + bonus + thr

	// BPJS Breakdown
	res.Breakdown.Deductions.UnpaidLeaveDeduction = unpaidLeaveDeduction
	res.Breakdown.Deductions.BpjsHealthEmployee = healthBasis * BpjsHealthEmployeeRate
	res.Breakdown.Deductions.BpjsJhtEmployee = bpjsBasis * BpjsJhtEmployeeRate
	res.Breakdown.Deductions.BpjsJpEmployee = jpBasis * BpjsJpEmployeeRate

	res.Breakdown.EmployerContributions.BpjsHealthCompany = healthBasis * BpjsHealthCompanyRate
	res.Breakdown.EmployerContributions.BpjsJhtCompany = bpjsBasis * BpjsJhtCompanyRate
	res.Breakdown.EmployerContributions.BpjsJpCompany = jpBasis * BpjsJpCompanyRate
	res.Breakdown.EmployerContributions.BpjsJkk = bpjsBasis * BpjsJkkRate
	res.Breakdown.EmployerContributions.BpjsJkm = bpjsBasis * BpjsJkmRate
	res.Breakdown.EmployerContributions.TotalEmployerCost =
		res.Breakdown.EmployerContributions.BpjsHealthCompany +
			res.Breakdown.EmployerContributions.BpjsJhtCompany +
			res.Breakdown.EmployerContributions.BpjsJpCompany +
			res.Breakdown.EmployerContributions.BpjsJkk +
			res.Breakdown.EmployerContributions.BpjsJkm

	// 6. Tax Calculation (PPh 21)
	taxBruto := grossIncome +
		res.Breakdown.EmployerContributions.BpjsHealthCompany +
		res.Breakdown.EmployerContributions.BpjsJkk +
		res.Breakdown.EmployerContributions.BpjsJkm

	pph21 := s.calculatePPh21TER(req.PTKPStatus, taxBruto)

	// 7. Handle NET (Gross Up) Method
	if req.Method == model.MethodNet {
		// A. BPJS Gross-up: Add allowance to cover employee BPJS shares
		bpjsEmployeeTotal := res.Breakdown.Deductions.BpjsHealthEmployee +
			res.Breakdown.Deductions.BpjsJhtEmployee +
			res.Breakdown.Deductions.BpjsJpEmployee

		res.Breakdown.Earnings.BpjsAllowance = bpjsEmployeeTotal
		grossIncome += bpjsEmployeeTotal
		taxBruto += bpjsEmployeeTotal // BPJS Allowance increases taxable income

		// B. Tax Gross-up: Iterate to find tax allowance that covers the tax
		taxAllowance := pph21
		for i := 0; i < 5; i++ {
			newPph21 := s.calculatePPh21TER(req.PTKPStatus, taxBruto+taxAllowance)
			if math.Abs(newPph21-taxAllowance) < 1.0 {
				taxAllowance = newPph21
				break
			}
			taxAllowance = newPph21
		}
		res.Breakdown.Earnings.TaxAllowance = taxAllowance
		grossIncome += taxAllowance
		pph21 = taxAllowance
	}

	res.Breakdown.Earnings.GrossIncome = grossIncome
	res.Breakdown.Deductions.Pph21Amount = pph21

	res.Breakdown.Deductions.TotalDeductions =
		res.Breakdown.Deductions.Pph21Amount +
			res.Breakdown.Deductions.BpjsHealthEmployee +
			res.Breakdown.Deductions.BpjsJhtEmployee +
			res.Breakdown.Deductions.BpjsJpEmployee

	res.NetSalary = grossIncome - res.Breakdown.Deductions.TotalDeductions
	res.UnpaidLeaveDeduction = unpaidLeaveDeduction
	res.AttendanceDays = req.AttendanceDays
	res.WorkingDays = req.WorkingDaysInMonth
	res.UnpaidLeaveDays = req.UnpaidLeaveDays

	return res, nil
}

func (s *payrollService) calculatePPh21TER(ptkp string, bruto float64) float64 {
	category := "A"
	switch strings.ToUpper(ptkp) {
	case "TK/0", "TK/1", "K/0":
		category = "A"
	case "TK/2", "TK/3", "K/1", "K/2":
		category = "B"
	case "K/3":
		category = "C"
	}

	rate := 0.0
	switch category {
	case "A":
		if bruto <= 5400000 {
			rate = 0
		} else if bruto <= 5650000 {
			rate = 0.0025
		} else if bruto <= 5950000 {
			rate = 0.005
		} else if bruto <= 6300000 {
			rate = 0.0075
		} else if bruto <= 6750000 {
			rate = 0.01
		} else if bruto <= 7500000 {
			rate = 0.0125
		} else if bruto <= 8550000 {
			rate = 0.015
		} else if bruto <= 9650000 {
			rate = 0.0175
		} else if bruto <= 10950000 {
			rate = 0.02
		} else if bruto <= 13000000 {
			rate = 0.025
		} else {
			rate = 0.05
		}
	case "B":
		if bruto <= 6200000 {
			rate = 0
		} else if bruto <= 6500000 {
			rate = 0.0025
		} else if bruto <= 6900000 {
			rate = 0.005
		} else if bruto <= 7300000 {
			rate = 0.0075
		} else if bruto <= 7800000 {
			rate = 0.01
		} else if bruto <= 8850000 {
			rate = 0.0125
		} else if bruto <= 9800000 {
			rate = 0.015
		} else if bruto <= 10950000 {
			rate = 0.0175
		} else if bruto <= 12300000 {
			rate = 0.02
		} else if bruto <= 13000000 {
			rate = 0.025
		} else {
			rate = 0.05
		}
	default: // Category C
		if bruto <= 6600000 {
			rate = 0
		} else if bruto <= 6950000 {
			rate = 0.0025
		} else if bruto <= 7350000 {
			rate = 0.005
		} else if bruto <= 7800000 {
			rate = 0.0075
		} else if bruto <= 8350000 {
			rate = 0.01
		} else if bruto <= 9450000 {
			rate = 0.0125
		} else if bruto <= 10350000 {
			rate = 0.015
		} else if bruto <= 11350000 {
			rate = 0.0175
		} else if bruto <= 12700000 {
			rate = 0.02
		} else if bruto <= 13000000 {
			rate = 0.025
		} else {
			rate = 0.05
		}
	}

	return bruto * rate
}

func (s *payrollService) GeneratePayroll(ctx context.Context, tenantID uint, period string, runType model.PayrollRunType, method model.CalculationMethod) error {
	users, _, err := s.userRepo.FindAll(ctx, model.UserFilter{TenantID: tenantID}, []string{"position", "tenant"})
	if err != nil {
		return err
	}

	_ = s.repo.DeleteByPeriod(ctx, tenantID, period)

	for _, user := range users {
		sync, _ := s.SyncEmployeeAttendance(ctx, user.ID, period)

		profile, _ := s.profileRepo.FindByUserID(ctx, user.ID)
		ptkp := "TK/0"
		basic := user.BaseSalary
		fixed := 0.0
		meal := 0.0
		transport := 0.0

		if profile != nil {
			ptkp = string(profile.PtkpStatus)
			basic = profile.BasicSalary
			fixed = profile.FixedAllowance
			meal = profile.DailyMealAllowance
			transport = profile.DailyTransportAllowance
		}

		calcRes, _ := s.Calculate(ctx, PayrollRequest{
			UserID:                  user.ID,
			RunType:                 runType,
			Method:                  method,
			BasicSalary:             basic,
			FixedAllowances:         fixed,
			DailyMealAllowance:      meal,
			DailyTransportAllowance: transport,
			WorkingDaysInMonth:      sync.WorkingDaysInMonth,
			AttendanceDays:          sync.AttendanceDays,
			UnpaidLeaveDays:         sync.UnpaidLeaveDays,
			OvertimeHours:           sync.OvertimeHours,
			PTKPStatus:              ptkp,
			THR:                     0,
		})

		payroll := &model.Payroll{
			TenantID:             tenantID,
			UserID:               user.ID,
			Period:               period,
			RunType:              runType,
			Method:               method,
			EmployeeFullName:     user.Name,
			EmployeeID:           user.EmployeeID,
			EmployeePosition:     user.Position.Name,
			EmployeeDepartment:   user.Department,
			EmployeePtkpStatus:   ptkp,
			BasicSalary:          basic,
			FixedAllowances:      fixed,
			VariableAllowances:   calcRes.Breakdown.Earnings.VariableAllowances,
			Incentives:           0,
			Bonus:                0,
			THR:                  0,
			GrossIncome:          calcRes.Breakdown.Earnings.GrossIncome,
			Pph21Amount:          calcRes.Breakdown.Deductions.Pph21Amount,
			BpjsHealthEmployee:   calcRes.Breakdown.Deductions.BpjsHealthEmployee,
			BpjsJhtEmployee:      calcRes.Breakdown.Deductions.BpjsJhtEmployee,
			BpjsJpEmployee:       calcRes.Breakdown.Deductions.BpjsJpEmployee,
			BpjsHealthCompany:    calcRes.Breakdown.EmployerContributions.BpjsHealthCompany,
			BpjsJhtCompany:       calcRes.Breakdown.EmployerContributions.BpjsJhtCompany,
			BpjsJpCompany:        calcRes.Breakdown.EmployerContributions.BpjsJpCompany,
			BpjsJkk:              calcRes.Breakdown.EmployerContributions.BpjsJkk,
			BpjsJkm:              calcRes.Breakdown.EmployerContributions.BpjsJkm,
			TotalDeductions:      calcRes.Breakdown.Deductions.TotalDeductions,
			NetSalary:            calcRes.NetSalary,
			AttendanceDays:       sync.AttendanceDays,
			WorkingDays:          sync.WorkingDaysInMonth,
			UnpaidLeaveDays:      sync.UnpaidLeaveDays,
			UnpaidLeaveDeduction: calcRes.Breakdown.Deductions.UnpaidLeaveDeduction,
			OvertimePay:          calcRes.Breakdown.Earnings.OvertimePay,
			Status:               model.PayrollStatusDraft,
		}

		_ = s.repo.Save(ctx, payroll)
	}

	return nil
}

func (s *payrollService) GetAllPayroll(ctx context.Context, tenantID uint, period string, search string, limit, offset int) ([]PayrollResponse, int64, error) {
	payrolls, total, err := s.repo.FindAll(ctx, tenantID, period, search, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	setting, _ := s.settingRepo.FindByTenantID(ctx, tenantID)
	tenant, _ := s.tenantRepo.FindByID(ctx, tenantID)

	var responses []PayrollResponse
	for _, p := range payrolls {
		res := s.mapModelToResponse(&p, tenant, setting)
		responses = append(responses, res)
	}

	return responses, total, nil
}

func (s *payrollService) mapModelToResponse(p *model.Payroll, tenant *model.Tenant, setting *model.TenantSetting) PayrollResponse {
	res := PayrollResponse{
		ID:         p.ID,
		RunType:    p.RunType,
		Method:     p.Method,
		NetSalary:  p.NetSalary,
		PeriodText: p.Period,
		Status:     p.Status,
	}

	if tenant != nil {
		res.CompanyContext.Name = tenant.Name
	}
	if setting != nil {
		res.CompanyContext.LogoURL = setting.TenantLogo
	}

	res.User = EmployeeContext{
		FullName:   p.EmployeeFullName,
		EmployeeID: p.EmployeeID,
		Position:   p.EmployeePosition,
		Department: p.EmployeeDepartment,
		PTKPStatus: p.EmployeePtkpStatus,
	}

	// Add profile info if available
	profile, _ := s.profileRepo.FindByUserID(context.Background(), p.UserID)
	if profile != nil {
		res.User.BankName = profile.BankName
		res.User.BankAccountNumber = profile.BankAccountNumber
	}

	res.Breakdown.Earnings.BasicSalary = p.BasicSalary
	res.Breakdown.Earnings.FixedAllowances = p.FixedAllowances
	res.Breakdown.Earnings.VariableAllowances = p.VariableAllowances
	res.Breakdown.Earnings.Incentives = p.Incentives
	res.Breakdown.Earnings.Bonus = p.Bonus
	res.Breakdown.Earnings.THR = p.THR
	res.Breakdown.Earnings.OvertimePay = p.OvertimePay
	res.Breakdown.Earnings.GrossIncome = p.GrossIncome

	res.Breakdown.Deductions.Pph21Amount = p.Pph21Amount
	res.Breakdown.Deductions.UnpaidLeaveDeduction = p.UnpaidLeaveDeduction
	res.Breakdown.Deductions.BpjsHealthEmployee = p.BpjsHealthEmployee
	res.Breakdown.Deductions.BpjsJhtEmployee = p.BpjsJhtEmployee
	res.Breakdown.Deductions.BpjsJpEmployee = p.BpjsJpEmployee
	res.Breakdown.Deductions.TotalDeductions = p.TotalDeductions

	res.Breakdown.EmployerContributions.BpjsHealthCompany = p.BpjsHealthCompany
	res.Breakdown.EmployerContributions.BpjsJhtCompany = p.BpjsJhtCompany
	res.Breakdown.EmployerContributions.BpjsJpCompany = p.BpjsJpCompany
	res.Breakdown.EmployerContributions.BpjsJkk = p.BpjsJkk
	res.Breakdown.EmployerContributions.BpjsJkm = p.BpjsJkm
	res.Breakdown.EmployerContributions.TotalEmployerCost = p.BpjsHealthCompany + p.BpjsJhtCompany + p.BpjsJpCompany + p.BpjsJkk + p.BpjsJkm

	res.AttendanceDays = p.AttendanceDays
	res.WorkingDays = p.WorkingDays
	res.UnpaidLeaveDays = p.UnpaidLeaveDays
	res.UnpaidLeaveDeduction = p.UnpaidLeaveDeduction

	return res
}

func (s *payrollService) GetSummary(ctx context.Context, tenantID uint, period string) (model.PayrollSummary, error) {
	return s.repo.GetSummary(ctx, tenantID, period)
}

func (s *payrollService) PublishPayroll(ctx context.Context, tenantID uint, id uint) error {
	payroll, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if payroll.TenantID != tenantID {
		return errors.New("unauthorized")
	}

	payroll.Status = model.PayrollStatusPublished
	return s.repo.Update(ctx, payroll)
}

func (s *payrollService) GetMyPayrolls(ctx context.Context, userID uint) ([]PayrollResponse, error) {
	payrolls, err := s.repo.FindAllByUser(ctx, userID, string(model.PayrollStatusPublished))
	if err != nil {
		return nil, err
	}

	user, _ := s.userRepo.FindByID(ctx, userID, []string{"tenant.tenant_settings"})
	var tenant *model.Tenant
	var setting *model.TenantSetting
	if user != nil && user.Tenant != nil {
		tenant = user.Tenant
		setting = user.Tenant.TenantSettings
	}

	var responses []PayrollResponse
	for _, p := range payrolls {
		responses = append(responses, s.mapModelToResponse(&p, tenant, setting))
	}

	return responses, nil
}

func (s *payrollService) GetEmployeeBaseline(ctx context.Context, userID uint) (EmployeeBaselineResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID, []string{})
	if err != nil {
		return EmployeeBaselineResponse{}, err
	}

	profile, _ := s.profileRepo.FindByUserID(ctx, userID)
	ptkp := "TK/0"
	basic := user.BaseSalary
	fixed := 0.0

	if profile != nil {
		ptkp = string(profile.PtkpStatus)
		basic = profile.BasicSalary
		fixed = profile.FixedAllowance
	}

	return EmployeeBaselineResponse{
		UserID:          user.ID,
		EmployeeName:    user.Name,
		Department:      user.Department,
		PTKPStatus:      ptkp,
		BasicSalary:     basic,
		FixedAllowances: fixed,
	}, nil
}

func (s *payrollService) SyncEmployeeAttendance(ctx context.Context, userID uint, period string) (AttendanceSyncResponse, error) {
	// 1. Parse Period (YYYY-MM)
	parsedTime, err := time.Parse("2006-01", period)
	if err != nil {
		return AttendanceSyncResponse{}, errors.New("invalid period format, use YYYY-MM")
	}

	startOfMonth := time.Date(parsedTime.Year(), parsedTime.Month(), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

	// 2. Get User Info
	user, err := s.userRepo.FindByID(ctx, userID, []string{})
	if err != nil {
		return AttendanceSyncResponse{}, err
	}

	// 3. Calculate Working Days In Month
	workingDays := 0
	holidays, _ := s.hrOpsRepo.FindEvents(ctx, user.TenantID, startOfMonth.Year())
	holidayMap := make(map[string]bool)
	for _, h := range holidays {
		if h.Category == model.EventCategoryOfficeClosed {
			holidayMap[h.Date.Format("2006-01-02")] = true
		}
	}

	for d := startOfMonth; !d.After(endOfMonth); d = d.AddDate(0, 0, 1) {
		// Standard: Exclude weekends
		if d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
			continue
		}
		// Exclude Holidays
		if holidayMap[d.Format("2006-01-02")] {
			continue
		}
		workingDays++
	}

	// 4. Get Attendance Days
	counts, _ := s.attendanceRepo.GetSummaryCounts(ctx, model.AttendanceFilter{
		UserID:   userID,
		DateFrom: &startOfMonth,
		DateTo:   &endOfMonth,
	})
	attendanceDays := int(counts[model.StatusDone] + counts[model.StatusLate])

	// 5. Get Unpaid Leave Days
	leaves, _, _ := s.leaveRepo.FindAll(ctx, model.LeaveFilter{
		UserID:   userID,
		DateFrom: &startOfMonth,
		DateTo:   &endOfMonth,
		Status:   model.LeaveStatusApproved,
	}, 0, 0)

	unpaidLeaveDays := 0
	for _, l := range leaves {
		// Check if it's unpaid leave (Code: UNPAID or Name contains "Unpaid")
		if l.LeaveType != nil && (strings.ToUpper(l.LeaveType.Code) == "UNPAID" || strings.Contains(strings.ToLower(l.LeaveType.Name), "unpaid")) {
			// Calculate days within this month
			sDate := l.StartDate
			if sDate.Before(startOfMonth) {
				sDate = startOfMonth
			}
			eDate := l.EndDate
			if eDate.After(endOfMonth) {
				eDate = endOfMonth
			}

			// Add days
			for d := sDate; !d.After(eDate); d = d.AddDate(0, 0, 1) {
				if d.Weekday() != time.Saturday && d.Weekday() != time.Sunday {
					unpaidLeaveDays++
				}
			}
		}
	}

	// 6. Get Overtime Hours
	overtimes, _, _ := s.overtimeRepo.FindAll(ctx, model.OvertimeFilter{
		UserID:   userID,
		DateFrom: &startOfMonth,
		DateTo:   &endOfMonth,
		Status:   model.OvertimeStatusApproved,
	})

	totalOTHours := 0.0
	for _, ot := range overtimes {
		// Parse StartTime and EndTime (HH:mm)
		start, err1 := time.Parse("15:04", ot.StartTime)
		end, err2 := time.Parse("15:04", ot.EndTime)
		if err1 == nil && err2 == nil {
			duration := end.Sub(start).Hours()
			if duration < 0 {
				duration += 24 // Handle overnight
			}
			totalOTHours += duration
		}
	}

	return AttendanceSyncResponse{
		Period:             period,
		WorkingDaysInMonth: workingDays,
		AttendanceDays:     attendanceDays,
		UnpaidLeaveDays:    unpaidLeaveDays,
		OvertimeHours:      totalOTHours,
	}, nil
}

func (s *payrollService) SaveIndividualPayroll(ctx context.Context, tenantID uint, userID uint, req SaveIndividualPayrollRequest) error {
	user, _ := s.userRepo.FindByID(ctx, userID, []string{"position"})

	calcRes, err := s.Calculate(ctx, PayrollRequest{
		UserID:                  userID,
		RunType:                 req.RunType,
		Method:                  req.Method,
		BasicSalary:             req.BasicSalary,
		FixedAllowances:         req.FixedAllowances,
		DailyMealAllowance:      req.DailyMealAllowance,
		DailyTransportAllowance: req.DailyTransportAllowance,
		Incentives:              req.Incentives,
		Bonus:                   req.Bonus,
		THR:                     req.THR,
		AttendanceDays:          req.AttendanceDays,
		WorkingDaysInMonth:      req.WorkingDaysInMonth,
		OvertimeHours:           req.OvertimeHours,
		UnpaidLeaveDays:         req.UnpaidLeaveDays,
		PTKPStatus:              req.PTKPStatus,
	})
	if err != nil {
		return err
	}

	existing, _ := s.repo.FindByUserPeriod(ctx, userID, req.Period)

	payroll := &model.Payroll{
		TenantID:             tenantID,
		UserID:               userID,
		Period:               req.Period,
		RunType:              req.RunType,
		Method:               req.Method,
		EmployeeFullName:     user.Name,
		EmployeeID:           user.EmployeeID,
		EmployeePosition:     user.Position.Name,
		EmployeeDepartment:   user.Department,
		EmployeePtkpStatus:   req.PTKPStatus,
		BasicSalary:          req.BasicSalary,
		FixedAllowances:      req.FixedAllowances,
		VariableAllowances:   calcRes.Breakdown.Earnings.VariableAllowances,
		Incentives:           req.Incentives,
		Bonus:                req.Bonus,
		THR:                  req.THR,
		GrossIncome:          calcRes.Breakdown.Earnings.GrossIncome,
		Pph21Amount:          calcRes.Breakdown.Deductions.Pph21Amount,
		BpjsHealthEmployee:   calcRes.Breakdown.Deductions.BpjsHealthEmployee,
		BpjsJhtEmployee:      calcRes.Breakdown.Deductions.BpjsJhtEmployee,
		BpjsJpEmployee:       calcRes.Breakdown.Deductions.BpjsJpEmployee,
		BpjsHealthCompany:    calcRes.Breakdown.EmployerContributions.BpjsHealthCompany,
		BpjsJhtCompany:       calcRes.Breakdown.EmployerContributions.BpjsJhtCompany,
		BpjsJpCompany:        calcRes.Breakdown.EmployerContributions.BpjsJpCompany,
		BpjsJkk:              calcRes.Breakdown.EmployerContributions.BpjsJkk,
		BpjsJkm:              calcRes.Breakdown.EmployerContributions.BpjsJkm,
		TotalDeductions:      calcRes.Breakdown.Deductions.TotalDeductions,
		NetSalary:            calcRes.NetSalary,
		AttendanceDays:       req.AttendanceDays,
		WorkingDays:          req.WorkingDaysInMonth,
		UnpaidLeaveDays:      req.UnpaidLeaveDays,
		UnpaidLeaveDeduction: calcRes.Breakdown.Deductions.UnpaidLeaveDeduction,
		OvertimePay:          calcRes.Breakdown.Earnings.OvertimePay,
		Status:               req.Status,
	}

	if existing != nil {
		payroll.ID = existing.ID
		return s.repo.Update(ctx, payroll)
	}

	return s.repo.Save(ctx, payroll)
}

func (s *payrollService) GetUserPayrollProfile(ctx context.Context, userID uint) (*model.UserPayrollProfile, error) {
	return s.profileRepo.FindByUserID(ctx, userID)
}

func (s *payrollService) UpdateUserPayrollProfile(ctx context.Context, userID uint, req UpdateUserPayrollProfileRequest) error {
	profile, err := s.profileRepo.FindByUserID(ctx, userID)
	if err != nil {
		// If not found, create new
		profile = &model.UserPayrollProfile{UserID: userID}
	}

	profile.BankName = req.BankName
	profile.BankAccountNumber = req.BankAccountNumber
	profile.BankAccountHolder = req.BankAccountHolder
	profile.BpjsHealthNumber = req.BpjsHealthNumber
	profile.BpjsEmploymentNumber = req.BpjsEmploymentNumber
	profile.NpwpNumber = req.NpwpNumber
	profile.PtkpStatus = req.PtkpStatus
	profile.BasicSalary = req.BasicSalary
	profile.FixedAllowance = req.FixedAllowance
	profile.DailyMealAllowance = req.DailyMealAllowance
	profile.DailyTransportAllowance = req.DailyTransportAllowance

	return s.profileRepo.Upsert(ctx, profile)
}

func (s *payrollService) GetMyPayrollProfile(ctx context.Context, userID uint) (*model.UserPayrollProfile, error) {
	return s.profileRepo.FindByUserID(ctx, userID)
}

func (s *payrollService) GetMySlip(ctx context.Context, userID uint, period string) (*PayrollResponse, error) {
	p, err := s.repo.FindByUserPeriod(ctx, userID, period)
	if err != nil {
		return nil, err
	}

	user, _ := s.userRepo.FindByID(ctx, userID, []string{"tenant.tenant_settings"})
	var tenant *model.Tenant
	var setting *model.TenantSetting
	if user != nil && user.Tenant != nil {
		tenant = user.Tenant
		setting = user.Tenant.TenantSettings
	}

	res := s.mapModelToResponse(p, tenant, setting)
	return &res, nil
}
