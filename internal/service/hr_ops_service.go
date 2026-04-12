package service

import (
	"context"
	dto "go-attendance-api/internal/dto"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
	"strings"
	"time"

	"github.com/google/uuid"
)

type HrOpsService interface {
	// Shifts
	GetAllShifts(ctx context.Context, tenantID uint) ([]dto.WorkShiftResponse, error)
	CreateShift(ctx context.Context, tenantID uint, req dto.WorkShiftResponse) (dto.WorkShiftResponse, error)

	// Roster
	GetWeeklyRoster(ctx context.Context, tenantID uint, startDateStr, endDateStr string, deptID *uint) ([]dto.EmployeeScheduleResponse, error)
	SaveRoster(ctx context.Context, tenantID uint, req dto.SaveRosterRequest) error

	// Calendar
	GetHolidays(ctx context.Context, tenantID uint, year int) ([]dto.HolidayResponse, error)
	CreateHoliday(ctx context.Context, tenantID uint, req dto.CreateHolidayRequest) (dto.HolidayResponse, error)
	UpdateHoliday(ctx context.Context, tenantID uint, id uuid.UUID, req dto.UpdateHolidayRequest) error
	DeleteHoliday(ctx context.Context, tenantID uint, id uuid.UUID) error

	// Lifecycle
	GetEmployeeLifecycle(ctx context.Context, userID uint) (dto.EmployeeLifecycleResponse, error)
	UpdateLifecycleTask(ctx context.Context, userID uint, taskID uuid.UUID, isCompleted bool) error
}

type hrOpsService struct {
	repo     repository.HrOpsRepository
	userRepo repository.UserRepository
}

func NewHrOpsService(repo repository.HrOpsRepository, userRepo repository.UserRepository) HrOpsService {
	return &hrOpsService{repo: repo, userRepo: userRepo}
}

func (s *hrOpsService) GetAllShifts(ctx context.Context, tenantID uint) ([]dto.WorkShiftResponse, error) {
	shifts, err := s.repo.FindAllShifts(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	res := make([]dto.WorkShiftResponse, 0)
	for _, sh := range shifts {
		res = append(res, dto.WorkShiftResponse{
			ID:        sh.ID,
			Name:      sh.Name,
			StartTime: sh.StartTime,
			EndTime:   sh.EndTime,
			Type:      sh.Type,
			Color:     sh.Color,
			IsDefault: sh.IsDefault,
		})
	}
	return res, nil
}

func (s *hrOpsService) CreateShift(ctx context.Context, tenantID uint, req dto.WorkShiftResponse) (dto.WorkShiftResponse, error) {
	shift := &model.WorkShift{
		TenantID:  tenantID,
		Name:      req.Name,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Type:      req.Type,
		Color:     req.Color,
		IsDefault: req.IsDefault,
	}

	if err := s.repo.CreateShift(ctx, shift); err != nil {
		return dto.WorkShiftResponse{}, err
	}

	req.ID = shift.ID
	return req, nil
}

func (s *hrOpsService) GetWeeklyRoster(ctx context.Context, tenantID uint, startDateStr, endDateStr string, deptID *uint) ([]dto.EmployeeScheduleResponse, error) {
	start, _ := time.Parse("2006-01-02", startDateStr)
	end, _ := time.Parse("2006-01-02", endDateStr)

	users, _, err := s.userRepo.FindAll(ctx, model.UserFilter{TenantID: tenantID}, []string{"position"})
	if err != nil {
		return nil, err
	}

	rosters, err := s.repo.FindRoster(ctx, tenantID, 0, start, end)
	if err != nil {
		return nil, err
	}

	rosterMap := make(map[uint]map[string]string)
	for _, r := range rosters {
		if _, ok := rosterMap[r.UserID]; !ok {
			rosterMap[r.UserID] = make(map[string]string)
		}
		day := strings.ToLower(r.Date.Format("Monday"))
		if r.ShiftID != nil {
			rosterMap[r.UserID][day] = r.ShiftID.String()
		} else {
			rosterMap[r.UserID][day] = "off"
		}
	}

	res := make([]dto.EmployeeScheduleResponse, 0)
	for _, u := range users {
		weekly := rosterMap[u.ID]
		if weekly == nil {
			weekly = map[string]string{
				"monday": "off", "tuesday": "off", "wednesday": "off", "thursday": "off", "friday": "off", "saturday": "off", "sunday": "off",
			}
		}
		res = append(res, dto.EmployeeScheduleResponse{
			ID:           u.ID,
			Name:         u.Name,
			Avatar:       u.MediaUrl,
			Department:   u.Department,
			WeeklyRoster: weekly,
		})
	}

	return res, nil
}

func (s *hrOpsService) SaveRoster(ctx context.Context, tenantID uint, req dto.SaveRosterRequest) error {
	baseDate, _ := time.Parse("2006-01-02", req.StartDate)
	days := []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}

	var allRosters []model.EmployeeRoster
	for _, assign := range req.Assignments {
		for i, day := range days {
			shiftIDStr := assign.Roster[day]
			var shiftID *uuid.UUID
			if shiftIDStr != "off" && shiftIDStr != "" {
				id, _ := uuid.Parse(shiftIDStr)
				shiftID = &id
			}

			date := baseDate.AddDate(0, 0, i)
			allRosters = append(allRosters, model.EmployeeRoster{
				TenantID: tenantID,
				UserID:   assign.UserID,
				Date:     date,
				ShiftID:  shiftID,
			})
		}
	}

	return s.repo.SaveRoster(ctx, allRosters)
}

