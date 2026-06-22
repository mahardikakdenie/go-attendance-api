package service

import (
	"context"
	"errors"
	"fmt"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
	"go-attendance-api/internal/utils"
	"math"
	"strings"
	"time"
)

type PayrollService interface {
	Calculate(ctx context.Context, req PayrollRequest) (PayrollResponse, error)
	BulkGeneratePayroll(ctx context.Context, tenantID uint, req BulkGenerateRequest) (int, error)
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
	UserID                   uint                    `json:"user_id"`
	RunType                  model.PayrollRunType    `json:"run_type"`
	Method                   model.CalculationMethod `json:"method"`
	BasicSalary              float64                 `json:"basic_salary"`
	FixedAllowances          float64                 `json:"fixed_allowances"`
	DailyMealAllowance       float64                 `json:"daily_meal_allowance"`
	DailyTransportAllowance  float64                 `json:"daily_transport_allowance"`
	MealAllowanceType        string                  `json:"meal_allowance_type"`
	TransportAllowanceType   string                  `json:"transport_allowance_type"`
	VariableAllowances       float64                 `json:"variable_allowances"`
	CustomVariableAllowances []model.CustomAllowance `json:"custom_variable_allowances"`
	Incentives               float64                 `json:"incentives"`
	Bonus                    float64                 `json:"bonus"`
	THR                      float64                 `json:"thr"`
	CalculateTHR             bool                    `json:"calculate_thr"`
	AttendanceDays           int                     `json:"attendance_days"`
	WorkingDaysInMonth       int                     `json:"working_days_in_month"`
	OvertimeHours            float64                 `json:"overtime_hours"`
	UnpaidLeaveDays          int                     `json:"unpaid_leave_days"`
	PTKPStatus               string                  `json:"ptkp_status"`
	Period                   string                  `json:"period"`
}

