package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"

	"github.com/redis/go-redis/v9"
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
		requesterID uint,
		filter model.AttendanceFilter,
		includes []string,
		limit, offset int,
	) ([]model.AttendanceResponse, int64, error)

	GetSummary(ctx context.Context, tenantID uint, filter model.AttendanceFilter) (model.AttendanceSummaryResponse, error)

	GetTodayAttendance(ctx context.Context, userID uint) (*model.AttendanceResponse, error)
}

type attendanceService struct {
	repo         repository.AttendanceRepository
	userRepo     repository.UserRepository
	settingRepo  repository.TenantSettingRepository
	tenantRepo   repository.TenantRepository
	activityRepo repository.RecentActivityRepository
	hrOpsRepo    repository.HrOpsRepository
	leaveRepo    repository.LeaveRepository
	userService  UserService
	redis        *redis.Client
	// Queue for background processing
	recordQueue chan attendanceTask
}

type attendanceTask struct {
	ctx      context.Context
	userID   uint
	tenantID uint
	data     *model.Attendance
	isUpdate bool
}

func NewAttendanceService(
	repo repository.AttendanceRepository,
	userRepo repository.UserRepository,
	settingRepo repository.TenantSettingRepository,
	tenantRepo repository.TenantRepository,
	activityRepo repository.RecentActivityRepository,
	hrOpsRepo repository.HrOpsRepository,
	leaveRepo repository.LeaveRepository,
	userService UserService,
	redis *redis.Client,
) AttendanceService {
	s := &attendanceService{
		repo:         repo,
		userRepo:     userRepo,
		settingRepo:  settingRepo,
		tenantRepo:   tenantRepo,
		activityRepo: activityRepo,
		hrOpsRepo:    hrOpsRepo,
		leaveRepo:    leaveRepo,
		userService:  userService,
		redis:        redis,
		recordQueue:  make(chan attendanceTask, 1000), // Buffer for 1000 requests
	}

	// Start background workers (e.g., 5 workers)
	for i := 0; i < 5; i++ {
		go s.attendanceWorker()
	}

	return s
}

