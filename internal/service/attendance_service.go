package service

import (
	"context"
	"errors"
	"math"
	"time"

	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
)

type AttendanceService interface {
	RecordAttendance(
		ctx context.Context,
		userID uint,
		req model.AttendanceRequest,
	) (model.AttendanceResponse, error)

	GetAllData(ctx context.Context, filter model.AttendanceFilter, limit, offset int) ([]model.Attendance, int64, error)
}

type attendanceService struct {
	repo        repository.AttendanceRepository
	userRepo    repository.UserRepository
	settingRepo repository.TenantSettingRepository
	tenantRepo  repository.TenantRepository // 🔥 TAMBAHAN
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

func (s *attendanceService) RecordAttendance(
	ctx context.Context,
	userID uint,
	req model.AttendanceRequest,
) (model.AttendanceResponse, error) {

	if userID == 0 {
		return model.AttendanceResponse{}, errors.New("invalid user")
	}

	// ======================
	// GET USER
	// ======================
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return model.AttendanceResponse{}, errors.New("user not found")
	}

	if user.TenantID == 0 {
		return model.AttendanceResponse{}, errors.New("user tenant is invalid")
	}

	// ======================
	// 🔥 VALIDATE TENANT EXIST (FIX FK ERROR)
	// ======================
	tenant, err := s.tenantRepo.FindByID(ctx, user.TenantID)
	if err != nil || tenant == nil {
		return model.AttendanceResponse{}, errors.New("tenant not found")
	}

	// ======================
	// GET TENANT SETTING
	// ======================
	setting, err := s.settingRepo.FindByTenantID(ctx, user.TenantID)
	if err != nil {
		return model.AttendanceResponse{}, errors.New("tenant setting not found")
	}

	now := time.Now()

	// ======================
	// SAFE FIND TODAY
	// ======================
	todayAttendance, err := s.repo.FindTodayByUser(ctx, userID)
	if err != nil {
		return model.AttendanceResponse{}, err
	}

	// ======================
	// GEO VALIDATION
	// ======================
	if setting.RequireLocation && !setting.AllowRemote {
		distance := calculateDistance(
			setting.OfficeLatitude,
			setting.OfficeLongitude,
			req.Latitude,
			req.Longitude,
		)

		if distance > setting.MaxRadiusMeter {
			return model.AttendanceResponse{}, errors.New("outside allowed radius")
		}
	}

	// ======================
	// SELFIE VALIDATION
	// ======================
	if setting.RequireSelfie && req.MediaUrl == "" {
		return model.AttendanceResponse{}, errors.New("selfie is required")
	}

	switch req.Action {

	case model.ClockIn:

		if todayAttendance != nil && !setting.AllowMultipleCheck {
			return model.AttendanceResponse{}, errors.New("already clocked in today")
		}

		ok, err := isWithinTimeRange(now, setting.ClockInStartTime, setting.ClockInEndTime)
		if err != nil || !ok {
			return model.AttendanceResponse{}, errors.New("outside clock-in time window")
		}

		status := model.StatusWorking
		if now.Hour()*60+now.Minute() > setting.LateAfterMinute {
			status = model.StatusLate
		}

		data := model.Attendance{
			UserID:           userID,
			TenantID:         user.TenantID, // 🔥 sekarang aman
			ClockInTime:      now,
			ClockInLatitude:  req.Latitude,
			ClockInLongitude: req.Longitude,
			ClockInMediaUrl:  req.MediaUrl,
			Status:           status,
		}

		if err := s.repo.Save(ctx, &data); err != nil {
			return model.AttendanceResponse{}, err
		}

		return mapToResponse(&data), nil

	case model.ClockOut:

		if todayAttendance == nil {
			return model.AttendanceResponse{}, errors.New("you have not clocked in today")
		}

		if todayAttendance.ClockOutTime != nil {
			return model.AttendanceResponse{}, errors.New("already clocked out today")
		}

		ok, err := isWithinTimeRange(now, setting.ClockOutStartTime, setting.ClockOutEndTime)
		if err != nil || !ok {
			return model.AttendanceResponse{}, errors.New("outside clock-out time window")
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

		return mapToResponse(todayAttendance), nil

	default:
		return model.AttendanceResponse{}, errors.New("invalid action")
	}
}

// ======================
// GET ALL
// ======================
func (s *attendanceService) GetAllData(
	ctx context.Context,
	filter model.AttendanceFilter,
	limit, offset int,
) ([]model.Attendance, int64, error) {
	return s.repo.FindAll(ctx, filter, limit, offset)
}

// ======================
// HELPER
// ======================
func isWithinTimeRange(now time.Time, start, end string) (bool, error) {
	layout := "15:04"

	startTime, err := time.Parse(layout, start)
	if err != nil {
		return false, err
	}

	endTime, err := time.Parse(layout, end)
	if err != nil {
		return false, err
	}

	current := now.Hour()*60 + now.Minute()
	startMin := startTime.Hour()*60 + startTime.Minute()
	endMin := endTime.Hour()*60 + endTime.Minute()

	return current >= startMin && current <= endMin, nil
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

func mapToResponse(a *model.Attendance) model.AttendanceResponse {
	return model.AttendanceResponse{
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
}