func (s *payrollService) BulkGeneratePayroll(ctx context.Context, tenantID uint, req BulkGenerateRequest) (int, error) {
	// Set defaults
	if req.RunType == "" {
		req.RunType = model.RunTypeRegular
	}
	if req.Method == "" {
		req.Method = model.MethodGross
	}

	// 1. Fetch Users
	filter := model.UserFilter{
		TenantID: tenantID,
		IDs:      req.UserIDs,
	}
	users, _, err := s.userRepo.FindAll(ctx, filter, []string{"position", "tenant"})
	if err != nil {
		return 0, err
	}

	successCount := 0
	for _, user := range users {
		// Sync attendance for the period
		sync, err := s.SyncEmployeeAttendance(ctx, user.ID, req.Period)
		if err != nil {
			continue // Skip if sync fails for one user
		}

		// Get Payroll Profile
		profile, _ := s.profileRepo.FindByUserID(ctx, user.ID)
		ptkp := "TK/0"
		basic := user.BaseSalary
		fixed := 0.0
		meal := 0.0
		transport := 0.0

		mealType := "variable"
		transportType := "variable"
		if profile != nil {
			ptkp = string(profile.PtkpStatus)
			basic = profile.BasicSalary
			fixed = profile.FixedAllowance
			meal = profile.DailyMealAllowance
			transport = profile.DailyTransportAllowance
			if profile.MealAllowanceType != "" {
				mealType = profile.MealAllowanceType
			}
			if profile.TransportAllowanceType != "" {
				transportType = profile.TransportAllowanceType
			}
		}

		// Calculate Payroll
		calcRes, err := s.Calculate(ctx, PayrollRequest{
			UserID:                   user.ID,
			RunType:                  req.RunType,
			Method:                   req.Method,
			BasicSalary:              basic,
			FixedAllowances:          fixed,
			DailyMealAllowance:       meal,
			DailyTransportAllowance:  transport,
			MealAllowanceType:        mealType,
			TransportAllowanceType:   transportType,
			Incentives:               req.Incentives,
			Bonus:                    req.Bonus,
			CustomVariableAllowances: req.CustomVariableAllowances,
			CalculateTHR:             req.RunType == model.RunTypeTHR || req.RunType == model.RunTypeAll,
			WorkingDaysInMonth:       sync.WorkingDaysInMonth,
			AttendanceDays:           sync.AttendanceDays,
			UnpaidLeaveDays:          sync.UnpaidLeaveDays,
			OvertimeHours:            sync.OvertimeHours,
			PTKPStatus:               ptkp,
			Period:                   req.Period,
		})
		if err != nil {
			continue
		}

		posName := "-"
		if user.Position != nil {
			posName = user.Position.Name
		}

		// Prepare Payroll Model
		payroll := &model.Payroll{
			TenantID:                 tenantID,
			UserID:                   user.ID,
			Period:                   req.Period,
			RunType:                  req.RunType,
			Method:                   req.Method,
			EmployeeFullName:         user.Name,
			EmployeeID:               user.EmployeeID,
			EmployeePosition:         posName,
			EmployeeDepartment:       user.Department,
			EmployeePtkpStatus:       ptkp,
			BasicSalary:              basic,
			FixedAllowances:          fixed,
			VariableAllowances:       calcRes.Breakdown.Earnings.VariableAllowances,
			CustomVariableAllowances: calcRes.Breakdown.Earnings.CustomVariableAllowances,
			Incentives:               calcRes.Breakdown.Earnings.Incentives,
			Bonus:                    calcRes.Breakdown.Earnings.Bonus,
			THR:                      calcRes.Breakdown.Earnings.THR,
			GrossIncome:              calcRes.Breakdown.Earnings.GrossIncome,
			Pph21Amount:              calcRes.Breakdown.Deductions.Pph21Amount,
			BpjsHealthEmployee:       calcRes.Breakdown.Deductions.BpjsHealthEmployee,
			BpjsJhtEmployee:          calcRes.Breakdown.Deductions.BpjsJhtEmployee,
			BpjsJpEmployee:           calcRes.Breakdown.Deductions.BpjsJpEmployee,
			BpjsHealthCompany:        calcRes.Breakdown.EmployerContributions.BpjsHealthCompany,
			BpjsJhtCompany:           calcRes.Breakdown.EmployerContributions.BpjsJhtCompany,
			BpjsJpCompany:            calcRes.Breakdown.EmployerContributions.BpjsJpCompany,
			BpjsJkk:                  calcRes.Breakdown.EmployerContributions.BpjsJkk,
			BpjsJkm:                  calcRes.Breakdown.EmployerContributions.BpjsJkm,
			TotalDeductions:          calcRes.Breakdown.Deductions.TotalDeductions,
			NetSalary:                calcRes.NetSalary,
			AttendanceDays:           sync.AttendanceDays,
			WorkingDays:              sync.WorkingDaysInMonth,
			UnpaidLeaveDays:          sync.UnpaidLeaveDays,
			UnpaidLeaveDeduction:     calcRes.Breakdown.Deductions.UnpaidLeaveDeduction,
			OvertimePay:              calcRes.Breakdown.Earnings.OvertimePay,
			Status:                   model.PayrollStatusDraft,
		}

		// Upsert (Check if draft already exists)
		existing, _ := s.repo.FindByUserPeriod(ctx, user.ID, req.Period)
		if existing != nil {
			if existing.Status == model.PayrollStatusDraft {
				payroll.ID = existing.ID
				_ = s.repo.Update(ctx, payroll)
				successCount++
			}
			// If already published, we don't overwrite
		} else {
			_ = s.repo.Save(ctx, payroll)
			successCount++
		}
	}

	return successCount, nil
}