func (s *attendanceService) attendanceWorker() {
	for task := range s.recordQueue {
		// Use a fresh context or background context for DB operations
		bgCtx := context.Background()
		if task.isUpdate {
			_ = s.repo.Update(bgCtx, task.data)
		} else {
			_ = s.repo.Save(bgCtx, task.data)
		}

		// Record activity in background
		title := "Attendance Clock In"
		action := "ClockIn"
		if task.isUpdate {
			title = "Attendance Clock Out"
			action = "ClockOut"
		}

		_ = s.activityRepo.Create(bgCtx, &model.RecentActivity{
			UserID: task.userID,
			Title:  title,
			Action: action,
			Status: "success",
		})
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

	// 1. Redis Distributed Lock
	lockKey := fmt.Sprintf("lock:attendance:%d", userID)
	// Lock for 5 seconds to prevent double submit
	locked, err := s.redis.SetNX(ctx, lockKey, "locked", 5*time.Second).Result()
	if err != nil {
		return model.AttendanceResponse{}, fmt.Errorf("redis error: %v", err)
	}
	if !locked {
		return model.AttendanceResponse{}, errors.New("terlalu banyak permintaan, harap tunggu sebentar")
	}
	// Release lock after process finishes
	defer s.redis.Del(ctx, lockKey)

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

	// 1. Check Holiday
	holiday, _ := s.hrOpsRepo.FindHolidayByDate(ctx, user.TenantID, now)
	if holiday != nil {
		return model.AttendanceResponse{}, fmt.Errorf("hari ini adalah hari libur: %s", holiday.Name)
	}

	// 2. Check Approved Leave
	isOnLeave, _ := s.leaveRepo.CheckOnLeave(ctx, userID, now)
	if isOnLeave {
		return model.AttendanceResponse{}, errors.New("anda sedang cuti tidak bisa absen")
	}

	// 3. Check Roster/Shift
	rosters, _ := s.hrOpsRepo.FindRoster(ctx, user.TenantID, userID, now, now)
	var activeShift *model.WorkShift
	if len(rosters) > 0 && rosters[0].ShiftID != nil {
		activeShift = rosters[0].Shift
	}

	// 3. Determine Time Constraints
	clockInStart := setting.ClockInStartTime
	clockInEnd := setting.ClockInEndTime
	clockOutStart := setting.ClockOutStartTime
	clockOutEnd := setting.ClockOutEndTime
	lateMinutes := setting.LateAfterMinute

	if activeShift != nil {
		clockInStart = activeShift.StartTime
		// For shifts, we might want to be more flexible with end times
		// or use shift times + buffer. Let's use shift start as reference for lateness.
		
		// Parse shift start to minutes for lateness
		sTime, _ := time.Parse("15:04", activeShift.StartTime)
		lateMinutes = sTime.Hour()*60 + sTime.Minute()
		
		// Adjust clock in/out ranges based on shift (simple logic for now)
		// Assuming clock out is allowed after shift ends
		clockOutStart = activeShift.EndTime
	}

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
		ok, err := isWithinTimeRange(now, clockInStart, clockInEnd)
		if err != nil || !ok {
			return model.AttendanceResponse{}, fmt.Errorf(
				"clock-in gagal, anda melakukan pada %s, batas clock-in %s - %s",
				nowStr,
				normalizeTime(clockInStart),
				normalizeTime(clockInEnd),
			)
		}

		status := model.StatusWorking
		// now.Hour() dan now.Minute() sekarang aman dari bias UTC 0 server
		if now.Hour()*60+now.Minute() > lateMinutes {
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

		// 2. Queue for background processing (Peak hour handling)
		s.recordQueue <- attendanceTask{
			ctx:      ctx,
			userID:   userID,
			tenantID: user.TenantID,
			data:     &data,
			isUpdate: false,
		}

		// Invalidate Cache immediately
		cacheKey := fmt.Sprintf("cache:attendance:today:%d:%s", userID, now.Format("2006-01-02"))
		s.redis.Del(ctx, cacheKey)

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

		ok, err := isWithinTimeRange(now, clockOutStart, clockOutEnd)
		if err != nil || !ok {
			return model.AttendanceResponse{}, fmt.Errorf(
				"clock-out gagal, anda melakukan pada %s, batas clock-out %s - %s",
				nowStr,
				clockOutStart,
				clockOutEnd,
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

		// 2. Queue for background processing (Peak hour handling)
		s.recordQueue <- attendanceTask{
			ctx:      ctx,
			userID:   userID,
			tenantID: user.TenantID,
			data:     todayAttendance,
			isUpdate: true,
		}

		// Invalidate Cache immediately
		cacheKey := fmt.Sprintf("cache:attendance:today:%d:%s", userID, now.Format("2006-01-02"))
		s.redis.Del(ctx, cacheKey)

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
	requesterID uint,
	filter model.AttendanceFilter,
	includes []string,
	limit, offset int,
) ([]model.AttendanceResponse, int64, error) {
	includes = filterAttendanceIncludes(includes)

	// Apply Hierarchical Scoping
	if requesterID != 0 {
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

// GetTodayAttendance godoc
// @Summary Get Today Attendance
// @Description Get today's attendance for the logged-in user
// @Tags Attendance
func (s *attendanceService) GetTodayAttendance(ctx context.Context, userID uint) (*model.AttendanceResponse, error) {
	now := time.Now().In(WIB)
	dateStr := now.Format("2006-01-02")
	cacheKey := fmt.Sprintf("cache:attendance:today:%d:%s", userID, dateStr)

	// 1. Try get from cache
	cachedData, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil && cachedData != "" {
		var res model.AttendanceResponse
		if err := json.Unmarshal([]byte(cachedData), &res); err == nil {
			return &res, nil
		}
	}

	// 2. Get from DB
	attendance, err := s.repo.FindTodayByUser(ctx, userID, now)
	if err != nil {
		return nil, err
	}
	if attendance == nil {
		return nil, nil
	}

	res := applyPreloads(attendance, []string{})

	// 3. Save to cache (TTL 24h)
	jsonData, _ := json.Marshal(res)
	s.redis.Set(ctx, cacheKey, string(jsonData), 24*time.Hour)

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
		CreatedAt:         a.ClockInTime, // Using ClockInTime as CreatedAt
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
