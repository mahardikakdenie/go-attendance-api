package service

import (
	"context"
	"errors"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
)

type TimesheetService interface {
	// Project (Admin/HR)
	CreateProject(ctx context.Context, tenantID uint, req model.Project) (model.Project, error)
	GetProjects(ctx context.Context, tenantID uint) ([]model.Project, error)
	UpdateProject(ctx context.Context, id uint, tenantID uint, req model.Project) (model.Project, error)

	// Task (Employee)
	CreateTask(ctx context.Context, userID uint, req model.Task) (model.Task, error)
	GetTasksByProject(ctx context.Context, projectID uint) ([]model.Task, error)

	// Timesheet Entry (Employee)
	CreateTimesheet(ctx context.Context, userID uint, tenantID uint, req model.TimesheetEntry) (model.TimesheetEntry, error)
	GetMyTimesheet(ctx context.Context, userID uint, month, year int) (model.MonthlyTimesheetReport, error)
	
	// Monthly Report (HR)
	GetMonthlyReport(ctx context.Context, tenantID uint, userID uint, month, year int) (model.MonthlyTimesheetReport, error)
}

type timesheetService struct {
	repo     repository.TimesheetRepository
	userRepo repository.UserRepository
}

func NewTimesheetService(repo repository.TimesheetRepository, userRepo repository.UserRepository) TimesheetService {
	return &timesheetService{
		repo:     repo,
		userRepo: userRepo,
	}
}

func (s *timesheetService) CreateProject(ctx context.Context, tenantID uint, req model.Project) (model.Project, error) {
	req.TenantID = tenantID
	err := s.repo.CreateProject(ctx, &req)
	return req, err
}

func (s *timesheetService) GetProjects(ctx context.Context, tenantID uint) ([]model.Project, error) {
	return s.repo.FindProjectsByTenant(ctx, tenantID)
}

func (s *timesheetService) UpdateProject(ctx context.Context, id uint, tenantID uint, req model.Project) (model.Project, error) {
	project, err := s.repo.FindProjectByID(ctx, id, tenantID)
	if err != nil {
		return model.Project{}, errors.New("project not found")
	}

	project.Name = req.Name
	project.Description = req.Description
	project.ClientName = req.ClientName
	project.Status = req.Status
	project.StartDate = req.StartDate
	project.EndDate = req.EndDate

	err = s.repo.UpdateProject(ctx, project)
	return *project, err
}

func (s *timesheetService) CreateTask(ctx context.Context, userID uint, req model.Task) (model.Task, error) {
	req.UserID = userID
	err := s.repo.CreateTask(ctx, &req)
	return req, err
}

func (s *timesheetService) GetTasksByProject(ctx context.Context, projectID uint) ([]model.Task, error) {
	return s.repo.FindTasksByProject(ctx, projectID)
}

func (s *timesheetService) CreateTimesheet(ctx context.Context, userID uint, tenantID uint, req model.TimesheetEntry) (model.TimesheetEntry, error) {
	req.UserID = userID
	req.TenantID = tenantID
	err := s.repo.CreateEntry(ctx, &req)
	return req, err
}

func (s *timesheetService) GetMyTimesheet(ctx context.Context, userID uint, month, year int) (model.MonthlyTimesheetReport, error) {
	user, err := s.userRepo.FindByID(ctx, userID, []string{"manager"})
	if err != nil {
		return model.MonthlyTimesheetReport{}, err
	}

	entries, err := s.repo.FindEntriesByUserPeriod(ctx, userID, month, year)
	if err != nil {
		return model.MonthlyTimesheetReport{}, err
	}

	totalHours := 0.0
	projectBreakdown := make(map[string]float64)

	for _, entry := range entries {
		totalHours += entry.DurationHours
		if entry.Project != nil {
			projectBreakdown[entry.Project.Name] += entry.DurationHours
		}
	}

	managerName := "-"
	if user.Manager != nil {
		managerName = user.Manager.Name
	}

	report := model.MonthlyTimesheetReport{
		EmployeeName:     user.Name,
		EmployeeID:       user.EmployeeID,
		Department:       user.Department,
		Period:           "Current Month", // Simplified for now
		Entries:          entries,
		TotalHours:       totalHours,
		ProjectBreakdown: projectBreakdown,
		Signatures: model.TimesheetSignatures{
			EmployeeName: user.Name,
			ManagerName:  managerName,
			HRName:       "(HR Representative)", // Placeholder
		},
	}

	return report, nil
}

func (s *timesheetService) GetMonthlyReport(ctx context.Context, tenantID uint, userID uint, month, year int) (model.MonthlyTimesheetReport, error) {
	// For HR/Admin viewing a specific employee's report
	return s.GetMyTimesheet(ctx, userID, month, year)
}
