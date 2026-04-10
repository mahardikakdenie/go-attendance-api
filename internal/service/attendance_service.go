package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
)

// Inisialisasi zona waktu UTC+7 (WIB) secara global untuk package service
var WIB = time.FixedZone("WIB", 7*3600)

type AttendanceService interface {
	RecordAttendance(
		ctx context.Context,
		userID uint,
		req model.AttendanceRequest,
	) (model.AttendanceResponse, error)

	GetAllData(
		ctx context.Context,
		filter model.AttendanceFilter,
		includes []string,
		limit, offset int,
	) ([]model.AttendanceResponse, int64, error)

	GetSummary(ctx context.Context, tenantID uint) (model.AttendanceSummaryResponse, error)

	GetTodayAttendance(ctx context.Context, userID uint) (*model.AttendanceResponse, error)
}

type attendanceService struct {
	repo         repository.AttendanceRepository
	userRepo     repository.UserRepository
	settingRepo  repository.TenantSettingRepository
	tenantRepo   repository.TenantRepository
	activityRepo repository.RecentActivityRepository
}

func NewAttendanceService(
	repo repository.AttendanceRepository,
	userRepo repository.UserRepository,
	settingRepo repository.TenantSettingRepository,
	tenantRepo repository.TenantRepository,
	activityRepo repository.RecentActivityRepository,
) AttendanceService {
	return &attendanceService{
		repo:         repo,
		userRepo:     userRepo,
		settingRepo:  settingRepo,
		tenantRepo:   tenantRepo,
		activityRepo: activityRepo,
	}
}

var allowedAttendanceIncludes = map[string]bool{
	"user":    true,
	"tenant":  true,
	"setting": true,
}

func filterAttendanceIncludes(includes []string) []string {
	var result []string
	for _, inc := range includes {
		if allowedAttendanceIncludes[inc] {
			result = append(result, inc)
		}
	}
	return result
}

func hasAttendanceInclude(includes []string, key string) bool {
	for _, inc := range includes {
		if inc == key {
			return true
		}
	}
	return false
}

