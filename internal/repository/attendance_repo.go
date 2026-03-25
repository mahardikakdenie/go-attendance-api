package repository

import (
	// Sesuaikan dengan nama module di go.mod kamu
	"go-attendance-api/internal/model"

	"gorm.io/gorm"
)

// 1. Interface Repository
// Ini "kontrak" yang wajib dipenuhi oleh database (dibutuhkan oleh Service layer)
type AttendanceRepository interface {
	Save(attendance *model.Attendance) error
}

// 2. Struct private untuk menampung koneksi DB
type attendanceRepository struct {
	db *gorm.DB
}

// 3. Constructor
// Fungsi ini yang dipanggil di main.go: repository.NewAttendanceRepository(db)
func NewAttendanceRepository(db *gorm.DB) AttendanceRepository {
	return &attendanceRepository{
		db: db,
	}
}

// 4. Implementasi fungsi Save
// FIX: Gunakan *model.Attendance (entitas tabel database sesungguhnya), bukan Response
func (r *attendanceRepository) Save(attendance *model.Attendance) error {
	// GORM akan otomatis melakukan query INSERT INTO attendances ...
	return r.db.Create(attendance).Error
}
