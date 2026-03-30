package service

import (
	"errors"
	"time"

	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
)

type AttendanceService interface {
	RecordAttendance(req model.AttendanceRequest) (model.AttendanceResponse, error)
}

type attendanceService struct {
	repo repository.AttendanceRepository
}

func NewAttendanceService(repo repository.AttendanceRepository) AttendanceService {
	return &attendanceService{
		repo: repo,
	}
}

func (s *attendanceService) RecordAttendance(req model.AttendanceRequest) (model.AttendanceResponse, error) {
	if req.EmployeeID <= 0 {
		return model.AttendanceResponse{}, errors.New("invalid employee ID")
	}

	now := time.Now()

	if req.Action == "clock_in" {
		status := "On Time"
		if now.Hour() >= 8 {
			status = "Late"
		}

		data := model.Attendance{
			UserID:           uint(req.EmployeeID),
			ClockInTime:      now,
			ClockInLatitude:  req.Latitude,
			ClockInLongitude: req.Longitude,
			Status:           status,
			MediaUrl:         req.MediaUrl,
		}

		if err := s.repo.Save(&data); err != nil {
			return model.AttendanceResponse{}, err
		}

		return model.AttendanceResponse{
			ID:               data.ID,
			EmployeeID:       int(data.UserID),
			ClockInTime:      data.ClockInTime,
			ClockInLatitude:  data.ClockInLatitude,
			ClockInLongitude: data.ClockInLongitude,
			Status:           data.Status,
		}, nil
	}

	if req.Action == "clock_out" {
		data, err := s.repo.FindTodayByUser(uint(req.EmployeeID))
		if err != nil {
			return model.AttendanceResponse{}, errors.New("clock in record not found for today")
		}

		data.ClockOutTime = &now
		data.ClockOutLatitude = &req.Latitude
		data.ClockOutLongitude = &req.Longitude

		if err := s.repo.Update(&data); err != nil {
			return model.AttendanceResponse{}, err
		}

		return model.AttendanceResponse{
			ID:                data.ID,
			EmployeeID:        int(data.UserID),
			ClockInTime:       data.ClockInTime,
			ClockOutTime:      data.ClockOutTime,
			ClockInLatitude:   data.ClockInLatitude,
			ClockInLongitude:  data.ClockInLongitude,
			ClockOutLatitude:  data.ClockOutLatitude,
			ClockOutLongitude: data.ClockOutLongitude,
			Status:            data.Status,
		}, nil
	}

	return model.AttendanceResponse{}, errors.New("invalid action type")
}
