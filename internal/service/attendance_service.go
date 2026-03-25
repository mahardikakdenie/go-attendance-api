package service

import (
	"errors"
	"time"

	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
)

type AttendanceService interface {
	CheckIn(req model.AttendanceRequest) (model.AttendanceResponse, error)
}

type attendanceService struct {
	repo repository.AttendanceRepository
}

func NewAttendanceService(repo repository.AttendanceRepository) AttendanceService {
	return &attendanceService{
		repo: repo,
	}
}

func (s *attendanceService) CheckIn(req model.AttendanceRequest) (model.AttendanceResponse, error) {
	if req.KaryawanID <= 0 {
		return model.AttendanceResponse{}, errors.New("ID Karyawan tidak valid")
	}

	now := time.Now()
	status := "Tepat Waktu"

	if now.Hour() >= 8 {
		status = "Telat"
	}

	attendanceData := model.Attendance{
		UserID:     uint(req.KaryawanID),
		WaktuMasuk: now,
		Status:     status,
	}

	err := s.repo.Save(&attendanceData)
	if err != nil {
		return model.AttendanceResponse{}, err
	}

	response := model.AttendanceResponse{
		ID:         int(attendanceData.ID),
		KaryawanID: int(attendanceData.UserID),
		WaktuMasuk: attendanceData.WaktuMasuk,
		Status:     attendanceData.Status,
	}

	return response, nil
}