func (s *attendanceService) RecordAttendance(
	ctx context.Context,
	userID uint,
	req model.AttendanceRequest,
) (model.AttendanceResponse, error) {

	if userID == 0 {
		return model.AttendanceResponse{}, errors.New("invalid user")
	}

	user, err := s.userRepo.FindByID(ctx, userID, []string{""})
	if err != nil {
		return model.AttendanceResponse{}, errors.New("user not found")
	}

	if user.TenantID == 0 {
		return model.AttendanceResponse{}, errors.New("user tenant is invalid")
	}

	tenant, err := s.tenantRepo.FindByID(ctx, user.TenantID)
	if err != nil || tenant == nil {
		return model.AttendanceResponse{}, errors.New("tenant not found")
	}

	setting, err := s.settingRepo.FindByTenantID(ctx, user.TenantID)
	if err != nil {
		return model.AttendanceResponse{}, errors.New("tenant setting not found")
	}

	// 🔥 FIX: Mengunci waktu saat ini ke UTC+7 (WIB)
	now := time.Now().In(WIB)
	nowStr := now.Format("15:04:05")

	todayAttendance, err := s.repo.FindTodayByUser(ctx, userID, now)
	if err != nil {
		return model.AttendanceResponse{}, err
	}

	if setting.RequireLocation && !setting.AllowRemote {
		distance := calculateDistance(
			setting.OfficeLatitude,
			setting.OfficeLongitude,
			req.Latitude,
			req.Longitude,
		)

		if distance > setting.MaxRadiusMeter {
			return model.AttendanceResponse{}, fmt.Errorf(
				"lokasi anda %.2fm dari kantor, maksimal %.2fm",
				distance,
				setting.MaxRadiusMeter,
			)
		}
	}

	if setting.RequireSelfie && req.MediaUrl == "" {
		return model.AttendanceResponse{}, errors.New("selfie is required")
	}

	switch req.Action {

	case model.ClockIn:

		if todayAttendance != nil && !setting.AllowMultipleCheck {
			return model.AttendanceResponse{}, errors.New("anda sudah clock-in hari ini")
		}

		// Karena `now` sudah di UTC+7, pengecekan jam dan menit di fungsi ini akan langsung match
		// dengan setting jam yang juga berasumsi waktu lokal (WIB).
		ok, err := isWithinTimeRange(now, setting.ClockInStartTime, setting.ClockInEndTime)
		if err != nil || !ok {
			return model.AttendanceResponse{}, fmt.Errorf(
				"clock-in gagal, anda melakukan pada %s, batas clock-in %s - %s",
				nowStr,
				normalizeTime(setting.ClockInStartTime),
				normalizeTime(setting.ClockInEndTime),
			)
		}

		status := model.StatusWorking
		// now.Hour() dan now.Minute() sekarang aman dari bias UTC 0 server
		if now.Hour()*60+now.Minute() > setting.LateAfterMinute {
			status = model.StatusLate
		}

		data := model.Attendance{
			UserID:           userID,
			TenantID:         user.TenantID,
			ClockInTime:      now, // Disimpan ke struct/DB dengan timezone UTC+7
			ClockInLatitude:  req.Latitude,
			ClockInLongitude: req.Longitude,
			ClockInMediaUrl:  req.MediaUrl,
			Status:           status,
		}

		if err := s.repo.Save(ctx, &data); err != nil {
			return model.AttendanceResponse{}, err
		}

		// Record activity
		_ = s.activityRepo.Create(ctx, &model.RecentActivity{
			UserID: userID,
			Title:  "Attendance Clock In",
			Action: "ClockIn",
			Status: "success",
		})

		// Preload user to ensure it's returned in the response
		userData, _ := s.userRepo.FindByID(ctx, userID, []string{})
		if userData != nil {
			data.User = *userData
		}

		return applyPreloads(&data, []string{"user"}), nil

	case model.ClockOut:

		if todayAttendance == nil {
			return model.AttendanceResponse{}, errors.New("anda belum melakukan clock-in hari ini")
		}

		if todayAttendance.ClockOutTime != nil && !setting.AllowMultipleCheck {
			return model.AttendanceResponse{}, errors.New("anda sudah clock-out hari ini")
		}

		ok, err := isWithinTimeRange(now, setting.ClockOutStartTime, setting.ClockOutEndTime)
		if err != nil || !ok {
			return model.AttendanceResponse{}, fmt.Errorf(
				"clock-out gagal, anda melakukan pada %s, batas clock-out %s - %s",
				nowStr,
				setting.ClockOutStartTime,
				setting.ClockOutEndTime,
			)
		}

		todayAttendance.ClockOutTime = &now
		todayAttendance.ClockOutLatitude = &req.Latitude
		todayAttendance.ClockOutLongitude = &req.Longitude

		if req.MediaUrl != "" {
			todayAttendance.ClockOutMediaUrl = &req.MediaUrl
		}

		// Maintain StatusLate if it was already late
		if todayAttendance.Status != model.StatusLate {
			todayAttendance.Status = model.StatusDone
		}

		if err := s.repo.Update(ctx, todayAttendance); err != nil {
			return model.AttendanceResponse{}, err
		}

		// Record activity
		_ = s.activityRepo.Create(ctx, &model.RecentActivity{
			UserID: userID,
			Title:  "Attendance Clock Out",
			Action: "ClockOut",
			Status: "success",
		})

		// Preload user to ensure it's returned in the response
		userData, _ := s.userRepo.FindByID(ctx, userID, []string{})
		if userData != nil {
			todayAttendance.User = *userData
		}

		return applyPreloads(todayAttendance, []string{"user"}), nil

	default:
		return model.AttendanceResponse{}, errors.New("invalid action")
	}
}

