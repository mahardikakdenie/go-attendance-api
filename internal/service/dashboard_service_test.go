package service

import (
	"context"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repositories
type MockUserRepo struct{ mock.Mock; repository.UserRepository }
func (m *MockUserRepo) FindByID(ctx context.Context, id uint, includes []string) (*model.User, error) {
	args := m.Called(ctx, id, includes)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*model.User), args.Error(1)
}

type MockAttendanceRepo struct{ mock.Mock; repository.AttendanceRepository }
func (m *MockAttendanceRepo) FindAll(ctx context.Context, filter model.AttendanceFilter, includes []string, limit, offset int) ([]model.Attendance, int64, error) {
	args := m.Called(ctx, filter, includes, limit, offset)
	return args.Get(0).([]model.Attendance), args.Get(1).(int64), args.Error(2)
}

type MockLeaveRepo struct{ mock.Mock; repository.LeaveRepository }
func (m *MockLeaveRepo) FindAll(ctx context.Context, filter model.LeaveFilter, limit, offset int) ([]model.Leave, int64, error) {
	args := m.Called(ctx, filter, limit, offset)
	return args.Get(0).([]model.Leave), args.Get(1).(int64), args.Error(2)
}

type MockOvertimeRepo struct{ mock.Mock; repository.OvertimeRepository }
func (m *MockOvertimeRepo) FindAll(ctx context.Context, filter model.OvertimeFilter) ([]model.Overtime, int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]model.Overtime), args.Get(1).(int64), args.Error(2)
}

type MockTimesheetRepo struct{ mock.Mock; repository.TimesheetRepository }
func (m *MockTimesheetRepo) FindEntriesByUserPeriod(ctx context.Context, userID uint, month, year int) ([]model.TimesheetEntry, error) {
	args := m.Called(ctx, userID, month, year)
	return args.Get(0).([]model.TimesheetEntry), args.Error(1)
}

func TestGetEmployeeDNA(t *testing.T) {
	userRepo := new(MockUserRepo)
	attendanceRepo := new(MockAttendanceRepo)
	leaveRepo := new(MockLeaveRepo)
	overtimeRepo := new(MockOvertimeRepo)
	timesheetRepo := new(MockTimesheetRepo)

	s := &dashboardService{
		userRepo:       userRepo,
		attendanceRepo: attendanceRepo,
		leaveRepo:      leaveRepo,
		overtimeRepo:   overtimeRepo,
		timesheetRepo:  timesheetRepo,
	}

	ctx := context.Background()
	tenantID := uint(1)
	userID := uint(101)

	// Mock User
	userRepo.On("FindByID", ctx, userID, []string{"role", "position"}).Return(&model.User{
		ID: userID, TenantID: tenantID, Name: "Budi Santoso",
	}, nil)

	// Mock Attendance
	attendanceRepo.On("FindAll", ctx, mock.Anything, mock.Anything, 0, 0).Return([]model.Attendance{
		{Status: model.StatusWorking, ClockInTime: time.Now().In(WIB)},
	}, int64(1), nil)

	// Mock Overtime
	overtimeRepo.On("FindAll", ctx, mock.Anything).Return([]model.Overtime{}, int64(0), nil)

	// Mock Leave
	leaveRepo.On("FindAll", ctx, mock.Anything, 0, 0).Return([]model.Leave{}, int64(0), nil)

	// Mock Timesheet
	timesheetRepo.On("FindEntriesByUserPeriod", ctx, userID, mock.Anything, mock.Anything).Return([]model.TimesheetEntry{}, nil)

	res, err := s.GetEmployeeDNA(ctx, tenantID, userID)

	assert.NoError(t, err)
	assert.Equal(t, "Budi Santoso", res.User.(map[string]interface{})["name"])
	assert.NotNil(t, res.RadarMetrics)
	assert.NotNil(t, res.PunctualityDna)
}
