package attendance

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go-attendance-api/internal/model"
	"go-attendance-api/internal/utils"
)

// GetAllData mengambil daftar absensi berdasarkan filter, mendukung paginasi dan inklusi relasi.
// Fungsi ini juga mengamankan data dengan membatasi akses berdasarkan tenantID (Tenant Isolation).
//
// Parameters:
//   - ctx: context.Context.
//   - requesterID: uint, ID user yang meminta data (untuk scoping).
//   - filter: model.AttendanceFilter, kriteria pencarian (tanggal, status, dll).
//   - includes: []string, relasi yang ingin di-load (misal: "user", "tenant").
//   - limit, offset: int, untuk paginasi.
//
// Returns/Impact:
//   - []model.AttendanceResponse: list data absensi.
//   - int64: total data keseluruhan.
//   - error: jika query database gagal.
func (s *attendanceService) GetAllData(
	ctx context.Context,
	requesterID uint,
	filter model.AttendanceFilter,
	includes []string,
	limit, offset int,
) ([]model.AttendanceResponse, int64, error) {
	includes = filterAttendanceIncludes(includes)

	// Apply Hierarchical Scoping & Tenant Isolation
	if requesterID != 0 {
		// Fetch requester details to enforce TenantID isolation
		user, err := s.userRepo.FindByID(ctx, requesterID, []string{})
		if err != nil {
			return nil, 0, errors.New("requester not found")
		}

		// Enforce TenantID from the requester to isolate data per tenant.
		// This prevents cross-tenant data access even for superadmins when calling tenant-scoped endpoints.
		filter.TenantID = user.TenantID

		allowedRoleIDs, _ := s.userService.GetAllowedRoleIDs(ctx, requesterID)
		filter.AllowedRoleIDs = allowedRoleIDs
	}

	data, total, err := s.repo.FindAll(ctx, filter, includes, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	var responses []model.AttendanceResponse
	for _, a := range data {
		responses = append(responses, applyPreloads(&a, includes))
	}

	return responses, total, nil
}

// GetSummary menghitung ringkasan statistik absensi (tepat waktu, telat, total) dalam rentang tanggal tertentu,
// dan membandingkannya dengan periode sebelumnya untuk persentase kenaikan/penurunan.
//
// Parameters:
//   - ctx: context.Context.
//   - tenantID: uint, filter wajib agar data tidak bocor antar tenant.
//   - filter: model.AttendanceFilter, kriteria tanggal.
//
// Returns/Impact:
//   - model.AttendanceSummaryResponse: objek ringkasan berisi total, ontime, late, dan diff (persentase).
//   - error: jika query count di DB gagal.
func (s *attendanceService) GetSummary(ctx context.Context, tenantID uint, filter model.AttendanceFilter) (model.AttendanceSummaryResponse, error) {
	filter.TenantID = tenantID

	currentCounts, err := s.repo.GetSummaryCounts(ctx, filter)
	if err != nil {
		return model.AttendanceSummaryResponse{}, err
	}

	var comparisonCounts map[model.AttendanceStatus]int64
	if filter.DateFrom != nil && filter.DateTo != nil {
		duration := filter.DateTo.Sub(*filter.DateFrom)
		compFilter := filter
		compDateFrom := filter.DateFrom.Add(-duration)
		compDateTo := *filter.DateFrom
		compFilter.DateFrom = &compDateFrom
		compFilter.DateTo = &compDateTo

		comparisonCounts, _ = s.repo.GetSummaryCounts(ctx, compFilter)
	}

	var todayTotal int64
	for _, count := range currentCounts {
		todayTotal += count
	}

	var compTotal int64
	for _, count := range comparisonCounts {
		compTotal += count
	}

	todayOntime := currentCounts[model.StatusWorking] + currentCounts[model.StatusDone]
	todayLate := currentCounts[model.StatusLate]

	compOntime := comparisonCounts[model.StatusWorking] + comparisonCounts[model.StatusDone]
	compLate := comparisonCounts[model.StatusLate]

	return model.AttendanceSummaryResponse{
		TotalRecord:     todayTotal,
		TotalRecordDiff: calculateDiff(todayTotal, compTotal),

		OntimeSummary:     todayOntime,
		OntimeSummaryDiff: calculateDiff(todayOntime, compOntime),

		LateSummary:     todayLate,
		LateSummaryDiff: calculateDiff(todayLate, compLate),
	}, nil
}

// GetTodayAttendance mengambil data absensi hari ini khusus untuk satu user.
// Fungsi ini sangat dioptimalkan dengan Redis caching (berlaku 24 jam) untuk mengurangi beban DB.
//
// Parameters:
//   - ctx: context.Context.
//   - userID: uint, user target.
//   - forceSync: bool, jika true akan membypass dan menghapus cache, langsung tembak ke DB.
//
// Returns/Impact:
//   - []model.AttendanceResponse: List absensi hari ini (bisa multiple jika tenant mengizinkan).
//   - error: jika DB atau Redis bermasalah.
func (s *attendanceService) GetTodayAttendance(ctx context.Context, userID uint, forceSync bool) ([]model.AttendanceResponse, error) {
	now := utils.Now()
	dateStr := now.Format("2006-01-02")
	cacheKey := fmt.Sprintf("cache:attendance:today:%d:%s", userID, dateStr)

	// 1. Try get from cache
	if !forceSync {
		cachedData, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil && cachedData != "" {
			var res []model.AttendanceResponse
			if err := json.Unmarshal([]byte(cachedData), &res); err == nil {
				return res, nil
			}
		}
	}

	// 2. Get from DB
	attendances, err := s.repo.FindAllTodayByUser(ctx, userID, now)
	if err != nil {
		return nil, err
	}
	if len(attendances) == 0 {
		if forceSync {
			s.redis.Del(ctx, cacheKey)
		}
		return nil, nil
	}

	var res []model.AttendanceResponse
	for _, a := range attendances {
		// FIX DATA ANOMALY: Jika parent status Done, tapi log terakhir adalah clock_in,
		// kita sisipkan dummy clock_out log agar handler merender session dengan benar.
		if a.Status == model.StatusDone && len(a.Logs) > 0 {
			lastLog := a.Logs[len(a.Logs)-1]
			if lastLog.Action == string(model.ClockIn) && a.ClockOutTime != nil {
				dummyLog := model.AttendanceLog{
					AttendanceID: a.ID,
					Action:       string(model.ClockOut),
					LogTime:      *a.ClockOutTime,
				}
				a.Logs = append(a.Logs, dummyLog)
			}
		}

		res = append(res, applyPreloads(&a, []string{}))
	}

	// 3. Save to cache (TTL 24h)
	jsonData, _ := json.Marshal(res)
	s.redis.Set(ctx, cacheKey, string(jsonData), 24*time.Hour)

	return res, nil
}
