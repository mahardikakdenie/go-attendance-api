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
	"go-attendance-api/internal/utils"

	"github.com/redis/go-redis/v9"
)

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

	GetTodayAttendance(ctx context.Context, userID uint, forceSync bool) ([]model.AttendanceResponse, error)
	EndSession(ctx context.Context, userID uint) error
}

type attendanceService struct {
	repo         repository.AttendanceRepository
	logRepo      repository.AttendanceLogRepository
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
	logRepo repository.AttendanceLogRepository,
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
		logRepo:      logRepo,
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

	user, err := s.userRepo.FindByID(ctx, userID, []string{"Setting"})
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
	now := utils.Now()
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
		sTime, _ := utils.ParseTimeWIB("15:04", activeShift.StartTime)
		lateMinutes = sTime.Hour()*60 + sTime.Minute()

		// Adjust clock in/out ranges based on shift (simple logic for now)
		// Assuming clock out is allowed after shift ends
		clockOutStart = activeShift.EndTime
	}

	todayAttendance, err := s.repo.FindTodayByUser(ctx, userID, now)
	if err != nil {
		return model.AttendanceResponse{}, err
	}

	allowMultiple := resolveAllowMultipleCheck(setting, user.Setting)

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

	// ── Dynamic Multi-Session Logic ───────────────────────────────────────────
	action := req.Action

	if todayAttendance == nil {
		// No existing record for today — this MUST be a clock_in action.
		if action != string(model.ClockIn) {
			return model.AttendanceResponse{}, errors.New("anda belum melakukan clock-in hari ini")
		}

		// Validate clock-in time window.
		ok, err := isWithinTimeRange(now, clockInStart, clockInEnd)
		if err != nil || !ok {
			return model.AttendanceResponse{}, fmt.Errorf(
				"clock-in gagal, anda melakukan pada %s, batas clock-in %s - %s",
				nowStr, clockInStart, clockInEnd,
			)
		}

		status := model.StatusWorking
		if now.Hour()*60+now.Minute() > lateMinutes {
			status = model.StatusLate
		}

		data := model.Attendance{
			UserID:           userID,
			TenantID:         user.TenantID,
			ClockInTime:      now,
			ClockInLatitude:  req.Latitude,
			ClockInLongitude: req.Longitude,
			ClockInMediaUrl:  req.MediaUrl,
			Status:           status,
		}

		// Save parent record synchronously so we get an ID for the log.
		if err := s.repo.Save(ctx, &data); err != nil {
			return model.AttendanceResponse{}, fmt.Errorf("gagal menyimpan absensi: %v", err)
		}

		// Create AttendanceLog for the clock_in event synchronously.
		attLog := model.AttendanceLog{
			AttendanceID: data.ID,
			Action:       action,
			LogTime:      now,
			Latitude:     req.Latitude,
			Longitude:    req.Longitude,
			MediaUrl:     req.MediaUrl,
		}
		if err := s.logRepo.Save(ctx, &attLog); err != nil {
			return model.AttendanceResponse{}, fmt.Errorf("gagal menyimpan log absensi: %v", err)
		}
		data.Logs = append(data.Logs, attLog)

		// Invalidate cache.
		cacheKey := fmt.Sprintf("cache:attendance:today:%d:%s", userID, now.Format("2006-01-02"))
		s.redis.Del(ctx, cacheKey)

		userData, _ := s.userRepo.FindByID(ctx, userID, []string{})
		if userData != nil {
			data.User = *userData
		}

		return applyPreloads(&data, []string{"user"}), nil
	}

	// ── Subsequent actions (todayAttendance already exists) ───────────────────

	// Standard mode: prevent duplicate clock_in or clock_out.
	if !allowMultiple {
		if action == string(model.ClockIn) {
			return model.AttendanceResponse{}, errors.New("anda sudah clock-in hari ini")
		}
		if action == string(model.ClockOut) && todayAttendance.ClockOutTime != nil {
			return model.AttendanceResponse{}, errors.New("anda sudah clock-out hari ini")
		}
	}

	// For clock_out action: validate time window and update parent Attendance.
	if action == string(model.ClockOut) {
		ok, err := isWithinTimeRange(now, clockOutStart, clockOutEnd)
		if err != nil || !ok {
			return model.AttendanceResponse{}, fmt.Errorf(
				"clock-out gagal, anda melakukan pada %s, batas clock-out %s - %s",
				nowStr, clockOutStart, clockOutEnd,
			)
		}

		todayAttendance.ClockOutTime = &now
		todayAttendance.ClockOutLatitude = &req.Latitude
		todayAttendance.ClockOutLongitude = &req.Longitude
		if req.MediaUrl != "" {
			todayAttendance.ClockOutMediaUrl = &req.MediaUrl
		}

		// Only auto-set status to done if allowMultiple is false.
		// If allowMultiple is true, it remains working/late until EndSession is called.
		if !allowMultiple {
			if todayAttendance.Status != model.StatusLate {
				todayAttendance.Status = model.StatusDone
			}
		}

		// Update parent record synchronously.
		if err := s.repo.Update(ctx, todayAttendance); err != nil {
			return model.AttendanceResponse{}, fmt.Errorf("gagal update absensi: %v", err)
		}
	}

	// Always create an AttendanceLog entry for any action synchronously.
	attLog := model.AttendanceLog{
		AttendanceID: todayAttendance.ID,
		Action:       action,
		LogTime:      now,
		Latitude:     req.Latitude,
		Longitude:    req.Longitude,
		MediaUrl:     req.MediaUrl,
	}
	if err := s.logRepo.Save(ctx, &attLog); err != nil {
		return model.AttendanceResponse{}, fmt.Errorf("gagal menyimpan log absensi: %v", err)
	}
	todayAttendance.Logs = append(todayAttendance.Logs, attLog)

	// Invalidate Cache immediately.
	cacheKey := fmt.Sprintf("cache:attendance:today:%d:%s", userID, now.Format("2006-01-02"))
	s.redis.Del(ctx, cacheKey)

	userData, _ := s.userRepo.FindByID(ctx, userID, []string{})
	if userData != nil {
		todayAttendance.User = *userData
	}

	return applyPreloads(todayAttendance, []string{"user"}), nil
}

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
		res = append(res, applyPreloads(&a, []string{}))
	}

	// 3. Save to cache (TTL 24h)
	jsonData, _ := json.Marshal(res)
	s.redis.Set(ctx, cacheKey, string(jsonData), 24*time.Hour)

	return res, nil
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

	startTime, err := utils.ParseTimeWIB(layout, start)
	if err != nil {
		return false, err
	}

	endTime, err := utils.ParseTimeWIB(layout, end)
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

	// Map AttendanceLog entries to AttendanceLogResponse.
	for _, l := range a.Logs {
		res.Logs = append(res.Logs, model.AttendanceLogResponse{
			ID:           l.ID,
			AttendanceID: l.AttendanceID,
			Action:       l.Action,
			LogTime:      l.LogTime,
			Latitude:     l.Latitude,
			Longitude:    l.Longitude,
			MediaUrl:     l.MediaUrl,
		})
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

func resolveAllowMultipleCheck(tenantSetting *model.TenantSetting, userSetting *model.UserSetting) bool {
	if userSetting != nil && userSetting.AllowMultipleCheck != nil {
		return *userSetting.AllowMultipleCheck
	}
	return tenantSetting.AllowMultipleCheck
}

func (s *attendanceService) EndSession(ctx context.Context, userID uint) error {
	now := utils.Now()
	todayAttendance, err := s.repo.FindTodayByUser(ctx, userID, now)
	if err != nil {
		return err
	}
	if todayAttendance == nil {
		return errors.New("tidak ada sesi absensi aktif hari ini")
	}

	// If status is already done, return early or error
	if todayAttendance.Status == model.StatusDone {
		return errors.New("sesi absensi hari ini sudah diakhiri")
	}

	// Update status to StatusDone (keep StatusLate if it was late? Wait, the plan says: "keep StatusLate if already late? Or change to StatusDone as complete." Let's check: "Mengubah status di tabel attendances dari working/late menjadi done (sebagai penanda shift/hari kerjanya benar-benar selesai)")
	// Actually, wait, model.StatusDone means session completed. If it's model.StatusLate, is it considered done?
	// Let's check what model.AttendanceStatus supports: StatusWorking, StatusDone, StatusLate.
	// So if they were late, and they end session, changing it to StatusDone is standard.
	todayAttendance.Status = model.StatusDone

	// If ClockOutTime is not set, set it to now
	if todayAttendance.ClockOutTime == nil {
		todayAttendance.ClockOutTime = &now
	}

	if err := s.repo.Update(ctx, todayAttendance); err != nil {
		return fmt.Errorf("gagal mengakhiri sesi absensi: %v", err)
	}

	// Invalidate Cache
	cacheKey := fmt.Sprintf("cache:attendance:today:%d:%s", userID, now.Format("2006-01-02"))
	s.redis.Del(ctx, cacheKey)

	return nil
}


