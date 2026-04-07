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
}

type attendanceService struct {
	repo        repository.AttendanceRepository
	userRepo    repository.UserRepository
	settingRepo repository.TenantSettingRepository
	tenantRepo  repository.TenantRepository
}

func NewAttendanceService(
	repo repository.AttendanceRepository,
	userRepo repository.UserRepository,
	settingRepo repository.TenantSettingRepository,
	tenantRepo repository.TenantRepository,
) AttendanceService {
	return &attendanceService{
		repo:        repo,
		userRepo:    userRepo,
		settingRepo: settingRepo,
		tenantRepo:  tenantRepo,
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

	todayAttendance, err := s.repo.FindTodayByUser(ctx, userID)
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

		return applyPreloads(&data, []string{}), nil

	case model.ClockOut:

		if todayAttendance == nil {
			return model.AttendanceResponse{}, errors.New("anda belum melakukan clock-in hari ini")
		}

		if todayAttendance.ClockOutTime != nil {
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

		todayAttendance.Status = model.StatusDone

		if err := s.repo.Update(ctx, todayAttendance); err != nil {
			return model.AttendanceResponse{}, err
		}

		return applyPreloads(todayAttendance, []string{}), nil

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
		res.User = &model.UserResponse{
			ID:          a.User.ID,
			Name:        a.User.Name,
			Email:       a.User.Email,
			Role:        a.User.Role,
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