func (s *hrOpsService) GetHolidays(ctx context.Context, tenantID uint, year int) ([]dto.HolidayResponse, error) {
	holidays, err := s.repo.FindHolidays(ctx, tenantID, year)
	if err != nil {
		return nil, err
	}

	res := make([]dto.HolidayResponse, 0)
	for _, h := range holidays {
		res = append(res, dto.HolidayResponse{
			ID:     h.ID,
			Date:   h.Date.Format("2006-01-02"),
			Name:   h.Name,
			Type:   h.Type,
			IsPaid: h.IsPaid,
		})
	}
	return res, nil
}

func (s *hrOpsService) CreateHoliday(ctx context.Context, tenantID uint, req dto.CreateHolidayRequest) (dto.HolidayResponse, error) {
	date, _ := time.Parse("2006-01-02", req.Date)
	isPaid := req.IsPaid

	holiday := &model.Holiday{
		TenantID: tenantID,
		Date:     date,
		Name:     req.Name,
		Type:     req.Type,
		IsPaid:   isPaid,
	}

	if err := s.repo.CreateHoliday(ctx, holiday); err != nil {
		return dto.HolidayResponse{}, err
	}

	return dto.HolidayResponse{
		ID:     holiday.ID,
		Date:   req.Date,
		Name:   req.Name,
		Type:   req.Type,
		IsPaid: isPaid,
	}, nil
}

func (s *hrOpsService) UpdateHoliday(ctx context.Context, tenantID uint, id uuid.UUID, req dto.UpdateHolidayRequest) error {
	holiday, err := s.repo.FindHolidayByID(ctx, tenantID, id)
	if err != nil {
		return err
	}

	if req.Name != "" {
		holiday.Name = req.Name
	}
	if req.IsPaid != nil {
		holiday.IsPaid = *req.IsPaid
	}

	return s.repo.UpdateHoliday(ctx, holiday)
}

func (s *hrOpsService) DeleteHoliday(ctx context.Context, tenantID uint, id uuid.UUID) error {
	return s.repo.DeleteHoliday(ctx, tenantID, id)
}

func (s *hrOpsService) GetEmployeeLifecycle(ctx context.Context, userID uint) (dto.EmployeeLifecycleResponse, error) {
	tasks, err := s.repo.FindEmployeeLifecycle(ctx, userID)
	if err != nil {
		return dto.EmployeeLifecycleResponse{}, err
	}

	res := dto.EmployeeLifecycleResponse{
		EmployeeID: userID,
		Status:     model.LifecycleStatusOnboarding, // Default
		Tasks:      make([]dto.LifecycleTaskResponse, 0),
	}

	for _, t := range tasks {
		res.Tasks = append(res.Tasks, dto.LifecycleTaskResponse{
			ID:          t.TaskID,
			TaskName:    t.Task.TaskName,
			Category:    t.Task.Category,
			IsCompleted: t.IsCompleted,
			CompletedAt: t.CompletedAt,
		})
	}

	return res, nil
}

func (s *hrOpsService) UpdateLifecycleTask(ctx context.Context, userID uint, taskID uuid.UUID, isCompleted bool) error {
	return s.repo.UpdateEmployeeLifecycleTask(ctx, userID, taskID, isCompleted)
}