func (s *attendanceService) GetAllData(
	ctx context.Context,
	filter model.AttendanceFilter,
	includes []string,
	limit, offset int,
) ([]model.AttendanceResponse, int64, error) {
	includes = filterAttendanceIncludes(includes)

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

func (s *attendanceService) GetSummary(ctx context.Context, tenantID uint) (model.AttendanceSummaryResponse, error) {
	now := time.Now().In(WIB)
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, WIB)
	todayEnd := todayStart.Add(24 * time.Hour)

	yesterdayStart := todayStart.Add(-24 * time.Hour)
	yesterdayEnd := todayStart

	todayCounts, err := s.repo.GetSummaryCounts(ctx, tenantID, todayStart, todayEnd)
	if err != nil {
		return model.AttendanceSummaryResponse{}, err
	}

	yesterdayCounts, err := s.repo.GetSummaryCounts(ctx, tenantID, yesterdayStart, yesterdayEnd)
	if err != nil {
		return model.AttendanceSummaryResponse{}, err
	}

	// Check if yesterday has any data
	var yesterdayTotal int64
	for _, count := range yesterdayCounts {
		yesterdayTotal += count
	}

	comparisonCounts := yesterdayCounts
	if yesterdayTotal == 0 {
		// Try oldest data date
		oldestDate, _ := s.repo.GetOldestDataDate(ctx, tenantID)
		if oldestDate != nil {
			oldestStart := time.Date(oldestDate.Year(), oldestDate.Month(), oldestDate.Day(), 0, 0, 0, 0, oldestDate.Location())
			oldestEnd := oldestStart.Add(24 * time.Hour)
			if oldestStart.Before(todayStart) {
				comparisonCounts, _ = s.repo.GetSummaryCounts(ctx, tenantID, oldestStart, oldestEnd)
			}
		}
	}

	var compTotal int64
	for _, count := range comparisonCounts {
		compTotal += count
	}

	todayTotal := todayCounts[model.StatusWorking] + todayCounts[model.StatusDone] + todayCounts[model.StatusLate]
	todayOntime := todayCounts[model.StatusWorking] + todayCounts[model.StatusDone]
	todayLate := todayCounts[model.StatusLate]

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

// GetTodayAttendance godoc
// @Summary Get Today Attendance
// @Description Get today's attendance for the logged-in user
// @Tags Attendance
func (s *attendanceService) GetTodayAttendance(ctx context.Context, userID uint) (*model.AttendanceResponse, error) {
	now := time.Now().In(WIB)
	attendance, err := s.repo.FindTodayByUser(ctx, userID, now)
	if err != nil {
		return nil, err
	}
	if attendance == nil {
		return nil, nil
	}

	res := applyPreloads(attendance, []string{})
	return &res, nil
}

func calculateDiff(today, previous int64) float64 {
	if previous == 0 {
		if today > 0 {
			return 100.0
		}
		return 0.0
	}
	return float64(today-previous) / float64(previous) * 100.0
}

func isWithinTimeRange(now time.Time, start, end string) (bool, error) {
	layout := "15:04"

	if start == "24:00" {
		start = "00:00"
	}
	if end == "24:00" {
		end = "23:59"
	}

	startTime, err := time.Parse(layout, start)
	if err != nil {
		return false, err
	}

	endTime, err := time.Parse(layout, end)
	if err != nil {
		return false, err
	}

	// now.Hour() dan now.Minute() di sini akan mengekstrak jam lokal (UTC+7)
	// berkat `.In(WIB)` yang kita aplikasikan di atas.
	current := now.Hour()*60 + now.Minute()
	startMin := startTime.Hour()*60 + startTime.Minute()
	endMin := endTime.Hour()*60 + endTime.Minute()

	if startMin <= endMin {
		return current >= startMin && current <= endMin, nil
	}

	// Kasus jika shift malam (misal start 22:00, end 06:00)
	return current >= startMin || current <= endMin, nil
}

func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000

	dLat := (lat2 - lat1) * (math.Pi / 180)
	dLon := (lon2 - lon1) * (math.Pi / 180)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*(math.Pi/180))*
			math.Cos(lat2*(math.Pi/180))*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

func normalizeTime(t string) string {
	if t == "24:00" {
		return "23:59"
	}
	return t
}

func applyPreloads(a *model.Attendance, includes []string) model.AttendanceResponse {
	res := model.AttendanceResponse{
		ID:                a.ID,
		UserID:            a.UserID,
		ClockInTime:       a.ClockInTime,
		ClockOutTime:      a.ClockOutTime,
		ClockInLatitude:   a.ClockInLatitude,
		ClockInLongitude:  a.ClockInLongitude,
		ClockOutLatitude:  a.ClockOutLatitude,
		ClockOutLongitude: a.ClockOutLongitude,
		ClockInMediaUrl:   a.ClockInMediaUrl,
		ClockOutMediaUrl:  a.ClockOutMediaUrl,
		Status:            a.Status,
	}

	if hasAttendanceInclude(includes, "user") {
		var roleRes *model.RoleResponse
		if a.User.Role != nil {
			roleRes = &model.RoleResponse{
				ID:   a.User.Role.ID,
				Name: a.User.Role.Name,
			}
		}

		res.User = &model.UserResponse{
			ID:          a.User.ID,
			Name:        a.User.Name,
			Email:       a.User.Email,
			Role:        roleRes,
			TenantID:    a.User.TenantID,
			EmployeeID:  a.User.EmployeeID,
			Department:  a.User.Department,
			MediaUrl:    a.User.MediaUrl,
			Address:     a.User.Address,
			PhoneNumber: a.User.PhoneNumber,
			CreatedAt:   a.User.CreatedAt,
		}
	}

	return res
}
