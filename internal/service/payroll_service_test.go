package service

import (
	"context"
	"go-attendance-api/internal/model"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculate_StatelessTER(t *testing.T) {
	// Initialize service with nil repos since they won't be called for UserID = 0
	s := NewPayrollService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	ctx := context.Background()

	t.Run("Standard TER A Calculation - Under PTKP threshold", func(t *testing.T) {
		req := PayrollRequest{
			UserID:             0,
			RunType:            model.RunTypeRegular,
			Method:             model.MethodGross,
			BasicSalary:        5000000,
			PTKPStatus:         "TK/0", // Category A
			WorkingDaysInMonth: 20,
			AttendanceDays:     20,
		}

		res, err := s.Calculate(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, 5000000.0, res.Breakdown.Earnings.GrossIncome)
		assert.Equal(t, 0.0, res.Breakdown.Deductions.Pph21Amount) // Bruto 5M <= 5.4M -> 0%
	})

	t.Run("Standard TER A Calculation - In Bracket", func(t *testing.T) {
		req := PayrollRequest{
			UserID:             0,
			RunType:            model.RunTypeRegular,
			Method:             model.MethodGross,
			BasicSalary:        10000000,
			PTKPStatus:         "TK/0", // Category A
			WorkingDaysInMonth: 20,
			AttendanceDays:     20,
		}

		res, err := s.Calculate(ctx, req)
		assert.NoError(t, err)
		// Bruto including employer BPJS (Health 4% = 400k, JKK 0.24% = 24k, JKM 0.3% = 30k) = 10,454,000
		// Category A bracket > 9,650,000 <= 10,950,000 -> 2%
		// 10,454,000 * 2% = 209,080
		assert.Equal(t, 209080.0, res.Breakdown.Deductions.Pph21Amount)
	})

	t.Run("TER A Gross-Up (Net Method)", func(t *testing.T) {
		req := PayrollRequest{
			UserID:             0,
			RunType:            model.RunTypeRegular,
			Method:             model.MethodNet,
			BasicSalary:        9800000,
			PTKPStatus:         "TK/0", // Category A
			WorkingDaysInMonth: 20,
			AttendanceDays:     20,
		}

		res, err := s.Calculate(ctx, req)
		assert.NoError(t, err)

		// Verification: check that Take Home Pay (NetSalary) is equal to Gross income minus deductions,
		// and that NetSalary is close to the expected basic salary + allowances minus employee BPJS.
		// For Net method, tax_allowance should match PPh 21 exactly
		assert.Equal(t, res.Breakdown.Earnings.TaxAllowance, res.Breakdown.Deductions.Pph21Amount)
	})

	t.Run("December Pasal 17 Calculation", func(t *testing.T) {
		req := PayrollRequest{
			UserID:             0,
			RunType:            model.RunTypeRegular,
			Method:             model.MethodGross,
			BasicSalary:        10000000,
			PTKPStatus:         "TK/0", // Category A (Annual PTKP: 54,000,000)
			WorkingDaysInMonth: 20,
			AttendanceDays:     20,
			Period:             "2026-12", // December period -> triggers Pasal 17
		}

		res, err := s.Calculate(ctx, req)
		assert.NoError(t, err)

		// Gross monthly including employer BPJS = 10,454,000
		// Gross annual = 125,448,000
		// Biaya jabatan = 125,448,000 * 5% = 6,272,400 -> capped at 6,000,000
		// Employee BPJS (calculated on 10,000,000 basic):
		// JHT: 2% -> 200,000 * 12 = 2,400,000
		// JP: 1% -> 100,000 * 12 = 1,200,000
		// Net Annual = 125,448,000 - 6,000,000 - 2,400,000 - 1,200,000 = 115,848,000
		// PKP = 115,848,000 - 54,000,000 = 61,848,000
		// PPh 21 Pasal 17 Annual = (60,000,000 * 5%) + (1,848,000 * 15%) = 3,000,000 + 277,200 = 3,277,200
		// PPh 21 Monthly (Dec) = 3,277,200 / 12 = 273,100
		assert.Equal(t, 273100.0, res.Breakdown.Deductions.Pph21Amount)
	})
}
