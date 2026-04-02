package service

import (
	"context"
	"errors"
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
	repo repository.AttendanceRepository
}

func NewAttendanceService(repo repository.AttendanceRepository) AttendanceService {
	return &attendanceService{
		repo: repo,
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

	now := time.Now()

	todayAttendance, err := s.repo.FindTodayByUser(ctx, userID)
	if err != nil {
		return model.AttendanceResponse{}, err
	}

	switch req.Action {

	case model.ClockIn:
		if todayAttendance != nil {
			return model.AttendanceResponse{}, errors.New("already clocked in today")
		}

		status := model.StatusWorking
		if now.Hour() >= 8 {
			status = model.StatusLate
		}

		data := model.Attendance{
			UserID:           userID,
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

func (s *attendanceService) GetAllData(
	ctx context.Context,
	filter model.AttendanceFilter,
	limit, offset int,
) ([]model.Attendance, int64, error) {

	data, total, err := s.repo.FindAll(ctx, filter, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return data, total, nil
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
