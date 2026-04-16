package service

import (
	"context"
	"errors"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
	"time"

	"github.com/google/uuid"
)

type AttendanceCorrectionService interface {
	RequestCorrection(ctx context.Context, userID uint, tenantID uint, req model.CreateCorrectionRequest) (model.AttendanceCorrectionResponse, error)
	GetCorrections(ctx context.Context, tenantID uint, userID uint, status string, limit, offset int) ([]model.AttendanceCorrectionResponse, int64, error)
	ApproveCorrection(ctx context.Context, requestID uint, adminID uint, notes string) error
	RejectCorrection(ctx context.Context, requestID uint, adminID uint, notes string) error
}

type attendanceCorrectionService struct {
	repo           repository.AttendanceCorrectionRepository
	attendanceRepo repository.AttendanceRepository
	userRepo       repository.UserRepository
	activityRepo   repository.RecentActivityRepository
}

func NewAttendanceCorrectionService(
	repo repository.AttendanceCorrectionRepository,
	attendanceRepo repository.AttendanceRepository,
	userRepo repository.UserRepository,
	activityRepo repository.RecentActivityRepository,
) AttendanceCorrectionService {
	return &attendanceCorrectionService{
		repo:           repo,
		attendanceRepo: attendanceRepo,
		userRepo:       userRepo,
		activityRepo:   activityRepo,
	}
}

func (s *attendanceCorrectionService) RequestCorrection(ctx context.Context, userID uint, tenantID uint, req model.CreateCorrectionRequest) (model.AttendanceCorrectionResponse, error) {
	parsedDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return model.AttendanceCorrectionResponse{}, errors.New("invalid date format, use YYYY-MM-DD")
	}

	if parsedDate.After(time.Now()) {
		return model.AttendanceCorrectionResponse{}, errors.New("cannot request correction for future dates")
	}

	var clockIn, clockOut *time.Time
	if req.ClockInTime != nil && *req.ClockInTime != "" {
		t, err := time.Parse("2006-01-02 15:04:05", req.Date+" "+*req.ClockInTime)
		if err != nil {
			return model.AttendanceCorrectionResponse{}, errors.New("invalid clock_in_time format, use HH:mm:ss")
		}
		clockIn = &t
	}

	if req.ClockOutTime != nil && *req.ClockOutTime != "" {
		t, err := time.Parse("2006-01-02 15:04:05", req.Date+" "+*req.ClockOutTime)
		if err != nil {
			return model.AttendanceCorrectionResponse{}, errors.New("invalid clock_out_time format, use HH:mm:ss")
		}
		clockOut = &t
	}

	correction := &model.AttendanceCorrection{
		UserID:       userID,
		TenantID:     tenantID,
		AttendanceID: req.AttendanceID,
		Date:         parsedDate,
		ClockInTime:  clockIn,
		ClockOutTime: clockOut,
		Reason:       req.Reason,
		Status:       model.CorrectionPending,
	}

	if err := s.repo.Create(ctx, correction); err != nil {
		return model.AttendanceCorrectionResponse{}, err
	}

	return mapToCorrectionResponse(correction), nil
}

func (s *attendanceCorrectionService) GetCorrections(ctx context.Context, tenantID uint, userID uint, status string, limit, offset int) ([]model.AttendanceCorrectionResponse, int64, error) {
	corrections, total, err := s.repo.FindAll(ctx, tenantID, userID, status, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	var responses []model.AttendanceCorrectionResponse
	for _, c := range corrections {
		responses = append(responses, mapToCorrectionResponse(&c))
	}
	return responses, total, nil
}

func (s *attendanceCorrectionService) ApproveCorrection(ctx context.Context, requestID uint, adminID uint, notes string) error {
	correction, err := s.repo.FindByID(ctx, requestID, []string{"user"})
	if err != nil {
		return errors.New("correction request not found")
	}

	if correction.Status != model.CorrectionPending {
		return errors.New("request is already processed")
	}

	now := time.Now()
	correction.Status = model.CorrectionApproved
	correction.ApprovedBy = &adminID
	correction.ApprovedAt = &now
	correction.AdminNotes = notes

	// Logic to update/create attendance
	var attendance *model.Attendance
	if correction.AttendanceID != nil {
		attendance, _ = s.attendanceRepo.FindByID(ctx, *correction.AttendanceID, nil)
	} else {
		// Try to find by user and date if ID not provided
		attendance, _ = s.attendanceRepo.FindTodayByUser(ctx, correction.UserID, correction.Date)
	}

	if attendance != nil {
		if correction.ClockInTime != nil {
			attendance.ClockInTime = *correction.ClockInTime
		}
		if correction.ClockOutTime != nil {
			attendance.ClockOutTime = correction.ClockOutTime
		}
		attendance.Status = model.StatusDone
		_ = s.attendanceRepo.Update(ctx, attendance)
	} else {
		// Create new attendance
		newAttendance := &model.Attendance{
			ID:          uuid.New(),
			UserID:      correction.UserID,
			TenantID:    correction.TenantID,
			ClockInTime: *correction.ClockInTime,
			Status:      model.StatusDone,
		}
		if correction.ClockOutTime != nil {
			newAttendance.ClockOutTime = correction.ClockOutTime
		}
		// Use defaults for location as it is a manual correction
		newAttendance.ClockInLatitude = 0
		newAttendance.ClockInLongitude = 0
		_ = s.attendanceRepo.Save(ctx, newAttendance)
	}

	// Log Activity
	s.activityRepo.Create(ctx, &model.RecentActivity{
		UserID: correction.UserID,
		Title:  "Attendance Corrected",
		Action: "CorrectionApproved",
		Status: "success",
	})

	return s.repo.Update(ctx, correction)
}

func (s *attendanceCorrectionService) RejectCorrection(ctx context.Context, requestID uint, adminID uint, notes string) error {
	correction, err := s.repo.FindByID(ctx, requestID, []string{})
	if err != nil {
		return errors.New("correction request not found")
	}

	if correction.Status != model.CorrectionPending {
		return errors.New("request is already processed")
	}

	now := time.Now()
	correction.Status = model.CorrectionRejected
	correction.ApprovedBy = &adminID
	correction.ApprovedAt = &now
	correction.AdminNotes = notes

	return s.repo.Update(ctx, correction)
}

func mapToCorrectionResponse(c *model.AttendanceCorrection) model.AttendanceCorrectionResponse {
	userName := ""
	if c.User != nil {
		userName = c.User.Name
	}

	var clockIn, clockOut *string
	if c.ClockInTime != nil {
		s := c.ClockInTime.Format("15:04:05")
		clockIn = &s
	}
	if c.ClockOutTime != nil {
		s := c.ClockOutTime.Format("15:04:05")
		clockOut = &s
	}

	return model.AttendanceCorrectionResponse{
		ID:           c.ID,
		UserID:       c.UserID,
		UserName:     userName,
		Date:         c.Date.Format("2006-01-02"),
		ClockInTime:  clockIn,
		ClockOutTime: clockOut,
		Reason:       c.Reason,
		Status:       c.Status,
		AdminNotes:   c.AdminNotes,
		CreatedAt:    c.CreatedAt,
	}
}
