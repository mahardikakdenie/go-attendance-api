package repository

import (
	"go-attendance-api/internal/model"

	"gorm.io/gorm"
)

type AttendanceRepository interface {
	Save(attendance *model.Attendance) error
	FindTodayByUser(userID uint) (model.Attendance, error)
	Update(attendance *model.Attendance) error
	FindAll(filters ...model.AttendanceFilter) ([]model.Attendance, error)
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

func (r *attendanceRepository) FindTodayByUser(userID uint) (model.Attendance, error) {
	var attendance model.Attendance
	err := r.db.Where("user_id = ? AND DATE(clock_in_time) = CURRENT_DATE", userID).First(&attendance).Error
	return attendance, err
}

func (r *attendanceRepository) Update(attendance *model.Attendance) error {
	return r.db.Save(attendance).Error
}

func (r *attendanceRepository) FindAll(filters ...model.AttendanceFilter) ([]model.Attendance, error) {
	var attendances []model.Attendance
	query := r.db.Preload("User").Model(&model.Attendance{})

	if len(filters) > 0 {
		filter := filters[0]

		if filter.UserID != 0 {
			query = query.Where("user_id = ?", filter.UserID)
		}

		if filter.Status != "" {
			query = query.Where("status = ?", filter.Status)
		}
	}

	err := query.Find(&attendances).Error
	return attendances, err
}