type BulkGenerateRequest struct {
	Period                   string                  `json:"period" binding:"required"`
	UserIDs                  []uint                  `json:"user_ids"`
	RunType                  model.PayrollRunType    `json:"run_type"`
	Method                   model.CalculationMethod `json:"method"`
	Incentives               float64                 `json:"incentives"`
	Bonus                    float64                 `json:"bonus"`
	CustomVariableAllowances []model.CustomAllowance `json:"custom_variable_allowances"`
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
		BasicSalary              float64                 `json:"basic_salary"`
		FixedAllowances          float64                 `json:"fixed_allowances"`
		VariableAllowances       float64                 `json:"variable_allowances"`
		CustomVariableAllowances []model.CustomAllowance `json:"custom_variable_allowances,omitempty"`
		OvertimePay              float64                 `json:"overtime_pay"`
		Incentives               float64                 `json:"incentives"`
		Bonus                    float64                 `json:"bonus"`
		THR                      float64                 `json:"thr"`
		TaxAllowance             float64                 `json:"tax_allowance,omitempty"`  // For Net/Gross-up
		BpjsAllowance            float64                 `json:"bpjs_allowance,omitempty"` // For Net/Gross-up
		GrossIncome              float64                 `json:"gross_income"`
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
	Period                   string                  `json:"period" binding:"required"`
	RunType                  model.PayrollRunType    `json:"run_type"`
	Method                   model.CalculationMethod `json:"method"`
	BasicSalary              float64                 `json:"basic_salary"`
	FixedAllowances          float64                 `json:"fixed_allowances"`
	DailyMealAllowance       float64                 `json:"daily_meal_allowance"`
	DailyTransportAllowance  float64                 `json:"daily_transport_allowance"`
	MealAllowanceType        string                  `json:"meal_allowance_type"`
	TransportAllowanceType   string                  `json:"transport_allowance_type"`
	VariableAllowances       float64                 `json:"variable_allowances"`
	CustomVariableAllowances []model.CustomAllowance `json:"custom_variable_allowances"`
	Incentives               float64                 `json:"incentives"`
	Bonus                    float64                 `json:"bonus"`
	THR                      float64                 `json:"thr"`
	AttendanceDays           int                     `json:"attendance_days"`
	WorkingDaysInMonth       int                     `json:"working_days_in_month"`
	OvertimeHours            float64                 `json:"overtime_hours"`
	UnpaidLeaveDays          int                     `json:"unpaid_leave_days"`
	PTKPStatus               string                  `json:"ptkp_status"`
	Status                   model.PayrollStatus     `json:"status"`
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
	MealAllowanceType       string           `json:"meal_allowance_type"`
	TransportAllowanceType  string           `json:"transport_allowance_type"`
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
	notifService   NotificationService
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
	notifService NotificationService,
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
		notifService:   notifService,
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

	// 🆕 Auto-enable CalculateTHR for THR-related runs if not explicitly false
	if (req.RunType == model.RunTypeTHR || req.RunType == model.RunTypeAll) && req.THR == 0 {
		req.CalculateTHR = true
	}

	res := PayrollResponse{
		RunType: req.RunType,
		Method:  req.Method,
	}

	// If profile exists, use it as baseline
	var joinDate time.Time
	var tenantSettings *model.TenantSetting
	if req.UserID != 0 {
		user, _ := s.userRepo.FindByID(ctx, req.UserID, []string{"tenant.tenant_settings", "position"})
		if user != nil {
			joinDate = user.CreatedAt
			if req.BasicSalary == 0 {
				req.BasicSalary = user.BaseSalary
			}
 
			// 🆕 Populate Response Context
			if user.Tenant != nil {
				tenantSettings = user.Tenant.TenantSettings
				res.CompanyContext.Name = user.Tenant.Name
				if user.Tenant.TenantSettings != nil {
					res.CompanyContext.LogoURL = user.Tenant.TenantSettings.TenantLogo
				}
			}

			res.User = EmployeeContext{
				FullName:   user.Name,
				EmployeeID: user.EmployeeID,
				Department: user.Department,
			}
			if user.Position != nil {
				res.User.Position = user.Position.Name
			}
		}

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
			if req.MealAllowanceType == "" {
				req.MealAllowanceType = profile.MealAllowanceType
			}
			if req.TransportAllowanceType == "" {
				req.TransportAllowanceType = profile.TransportAllowanceType
			}

			// 🆕 Populate Bank & PTKP Info
			res.User.BankName = profile.BankName
			res.User.BankAccountNumber = profile.BankAccountNumber
			res.User.PTKPStatus = string(profile.PtkpStatus)
		}
	}

	if req.MealAllowanceType == "" {
		req.MealAllowanceType = "variable"
	}
	if req.TransportAllowanceType == "" {
		req.TransportAllowanceType = "variable"
	}

	if req.MealAllowanceType == "fixed" {
		req.FixedAllowances += req.DailyMealAllowance
		req.DailyMealAllowance = 0
	}
	if req.TransportAllowanceType == "fixed" {
		req.FixedAllowances += req.DailyTransportAllowance
		req.DailyTransportAllowance = 0
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
		if req.CalculateTHR && req.THR == 0 && !joinDate.IsZero() {
			tenureMonths := s.calculateTenureMonths(joinDate)
			if tenureMonths >= 12 {
				thr = req.BasicSalary + req.FixedAllowances
			} else if tenureMonths > 0 {
				thr = (float64(tenureMonths) / 12.0) * (req.BasicSalary + req.FixedAllowances)
			}
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
		if req.CalculateTHR && req.THR == 0 && !joinDate.IsZero() {
			tenureMonths := s.calculateTenureMonths(joinDate)
			if tenureMonths >= 12 {
				thr = req.BasicSalary + req.FixedAllowances
			} else if tenureMonths > 0 {
				thr = (float64(tenureMonths) / 12.0) * (req.BasicSalary + req.FixedAllowances)
			}
		} else {
			thr = req.THR
		}
	}

	// 2. Prorate Calculation (Only for regular components)
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
	var calculatedCustomAllowances []model.CustomAllowance

	if len(req.CustomVariableAllowances) > 0 {
		customSum := 0.0
		for _, ca := range req.CustomVariableAllowances {
			// Multiply daily rate by attendance days as requested
			itemTotal := ca.Amount * float64(req.AttendanceDays)
			calculatedCustomAllowances = append(calculatedCustomAllowances, model.CustomAllowance{
				Name:   ca.Name,
				Amount: itemTotal,
			})
			customSum += itemTotal
		}
		variableAllowances = customSum
	} else if req.VariableAllowances > 0 {
		variableAllowances = req.VariableAllowances
	}

	// 4. Overtime (Basis: Basic + Fixed Allowances)
	hourlyRate := (baseSalary + fixedAllowance) / OvertimeHourlyDivider
	overtimePay := overtime * hourlyRate * OvertimeRate

	// 5. BPJS Calculation (Only for Regular components)
	bpjsBasis := proratedBasic + proratedFixedAllowance
	healthMaxBasis := MaxHealthBasis
	jpMaxBasis := MaxJPBasis
	if tenantSettings != nil {
		if tenantSettings.BpjsHealthMaxBasis > 0 {
			healthMaxBasis = tenantSettings.BpjsHealthMaxBasis
		}
		if tenantSettings.BpjsJpMaxBasis > 0 {
			jpMaxBasis = tenantSettings.BpjsJpMaxBasis
		}
	}

	healthBasis := math.Min(bpjsBasis, healthMaxBasis)
	jpBasis := math.Min(bpjsBasis, jpMaxBasis)

	res.Breakdown.Earnings.BasicSalary = proratedBasic
	res.Breakdown.Earnings.FixedAllowances = proratedFixedAllowance
	res.Breakdown.Earnings.VariableAllowances = variableAllowances
	res.Breakdown.Earnings.CustomVariableAllowances = calculatedCustomAllowances
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

	// 6. Tax Calculation (PPh 21 Bruto)
	// Base taxable income includes employer-paid Health, JKK, JKM
	taxBruto := grossIncome +
		res.Breakdown.EmployerContributions.BpjsHealthCompany +
		res.Breakdown.EmployerContributions.BpjsJkk +
		res.Breakdown.EmployerContributions.BpjsJkm

	isDecember := strings.HasSuffix(req.Period, "-12")
	var pph21 float64
	if isDecember {
		pph21 = s.calculatePPh21Pasal17(req.PTKPStatus, taxBruto, res.Breakdown.Deductions.BpjsJhtEmployee, res.Breakdown.Deductions.BpjsJpEmployee)
	} else {
		pph21 = s.calculatePPh21TER(req.PTKPStatus, taxBruto)
	}

	// 7. Handle NET (Gross Up) Method
	if req.Method == model.MethodNet {
		// A. BPJS Gross-up: Add allowance to cover employee BPJS shares
		bpjsEmployeeTotal := res.Breakdown.Deductions.BpjsHealthEmployee +
			res.Breakdown.Deductions.BpjsJhtEmployee +
			res.Breakdown.Deductions.BpjsJpEmployee

		res.Breakdown.Earnings.BpjsAllowance = bpjsEmployeeTotal
		grossIncome += bpjsEmployeeTotal
		taxBruto += bpjsEmployeeTotal // BPJS Allowance is taxable

		if isDecember {
			// B. Tax Gross-up Iteration for Pasal 17 (Continuous function)
			taxAllowance := pph21
			for i := 0; i < 10; i++ {
				newPph21 := s.calculatePPh21Pasal17(req.PTKPStatus, taxBruto+taxAllowance, res.Breakdown.Deductions.BpjsJhtEmployee, res.Breakdown.Deductions.BpjsJpEmployee)
				if math.Abs(newPph21-taxAllowance) < 1.0 {
					taxAllowance = newPph21
					break
				}
				taxAllowance = (taxAllowance + newPph21) / 2
			}
			pph21 = s.calculatePPh21Pasal17(req.PTKPStatus, taxBruto+taxAllowance, res.Breakdown.Deductions.BpjsJhtEmployee, res.Breakdown.Deductions.BpjsJpEmployee)
			res.Breakdown.Earnings.TaxAllowance = taxAllowance
			grossIncome += taxAllowance
		} else {
			// B. Mathematically exact Tax Gross-up for TER
			taxAllowance, _ := s.calculateGrossUpTER(req.PTKPStatus, taxBruto)
			pph21 = taxAllowance
			res.Breakdown.Earnings.TaxAllowance = taxAllowance
			grossIncome += taxAllowance
		}
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

type TERBracket struct {
	MaxBruto float64
	Rate     float64
}

var terA = []TERBracket{
	{5400000, 0.0},
	{5650000, 0.0025},
	{5950000, 0.005},
	{6300000, 0.0075},
	{6750000, 0.01},
	{7500000, 0.0125},
	{8550000, 0.015},
	{9650000, 0.0175},
	{10950000, 0.02},
	{13000000, 0.025},
	{15000000, 0.03},
	{20000000, 0.04},
	{25000000, 0.05},
	{30000000, 0.06},
	{35000000, 0.07},
	{40000000, 0.08},
	{45000000, 0.09},
	{54000000, 0.10},
	{68000000, 0.11},
	{83000000, 0.12},
	{97000000, 0.13},
	{118000000, 0.14},
	{145000000, 0.15},
	{180000000, 0.16},
	{230000000, 0.17},
	{290000000, 0.18},
	{360000000, 0.19},
	{440000000, 0.20},
	{540000000, 0.21},
	{660000000, 0.22},
	{800000000, 0.23},
	{1000000000, 0.24},
	{1400000000, 0.25},
	{math.MaxFloat64, 0.34},
}

var terB = []TERBracket{
	{6200000, 0.0},
	{6500000, 0.0025},
	{6850000, 0.005},
	{7300000, 0.0075},
	{7800000, 0.01},
	{8850000, 0.0125},
	{9800000, 0.015},
	{10950000, 0.0175},
	{12200000, 0.02},
	{14000000, 0.025},
	{16000000, 0.03},
	{21000000, 0.04},
	{26000000, 0.05},
	{31000000, 0.06},
	{36000000, 0.07},
	{41000000, 0.08},
	{47000000, 0.09},
	{56000000, 0.10},
	{70000000, 0.11},
	{85000000, 0.12},
	{100000000, 0.13},
	{121000000, 0.14},
	{149000000, 0.15},
	{184000000, 0.16},
	{234000000, 0.17},
	{293000000, 0.18},
	{363000000, 0.19},
	{441000000, 0.20},
	{541000000, 0.21},
	{661000000, 0.22},
	{802000000, 0.23},
	{1003000000, 0.24},
	{1405000000, 0.25},
	{math.MaxFloat64, 0.34},
}

var terC = []TERBracket{
	{6600000, 0.0},
	{6950000, 0.0025},
	{7350000, 0.005},
	{7800000, 0.0075},
	{8350000, 0.01},
	{9450000, 0.0125},
	{10350000, 0.015},
	{11350000, 0.0175},
	{12700000, 0.02},
	{14200000, 0.025},
	{16300000, 0.03},
	{21100000, 0.04},
	{26200000, 0.05},
	{31100000, 0.06},
	{35800000, 0.07},
	{40500000, 0.08},
	{45600000, 0.09},
	{55500000, 0.10},
	{70400000, 0.11},
	{85100000, 0.12},
	{100200000, 0.13},
	{120000000, 0.14},
	{147700000, 0.15},
	{181700000, 0.16},
	{231400000, 0.17},
	{290100000, 0.18},
	{359700000, 0.19},
	{437900000, 0.20},
	{538200000, 0.21},
	{657300000, 0.22},
	{796200000, 0.23},
	{995800000, 0.24},
	{1395200000, 0.25},
	{math.MaxFloat64, 0.34},
}

func (s *payrollService) getTERRate(ptkp string, bruto float64) float64 {
	category := "A"
	switch strings.ToUpper(ptkp) {
	case "TK/0", "TK/1", "K/0":
		category = "A"
	case "TK/2", "TK/3", "K/1", "K/2":
		category = "B"
	case "K/3":
		category = "C"
	}

	var brackets []TERBracket
	switch category {
	case "A":
		brackets = terA
	case "B":
		brackets = terB
	default:
		brackets = terC
	}

	for _, b := range brackets {
		if bruto <= b.MaxBruto {
			return b.Rate
		}
	}
	return 0.34
}

func (s *payrollService) calculatePPh21TER(ptkp string, bruto float64) float64 {
	rate := s.getTERRate(ptkp, bruto)
	return bruto * rate
}

func (s *payrollService) calculateGrossUpTER(ptkp string, taxBrutoBase float64) (float64, float64) {
	category := "A"
	switch strings.ToUpper(ptkp) {
	case "TK/0", "TK/1", "K/0":
		category = "A"
	case "TK/2", "TK/3", "K/1", "K/2":
		category = "B"
	case "K/3":
		category = "C"
	}

	var brackets []TERBracket
	switch category {
	case "A":
		brackets = terA
	case "B":
		brackets = terB
	default:
		brackets = terC
	}

	var prevMax float64 = 0.0
	for _, b := range brackets {
		if b.Rate >= 1.0 {
			continue
		}
		finalGross := taxBrutoBase / (1.0 - b.Rate)
		if finalGross > prevMax && finalGross <= b.MaxBruto {
			taxAllowance := finalGross - taxBrutoBase
			return taxAllowance, b.Rate
		}
		prevMax = b.MaxBruto
	}

	highestRate := brackets[len(brackets)-1].Rate
	finalGross := taxBrutoBase / (1.0 - highestRate)
	return finalGross - taxBrutoBase, highestRate
}

func getPTKPValue(ptkp string) float64 {
	switch strings.ToUpper(ptkp) {
	case "TK/0":
		return 54000000
	case "TK/1", "K/0":
		return 58500000
	case "TK/2", "K/1":
		return 63000000
	case "TK/3", "K/2":
		return 67500000
	case "K/3":
		return 72000000
	default:
		return 54000000
	}
}

func (s *payrollService) calculatePPh21Pasal17(ptkp string, monthlyGross float64, monthlyJHT float64, monthlyJP float64) float64 {
	annualGross := monthlyGross * 12

	biayaJabatan := annualGross * 0.05
	if biayaJabatan > 6000000 {
		biayaJabatan = 6000000
	}

	annualJHT := monthlyJHT * 12
	annualJP := monthlyJP * 12

	netAnnual := annualGross - biayaJabatan - annualJHT - annualJP
	ptkpVal := getPTKPValue(ptkp)

	pkp := netAnnual - ptkpVal
	if pkp <= 0 {
		return 0
	}

	tax := 0.0
	if pkp <= 60000000 {
		tax = pkp * 0.05
	} else if pkp <= 250000000 {
		tax = (60000000 * 0.05) + ((pkp - 60000000) * 0.15)
	} else if pkp <= 500000000 {
		tax = (60000000 * 0.05) + (190000000 * 0.15) + ((pkp - 250000000) * 0.25)
	} else if pkp <= 5000000000 {
		tax = (60000000 * 0.05) + (190000000 * 0.15) + (250000000 * 0.25) + ((pkp - 500000000) * 0.30)
	} else {
		tax = (60000000 * 0.05) + (190000000 * 0.15) + (250000000 * 0.25) + (4500000000 * 0.30) + ((pkp - 5000000000) * 0.35)
	}

	return tax / 12
}

func (s *payrollService) calculateTenureMonths(joinDate time.Time) int {
	now := utils.Now()
	years := now.Year() - joinDate.Year()
	months := int(now.Month()) - int(joinDate.Month())
	totalMonths := years*12 + months
	if totalMonths < 0 {
		return 0
	}
	return totalMonths
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

		mealType := "variable"
		transportType := "variable"
		if profile != nil {
			ptkp = string(profile.PtkpStatus)
			basic = profile.BasicSalary
			fixed = profile.FixedAllowance
			meal = profile.DailyMealAllowance
			transport = profile.DailyTransportAllowance
			if profile.MealAllowanceType != "" {
				mealType = profile.MealAllowanceType
			}
			if profile.TransportAllowanceType != "" {
				transportType = profile.TransportAllowanceType
			}
		}

		calcRes, _ := s.Calculate(ctx, PayrollRequest{
			UserID:                  user.ID,
			RunType:                 runType,
			Method:                  method,
			BasicSalary:             basic,
			FixedAllowances:         fixed,
			DailyMealAllowance:      meal,
			DailyTransportAllowance: transport,
			MealAllowanceType:       mealType,
			TransportAllowanceType:  transportType,
			WorkingDaysInMonth:      sync.WorkingDaysInMonth,
			AttendanceDays:          sync.AttendanceDays,
			UnpaidLeaveDays:         sync.UnpaidLeaveDays,
			OvertimeHours:           sync.OvertimeHours,
			PTKPStatus:              ptkp,
			CalculateTHR:            runType == model.RunTypeTHR || runType == model.RunTypeAll,
			Period:                  period,
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
			Incentives:           calcRes.Breakdown.Earnings.Incentives,
			Bonus:                calcRes.Breakdown.Earnings.Bonus,
			THR:                  calcRes.Breakdown.Earnings.THR,
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
	if err := s.repo.Update(ctx, payroll); err != nil {
		return err
	}

	// NOTIFICATION
	s.notifService.SendNotification(ctx, tenantID, payroll.UserID, "Payslip Published", fmt.Sprintf("Your payslip for period %s has been published", payroll.Period), model.NotificationTypePayroll)

	return nil
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
	parsedTime, err := utils.ParseTimeWIB("2006-01", period)
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
		start, err1 := utils.ParseTimeWIB("15:04", ot.StartTime)
		end, err2 := utils.ParseTimeWIB("15:04", ot.EndTime)
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
		UserID:                   userID,
		RunType:                  req.RunType,
		Method:                   req.Method,
		BasicSalary:              req.BasicSalary,
		FixedAllowances:          req.FixedAllowances,
		DailyMealAllowance:       req.DailyMealAllowance,
		DailyTransportAllowance:  req.DailyTransportAllowance,
		MealAllowanceType:        req.MealAllowanceType,
		TransportAllowanceType:   req.TransportAllowanceType,
		VariableAllowances:       req.VariableAllowances,
		CustomVariableAllowances: req.CustomVariableAllowances,
		Incentives:               req.Incentives,
		Bonus:                    req.Bonus,
		THR:                      req.THR,
		CalculateTHR:             (req.RunType == model.RunTypeTHR || req.RunType == model.RunTypeAll) && req.THR == 0,
		AttendanceDays:           req.AttendanceDays,
		WorkingDaysInMonth:       req.WorkingDaysInMonth,
		OvertimeHours:            req.OvertimeHours,
		UnpaidLeaveDays:          req.UnpaidLeaveDays,
		PTKPStatus:               req.PTKPStatus,
		Period:                   req.Period,
	})
	if err != nil {
		return err
	}

	existing, _ := s.repo.FindByUserPeriod(ctx, userID, req.Period)

	posName := "-"
	if user.Position != nil {
		posName = user.Position.Name
	}

	payroll := &model.Payroll{
		TenantID:                 tenantID,
		UserID:                   userID,
		Period:                   req.Period,
		RunType:                  req.RunType,
		Method:                   req.Method,
		EmployeeFullName:         user.Name,
		EmployeeID:               user.EmployeeID,
		EmployeePosition:         posName,
		EmployeeDepartment:       user.Department,
		EmployeePtkpStatus:       req.PTKPStatus,
		BasicSalary:              req.BasicSalary,
		FixedAllowances:          req.FixedAllowances,
		VariableAllowances:       calcRes.Breakdown.Earnings.VariableAllowances,
		CustomVariableAllowances: calcRes.Breakdown.Earnings.CustomVariableAllowances,
		Incentives:               req.Incentives,
		Bonus:                    req.Bonus,
		THR:                      req.THR,
		GrossIncome:              calcRes.Breakdown.Earnings.GrossIncome,
		Pph21Amount:              calcRes.Breakdown.Deductions.Pph21Amount,
		BpjsHealthEmployee:       calcRes.Breakdown.Deductions.BpjsHealthEmployee,
		BpjsJhtEmployee:          calcRes.Breakdown.Deductions.BpjsJhtEmployee,
		BpjsJpEmployee:           calcRes.Breakdown.Deductions.BpjsJpEmployee,
		BpjsHealthCompany:        calcRes.Breakdown.EmployerContributions.BpjsHealthCompany,
		BpjsJhtCompany:           calcRes.Breakdown.EmployerContributions.BpjsJhtCompany,
		BpjsJpCompany:            calcRes.Breakdown.EmployerContributions.BpjsJpCompany,
		BpjsJkk:                  calcRes.Breakdown.EmployerContributions.BpjsJkk,
		BpjsJkm:                  calcRes.Breakdown.EmployerContributions.BpjsJkm,
		TotalDeductions:          calcRes.Breakdown.Deductions.TotalDeductions,
		NetSalary:                calcRes.NetSalary,
		AttendanceDays:           req.AttendanceDays,
		WorkingDays:              req.WorkingDaysInMonth,
		UnpaidLeaveDays:          req.UnpaidLeaveDays,
		UnpaidLeaveDeduction:     calcRes.Breakdown.Deductions.UnpaidLeaveDeduction,
		OvertimePay:              calcRes.Breakdown.Earnings.OvertimePay,
		Status:                   req.Status,
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
	profile.MealAllowanceType = req.MealAllowanceType
	profile.TransportAllowanceType = req.TransportAllowanceType
 
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
