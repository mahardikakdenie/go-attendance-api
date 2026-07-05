package attendance

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go-attendance-api/internal/model"
	"go-attendance-api/internal/utils"
)

// RecordAttendance memproses pencatatan absensi masuk (clock-in) atau pulang (clock-out).
// Fungsi ini memiliki mekanisme Redis lock untuk mencegah double submit, validasi jam shift, radius lokasi, dan selfie.
// 
// Parameters:
//   - ctx: context.Context, digunakan untuk timeout.
//   - userID: uint, ID dari user yang sedang melakukan absensi.
//   - req: model.AttendanceRequest, data payload (latitude, longitude, mediaUrl, dll).
// 
// Returns/Impact:
//   - model.AttendanceResponse: data absensi yang berhasil disimpan (disinkronisasi ke DB oleh worker).
//   - error: pesan kegagalan (misal: di luar radius, sedang cuti, dll) yang akan ditolak oleh sistem.
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

// EndSession secara eksplisit mengakhiri sesi absensi hari ini (mengubah status menjadi Done).
// Jika log terakhir adalah clock_in, sistem akan otomatis menambahkan log clock_out (dummy) ke database.
//
// Parameters:
//   - ctx: context.Context, untuk query database.
//   - userID: uint, ID user yang sesinya akan diakhiri.
//
// Returns/Impact:
//   - error: jika gagal update ke DB atau sesi tidak ditemukan. Jika nil, sesi berhasil diakhiri dan cache dihapus.
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

	// Cek apakah log terakhir adalah clock_in. Jika ya, tambahkan log clock_out.
	if len(todayAttendance.Logs) > 0 {
		lastLog := todayAttendance.Logs[len(todayAttendance.Logs)-1]
		if lastLog.Action == string(model.ClockIn) {
			attLog := model.AttendanceLog{
				AttendanceID: todayAttendance.ID,
				Action:       string(model.ClockOut),
				LogTime:      now,
			}
			if err := s.logRepo.Save(ctx, &attLog); err != nil {
				return fmt.Errorf("gagal menyimpan log absensi otomatis: %v", err)
			}
			todayAttendance.Logs = append(todayAttendance.Logs, attLog)
		}
	}

	// Update status to StatusDone
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
