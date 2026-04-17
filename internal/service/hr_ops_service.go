package service

import (
	"context"
	dto "go-attendance-api/internal/dto"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type HrOpsService interface {
	// Shifts
	GetAllShifts(ctx context.Context, tenantID uint) ([]model.WorkShiftResponse, error)
	CreateShift(ctx context.Context, tenantID uint, req model.WorkShiftResponse) (model.WorkShiftResponse, error)

	// Roster
	GetWeeklyRoster(ctx context.Context, tenantID uint, startDateStr, endDateStr string, deptID *uint) ([]dto.EmployeeScheduleResponse, error)
	SaveRoster(ctx context.Context, tenantID uint, req dto.SaveRosterRequest) error

	// Calendar Events
	GetHolidays(ctx context.Context, tenantID uint, year int) ([]dto.CalendarEventResponse, error)
	CreateHoliday(ctx context.Context, tenantID uint, req dto.CreateCalendarEventRequest) (dto.CalendarEventResponse, error)
	UpdateHoliday(ctx context.Context, tenantID uint, id uuid.UUID, req dto.UpdateCalendarEventRequest) error
	DeleteHoliday(ctx context.Context, tenantID uint, id uuid.UUID) error

	// Lifecycle
	GetLifecycleTemplates(ctx context.Context, tenantID uint, category *model.LifecycleStatus) ([]dto.LifecycleTaskResponse, error)
	CreateLifecycleTemplate(ctx context.Context, tenantID uint, req dto.CreateLifecycleTemplateRequest) (dto.LifecycleTaskResponse, error)
	DeleteLifecycleTemplate(ctx context.Context, tenantID uint, id uuid.UUID) error
	GetEmployeeLifecycle(ctx context.Context, userID uint) (dto.EmployeeLifecycleResponse, error)
	UpdateLifecycleTask(ctx context.Context, userID uint, taskID uuid.UUID, isCompleted bool) error
}

type hrOpsService struct {
	repo        repository.HrOpsRepository
	userRepo    repository.UserRepository
	leaveRepo   repository.LeaveRepository
	settingRepo repository.TenantSettingRepository
}

func NewHrOpsService(repo repository.HrOpsRepository, userRepo repository.UserRepository, leaveRepo repository.LeaveRepository, settingRepo repository.TenantSettingRepository) HrOpsService {
	return &hrOpsService{repo: repo, userRepo: userRepo, leaveRepo: leaveRepo, settingRepo: settingRepo}
}

func (s *hrOpsService) GetAllShifts(ctx context.Context, tenantID uint) ([]model.WorkShiftResponse, error) {
	shifts, err := s.repo.FindAllShifts(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	res := make([]model.WorkShiftResponse, 0)
	for _, sh := range shifts {
		res = append(res, model.WorkShiftResponse{
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

func (s *hrOpsService) CreateShift(ctx context.Context, tenantID uint, req model.WorkShiftResponse) (model.WorkShiftResponse, error) {
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
		return model.WorkShiftResponse{}, err
	}

	req.ID = shift.ID
	return req, nil
}

func (s *hrOpsService) GetWeeklyRoster(ctx context.Context, tenantID uint, startDateStr, endDateStr string, deptID *uint) ([]dto.EmployeeScheduleResponse, error) {
	start, _ := time.Parse("2006-01-02", startDateStr)
	end, _ := time.Parse("2006-01-02", endDateStr)

	// 1. Fetch approved leaves for the range
	leaves, _, _ := s.leaveRepo.FindAll(ctx, model.LeaveFilter{
		TenantID: tenantID,
		DateFrom: &start,
		DateTo:   &end,
		Status:   model.LeaveStatusApproved,
	}, 0, 0)

	// Create leave lookup map [userID][YYYY-MM-DD]bool
	leaveMap := make(map[uint]map[string]bool)
	for _, l := range leaves {
		if _, ok := leaveMap[l.UserID]; !ok {
			leaveMap[l.UserID] = make(map[string]bool)
		}
		// Mark all dates within leave range
		curr := l.StartDate
		for !curr.After(l.EndDate) {
			leaveMap[l.UserID][curr.Format("2006-01-02")] = true
			curr = curr.AddDate(0, 0, 1)
		}
	}

	// 2. Fetch default shift for this tenant
	defaultShift, _ := s.repo.FindDefaultShift(ctx, tenantID)
	defaultVal := "work_shift_tenant"

	if defaultShift != nil {
		defaultVal = defaultShift.ID.String()
	} else {
		// If no default shift, try to get times from tenant settings
		if setting, err := s.settingRepo.FindByTenantID(ctx, tenantID); err == nil && setting != nil {
			defaultVal = fmt.Sprintf("work_shift_tenant (%s - %s)",
				setting.ClockInStartTime, setting.ClockOutStartTime)
		}
	}

	// 3. Fetch users
	users, _, err := s.userRepo.FindAll(ctx, model.UserFilter{TenantID: tenantID}, []string{"position"})
	if err != nil {
		return nil, err
	}

	// 4. Fetch rosters
	rosters, err := s.repo.FindRoster(ctx, tenantID, 0, start, end)
	if err != nil {
		return nil, err
	}

	// Create roster lookup map [userID][date_str]string
	rosterMap := make(map[uint]map[string]string)
	for _, r := range rosters {
		if _, ok := rosterMap[r.UserID]; !ok {
			rosterMap[r.UserID] = make(map[string]string)
		}
		dateStr := r.Date.Format("2006-01-02")
		if r.ShiftID != nil {
			rosterMap[r.UserID][dateStr] = r.ShiftID.String()
		} else {
			rosterMap[r.UserID][dateStr] = defaultVal
		}
	}

	// 5. Build Final Grid
	res := make([]dto.EmployeeScheduleResponse, 0)

	for _, u := range users {
		weekly := make(map[string]string)
		userRoster := rosterMap[u.ID]
		userLeaves := leaveMap[u.ID]

		// Iterate through each date in the range
		for currentDate := start; !currentDate.After(end); currentDate = currentDate.AddDate(0, 0, 1) {
			day := strings.ToLower(currentDate.Format("Monday"))
			dateStr := currentDate.Format("2006-01-02")

			// Priority 1: Check Leave
			if userLeaves[dateStr] {
				weekly[day] = "leave"
				continue
			}

			// Priority 2 & 3: Check Explicit Roster
			if val, exists := userRoster[dateStr]; exists {
				weekly[day] = val
				continue
			}

			// Priority 4 & 5: Fallback
			weekly[day] = defaultVal
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

	var allRosters []model.EmployeeRoster
	for _, assign := range req.Assignments {
		// Assuming we save a 7-day window based on common weekly view, 
		// but we should iterate based on what the front-end provides if possible.
		// For now, we iterate through the days present in the range from baseDate.
		for i := 0; i < 7; i++ {
			date := baseDate.AddDate(0, 0, i)
			dayName := strings.ToLower(date.Format("Monday"))
			
			shiftIDStr, exists := assign.Roster[dayName]
			if !exists {
				continue
			}

			var shiftID *uuid.UUID

			// If shiftIDStr is empty, "off", or starts with "work_shift_tenant", 
			// we set shiftID to nil to fallback to the default company shift.
			if shiftIDStr != "" && shiftIDStr != "off" && shiftIDStr != "leave" && !strings.HasPrefix(shiftIDStr, "work_shift_tenant") {
				if id, err := uuid.Parse(shiftIDStr); err == nil {
					shiftID = &id
				}
			}

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

func (s *hrOpsService) GetHolidays(ctx context.Context, tenantID uint, year int) ([]dto.CalendarEventResponse, error) {
	events, err := s.repo.FindEvents(ctx, tenantID, year)
	if err != nil {
		return nil, err
	}

	res := make([]dto.CalendarEventResponse, 0)
	for _, e := range events {
		res = append(res, dto.CalendarEventResponse{
			ID:          e.ID,
			Date:        e.Date.Format("2006-01-02"),
			Name:        e.Name,
			Type:        e.Type,
			Category:    e.Category,
			IsPaid:      e.IsPaid,
			Description: e.Description,
			IsAllUsers:  e.IsAllUsers,
			UserIDs:     s.mapUserIDs(e.Users),
		})
	}
	return res, nil
}

func (s *hrOpsService) mapUserIDs(users []model.User) []uint {
	ids := make([]uint, len(users))
	for i, u := range users {
		ids[i] = u.ID
	}
	return ids
}

func (s *hrOpsService) CreateHoliday(ctx context.Context, tenantID uint, req dto.CreateCalendarEventRequest) (dto.CalendarEventResponse, error) {
	date, _ := time.Parse("2006-01-02", req.Date)
	isPaid := req.IsPaid

	// Default category based on type if not provided
	category := req.Category
	if category == "" {
		if req.Type == model.EventTypeMeeting {
			category = model.EventCategoryInformation
		} else {
			category = model.EventCategoryOfficeClosed
		}
	}

	event := &model.CalendarEvent{
		TenantID:    tenantID,
		Date:        date,
		Name:        req.Name,
		Type:        req.Type,
		Category:    category,
		IsPaid:      isPaid,
		Description: req.Description,
		IsAllUsers:  req.IsAllUsers,
	}

	if !req.IsAllUsers && len(req.UserIDs) > 0 {
		for _, id := range req.UserIDs {
			event.Users = append(event.Users, model.User{ID: id})
		}
	}

	if err := s.repo.CreateEvent(ctx, event); err != nil {
		return dto.CalendarEventResponse{}, err
	}

	return dto.CalendarEventResponse{
		ID:          event.ID,
		Date:        req.Date,
		Name:        req.Name,
		Type:        req.Type,
		Category:    category,
		IsPaid:      isPaid,
		Description: req.Description,
		IsAllUsers:  req.IsAllUsers,
		UserIDs:     req.UserIDs,
	}, nil
}

func (s *hrOpsService) UpdateHoliday(ctx context.Context, tenantID uint, id uuid.UUID, req dto.UpdateCalendarEventRequest) error {
	event, err := s.repo.FindEventByID(ctx, tenantID, id)
	if err != nil {
		return err
	}

	if req.Name != "" {
		event.Name = req.Name
	}
	if req.Category != nil {
		event.Category = *req.Category
	}
	if req.IsPaid != nil {
		event.IsPaid = *req.IsPaid
	}
	if req.Description != "" {
		event.Description = req.Description
	}
	if req.IsAllUsers != nil {
		event.IsAllUsers = *req.IsAllUsers
	}

	if req.IsAllUsers != nil && !*req.IsAllUsers {
		event.Users = []model.User{}
		for _, uID := range req.UserIDs {
			event.Users = append(event.Users, model.User{ID: uID})
		}
	} else if req.IsAllUsers != nil && *req.IsAllUsers {
		event.Users = []model.User{}
	}

	return s.repo.UpdateEvent(ctx, event)
}

func (s *hrOpsService) DeleteHoliday(ctx context.Context, tenantID uint, id uuid.UUID) error {
	return s.repo.DeleteEvent(ctx, tenantID, id)
}

func (s *hrOpsService) GetLifecycleTemplates(ctx context.Context, tenantID uint, category *model.LifecycleStatus) ([]dto.LifecycleTaskResponse, error) {
	tasks, err := s.repo.FindLifecycleTasks(ctx, tenantID, category)
	if err != nil {
		return nil, err
	}

	res := make([]dto.LifecycleTaskResponse, 0)
	for _, t := range tasks {
		res = append(res, dto.LifecycleTaskResponse{
			ID:       t.ID,
			TaskName: t.TaskName,
			Category: t.Category,
		})
	}
	return res, nil
}

func (s *hrOpsService) CreateLifecycleTemplate(ctx context.Context, tenantID uint, req dto.CreateLifecycleTemplateRequest) (dto.LifecycleTaskResponse, error) {
	task := &model.LifecycleTask{
		TenantID: tenantID,
		TaskName: req.TaskName,
		Category: req.Category,
		IsSystem: false,
	}

	if err := s.repo.CreateLifecycleTask(ctx, task); err != nil {
		return dto.LifecycleTaskResponse{}, err
	}

	return dto.LifecycleTaskResponse{
		ID:       task.ID,
		TaskName: task.TaskName,
		Category: task.Category,
	}, nil
}

func (s *hrOpsService) DeleteLifecycleTemplate(ctx context.Context, tenantID uint, id uuid.UUID) error {
	return s.repo.DeleteLifecycleTask(ctx, tenantID, id)
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
