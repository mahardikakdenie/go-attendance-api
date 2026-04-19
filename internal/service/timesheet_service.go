package service

import (
	"context"
	"errors"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
	"time"
)

type TimesheetService interface {
	// Project (Admin/HR)
	CreateProject(ctx context.Context, tenantID uint, req model.ProjectRequest) (model.Project, error)
	GetProjects(ctx context.Context, tenantID uint, status string, search string) ([]model.Project, error)
	UpdateProject(ctx context.Context, id uint, tenantID uint, req model.ProjectRequest) (model.Project, error)
	DeleteProject(ctx context.Context, id uint, tenantID uint) error
	SuggestProjectStatus(project *model.Project) model.ProjectStatus

	// Project Members
	AddMembers(ctx context.Context, projectID uint, tenantID uint, members []model.ProjectMember) error
	RemoveMember(ctx context.Context, projectID uint, tenantID uint, userID uint) error
	GetMembers(ctx context.Context, projectID uint, tenantID uint) ([]model.ProjectMember, error)

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

func (s *timesheetService) CreateProject(ctx context.Context, tenantID uint, req model.ProjectRequest) (model.Project, error) {
	startDate, _ := time.Parse("2006-01-02", req.StartDate)
	var endDate *time.Time
	if req.EndDate != "" {
		t, err := time.Parse("2006-01-02", req.EndDate)
		if err == nil {
			endDate = &t
		}
	}

	project := &model.Project{
		TenantID:    tenantID,
		Name:        req.Name,
		Description: req.Description,
		ClientName:  req.ClientName,
		StartDate:   &startDate,
		EndDate:     endDate,
		Status:      req.Status,
		Budget:      req.Budget,
	}

	if project.Status == "" {
		project.Status = model.ProjectStatusActive
	}

	err := s.repo.CreateProject(ctx, project)
	return *project, err
}

func (s *timesheetService) GetProjects(ctx context.Context, tenantID uint, status string, search string) ([]model.Project, error) {
	return s.repo.FindProjects(ctx, tenantID, status, search)
}

func (s *timesheetService) UpdateProject(ctx context.Context, id uint, tenantID uint, req model.ProjectRequest) (model.Project, error) {
	project, err := s.repo.FindProjectByID(ctx, id, tenantID)
	if err != nil {
		return model.Project{}, errors.New("project not found")
	}

	if req.StartDate != "" {
		t, err := time.Parse("2006-01-02", req.StartDate)
		if err == nil {
			project.StartDate = &t
		}
	}

	if req.EndDate != "" {
		t, err := time.Parse("2006-01-02", req.EndDate)
		if err == nil {
			project.EndDate = &t
		}
	}

	project.Name = req.Name
	project.Description = req.Description
	project.ClientName = req.ClientName
	project.Status = req.Status
	project.Budget = req.Budget

	err = s.repo.UpdateProject(ctx, project)
	return *project, err
}

func (s *timesheetService) DeleteProject(ctx context.Context, id uint, tenantID uint) error {
	return s.repo.DeleteProject(ctx, id, tenantID)
}

func (s *timesheetService) SuggestProjectStatus(project *model.Project) model.ProjectStatus {
	if project.EndDate != nil && project.EndDate.Before(time.Now()) && project.Status == model.ProjectStatusActive {
		return model.ProjectStatusCompleted
	}
	return project.Status
}

func (s *timesheetService) AddMembers(ctx context.Context, projectID uint, tenantID uint, members []model.ProjectMember) error {
	// Verify project belongs to tenant
	_, err := s.repo.FindProjectByID(ctx, projectID, tenantID)
	if err != nil {
		return errors.New("project not found or access denied")
	}

	for i := range members {
		members[i].ProjectID = projectID
		if err := s.repo.AddProjectMember(ctx, &members[i]); err != nil {
			return err
		}
	}
	return nil
}

func (s *timesheetService) RemoveMember(ctx context.Context, projectID uint, tenantID uint, userID uint) error {
	_, err := s.repo.FindProjectByID(ctx, projectID, tenantID)
	if err != nil {
		return errors.New("project not found or access denied")
	}
	return s.repo.RemoveProjectMember(ctx, projectID, userID)
}

func (s *timesheetService) GetMembers(ctx context.Context, projectID uint, tenantID uint) ([]model.ProjectMember, error) {
	_, err := s.repo.FindProjectByID(ctx, projectID, tenantID)
	if err != nil {
		return nil, errors.New("project not found or access denied")
	}
	return s.repo.FindProjectMembers(ctx, projectID)
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
