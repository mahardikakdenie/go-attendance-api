package repository

import (
	"time"

	"go-attendance-api/internal/model"

	"gorm.io/gorm"
)

type AttendanceRepository interface {
	Save(attendance *model.Attendance) error
	Update(attendance *model.Attendance) error
	FindTodayByUser(userID uint) (model.Attendance, error)
}

type attendanceRepository struct {
	db *gorm.DB
}

func NewAttendanceRepository(db *gorm.DB) AttendanceRepository {
	return &attendanceRepository{
		db: db,
	}
}

func (r *attendanceRepository) Save(attendance *model.Attendance) error {
	return r.db.Create(attendance).Error
}

func (r *attendanceRepository) Update(attendance *model.Attendance) error {
	return r.db.Save(attendance).Error
}

func (r *attendanceRepository) FindTodayByUser(userID uint) (model.Attendance, error) {
	var attendance model.Attendance

	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	err := r.db.Where("user_id = ? AND clock_in_time >= ?", userID, startOfDay).First(&attendance).Error

	return attendance, err
}
