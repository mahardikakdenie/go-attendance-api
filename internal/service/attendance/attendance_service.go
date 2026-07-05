package attendance

import (
	"context"

	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
	"go-attendance-api/internal/service"

	"github.com/redis/go-redis/v9"
)

// AttendanceService mendefinisikan kontrak (interface) untuk layanan absensi.
// Menyediakan metode untuk merekam absensi, mengakhiri sesi, dan mengambil data absensi.
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

// attendanceService adalah implementasi konkret dari AttendanceService.
// Struct ini menyimpan referensi ke berbagai repository dan Redis client untuk memproses logika bisnis absensi.
type attendanceService struct {
	repo         repository.AttendanceRepository
	logRepo      repository.AttendanceLogRepository
	userRepo     repository.UserRepository
	settingRepo  repository.TenantSettingRepository
	tenantRepo   repository.TenantRepository
	activityRepo repository.RecentActivityRepository
	hrOpsRepo    repository.HrOpsRepository
	leaveRepo    repository.LeaveRepository
	userService  service.UserService
	redis        *redis.Client
	// Queue for background processing
	recordQueue chan attendanceTask
}

// attendanceTask merepresentasikan payload untuk diproses oleh background worker.
// 
// Fields:
//   - ctx: context background.
//   - userID, tenantID: identitas user dan tenant.
//   - data: pointer ke model Attendance yang akan disimpan/diupdate.
//   - isUpdate: flag boolean apakah ini proses update (clock-out) atau insert (clock-in).
type attendanceTask struct {
	ctx      context.Context
	userID   uint
	tenantID uint
	data     *model.Attendance
	isUpdate bool
}

// NewAttendanceService adalah constructor untuk membuat instance AttendanceService baru.
// Fungsi ini juga menginisialisasi channel antrian dan menjalankan goroutine background worker.
//
// Parameters:
//   - repo, logRepo, dsb: Semua dependensi repository dan Redis client.
// 
// Returns/Impact:
//   - AttendanceService: Mengembalikan interface yang siap digunakan oleh handler.
func NewAttendanceService(
	repo repository.AttendanceRepository,
	logRepo repository.AttendanceLogRepository,
	userRepo repository.UserRepository,
	settingRepo repository.TenantSettingRepository,
	tenantRepo repository.TenantRepository,
	activityRepo repository.RecentActivityRepository,
	hrOpsRepo repository.HrOpsRepository,
	leaveRepo repository.LeaveRepository,
	userService service.UserService,
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

// attendanceWorker berjalan di background sebagai goroutine untuk memproses penulisan ke database (save/update)
// secara asinkron agar response ke client lebih cepat, serta merekam aktivitas ke RecentActivity.
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
