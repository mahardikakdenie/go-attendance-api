package service

import (
	"context"
	"errors"
	"fmt"
	modelDto "go-attendance-api/internal/dto"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
	"go-attendance-api/internal/utils"
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
	GetMyPaginatedReport(ctx context.Context, userID uint, tenantID uint, filter modelDto.TimesheetMonitoringFilter) (modelDto.PaginatedTimesheetReport, error)
	GetMyTimesheetRange(ctx context.Context, userID uint, start, end time.Time) (model.MonthlyTimesheetReport, error)

	// Monthly Report (HR)
	GetMonthlyReport(ctx context.Context, tenantID uint, userID uint, month, year int) (model.MonthlyTimesheetReport, error)
	GetMonthlyReportRange(ctx context.Context, tenantID uint, userID uint, start, end time.Time) (model.MonthlyTimesheetReport, error)

	// HR Monitoring & Analytics
	GetMonitoring(ctx context.Context, tenantID uint, filter modelDto.TimesheetMonitoringFilter) ([]modelDto.TimesheetMonitoringResponse, int64, error)
	GetAnalytics(ctx context.Context, tenantID uint, filter modelDto.TimesheetAnalyticsFilter) (modelDto.TimesheetAnalyticsResponse, error)
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
	loc, _ := time.LoadLocation("Asia/Jakarta")
	startDate, _ := time.ParseInLocation("2006-01-02", req.StartDate, loc)
	var endDate *time.Time
	if req.EndDate != "" {
		t, err := time.ParseInLocation("2006-01-02", req.EndDate, loc)
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
		t, err := utils.ParseDateWIB(req.StartDate)
		if err == nil {
			project.StartDate = &t
		}
	}

	if req.EndDate != "" {
		t, err := utils.ParseDateWIB(req.EndDate)
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
	if project.EndDate != nil && project.EndDate.Before(utils.Now()) && project.Status == model.ProjectStatusActive {
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

	// Parse date using WIB
	loc, _ := time.LoadLocation("Asia/Jakarta")

	// Konversi UTC dari frontend ke zona waktu WIB terlebih dahulu
	wibDate := req.Date.In(loc)

	// Ambil tanggal (Year, Month, Day) dari waktu WIB, set jam ke 00:00:00
	req.Date = time.Date(wibDate.Year(), wibDate.Month(), wibDate.Day(), 0, 0, 0, 0, loc)

	// FE mengirim "description", kita simpan ke kolom "notes" di DB
	if req.Description != "" && req.Notes == "" {
		req.Notes = req.Description
	}

	if req.TaskID == nil && req.TaskName != "" {		task, err := s.repo.FindTaskByName(ctx, req.ProjectID, req.TaskName)
		if err == nil && task != nil {
			req.TaskID = &task.ID
		} else {
			// Optionally create if not found
			newTask := &model.Task{
				ProjectID: req.ProjectID,
				UserID:    userID,
				Name:      req.TaskName,
			}
			if err := s.repo.CreateTask(ctx, newTask); err == nil {
				req.TaskID = &newTask.ID
			}
		}
	}

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

	for i := range entries {
		if entries[i].Task != nil {
			entries[i].TaskName = entries[i].Task.Name
		}
		entries[i].Description = entries[i].Notes
		totalHours += entries[i].DurationHours
		if entries[i].Project != nil {
			projectBreakdown[entries[i].Project.Name] += entries[i].DurationHours
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
		Period:           fmt.Sprintf("%d-%02d", year, month),
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

func (s *timesheetService) GetMyTimesheetRange(ctx context.Context, userID uint, start, end time.Time) (model.MonthlyTimesheetReport, error) {
	user, err := s.userRepo.FindByID(ctx, userID, []string{"manager"})
	if err != nil {
		return model.MonthlyTimesheetReport{}, err
	}

	entries, err := s.repo.FindEntriesByUserDateRange(ctx, userID, start, end)
	if err != nil {
		return model.MonthlyTimesheetReport{}, err
	}

	totalHours := 0.0
	projectBreakdown := make(map[string]float64)

	for i := range entries {
		if entries[i].Task != nil {
			entries[i].TaskName = entries[i].Task.Name
		}
		entries[i].Description = entries[i].Notes
		totalHours += entries[i].DurationHours
		if entries[i].Project != nil {
			projectBreakdown[entries[i].Project.Name] += entries[i].DurationHours
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
		Period:           fmt.Sprintf("%s to %s", start.Format("2006-01-02"), end.Format("2006-01-02")),
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
	return s.GetMyTimesheet(ctx, userID, month, year)
}

func (s *timesheetService) GetMonthlyReportRange(ctx context.Context, tenantID uint, userID uint, start, end time.Time) (model.MonthlyTimesheetReport, error) {
	return s.GetMyTimesheetRange(ctx, userID, start, end)
}

func (s *timesheetService) GetMonitoring(ctx context.Context, tenantID uint, filter modelDto.TimesheetMonitoringFilter) ([]modelDto.TimesheetMonitoringResponse, int64, error) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	startDate, _ := time.ParseInLocation("2006-01-02", filter.StartDate, loc)
	endDate, _ := time.ParseInLocation("2006-01-02", filter.EndDate, loc)

	repoFilter := model.TimesheetFilter{
		TenantID:  tenantID,
		UserID:    filter.UserID,
		ProjectID: filter.ProjectID,
		StartDate: &startDate,
		EndDate:   &endDate,
	}

	offset := (filter.Page - 1) * filter.Limit
	entries, total, err := s.repo.FindEntries(ctx, repoFilter, filter.Limit, offset)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]modelDto.TimesheetMonitoringResponse, 0)
	for _, entry := range entries {
		taskName := "-"
		if entry.Task != nil {
			taskName = entry.Task.Name
		}

		userName := "-"
		userRole := "-"
		if entry.User != nil {
			userName = entry.User.Name
			if entry.User.Role != nil {
				userRole = entry.User.Role.Name
			}
		}

		projectName := "-"
		if entry.Project != nil {
			projectName = entry.Project.Name
		}

		responses = append(responses, modelDto.TimesheetMonitoringResponse{
			ID: entry.ID.String(),
			User: modelDto.MappedUser{
				ID:   entry.UserID,
				Name: userName,
				Note: userRole,
			},
			Project: modelDto.ProjectItem{
				ID:   entry.ProjectID,
				Name: projectName,
			},
			TaskName:      taskName,
			Description:   entry.Notes,
			DurationHours: entry.DurationHours,
			Date:          entry.Date,
		})
	}

	return responses, total, nil
}

func (s *timesheetService) GetAnalytics(ctx context.Context, tenantID uint, filter modelDto.TimesheetAnalyticsFilter) (modelDto.TimesheetAnalyticsResponse, error) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	startDate, _ := time.ParseInLocation("2006-01-02", filter.StartDate, loc)
	endDate, _ := time.ParseInLocation("2006-01-02", filter.EndDate, loc)

	repoFilter := model.TimesheetFilter{
		TenantID:  tenantID,
		UserID:    filter.UserID,
		StartDate: &startDate,
		EndDate:   &endDate,
	}

	// 1. Get Summary (Total Hours & Active Employees)
	totalHours, activeEmployees, err := s.repo.GetAnalyticsSummary(ctx, repoFilter)
	if err != nil {
		return modelDto.TimesheetAnalyticsResponse{}, err
	}

	// 2. Get Project Distribution
	projectStats, err := s.repo.GetProjectDistribution(ctx, repoFilter)
	if err != nil {
		return modelDto.TimesheetAnalyticsResponse{}, err
	}

	projectDist := make([]modelDto.ProjectDistributionStats, 0)
	for _, ps := range projectStats {
		percentage := 0.0
		if totalHours > 0 {
			percentage = (ps.TotalHours / totalHours) * 100
		}
		projectDist = append(projectDist, modelDto.ProjectDistributionStats{
			ProjectID:   ps.ProjectID,
			ProjectName: ps.ProjectName,
			TotalHours:  ps.TotalHours,
			Percentage:  utils.RoundFloat(percentage, 1),
		})
	}

	// 3. Get Daily Stats
	dailyRaw, err := s.repo.GetDailyStats(ctx, repoFilter)
	if err != nil {
		return modelDto.TimesheetAnalyticsResponse{}, err
	}

	dailyHoursMap := make(map[string]float64)
	for _, dr := range dailyRaw {
		dailyHoursMap[dr.Date.Format("2006-01-02")] = dr.TotalHours
	}

	dailyStats := make([]modelDto.DailyTimesheetStats, 0)
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateKey := d.Format("2006-01-02")
		dailyStats = append(dailyStats, modelDto.DailyTimesheetStats{
			Date:       dateKey,
			TotalHours: dailyHoursMap[dateKey],
		})
	}

	return modelDto.TimesheetAnalyticsResponse{
		TotalHours:          utils.RoundFloat(totalHours, 1),
		ActiveEmployees:     activeEmployees,
		ProjectDistribution: projectDist,
		DailyStats:          dailyStats,
	}, nil
}

func (s *timesheetService) GetMyPaginatedReport(ctx context.Context, userID uint, tenantID uint, filter modelDto.TimesheetMonitoringFilter) (modelDto.PaginatedTimesheetReport, error) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	startDate, _ := time.ParseInLocation("2006-01-02", filter.StartDate, loc)
	endDate, _ := time.ParseInLocation("2006-01-02", filter.EndDate, loc)

	repoFilter := model.TimesheetFilter{
		TenantID:  tenantID,
		UserID:    userID,
		StartDate: &startDate,
		EndDate:   &endDate,
	}

	// Retrieve total hours
	totalHours, _, _ := s.repo.GetAnalyticsSummary(ctx, repoFilter)

	offset := (filter.Page - 1) * filter.Limit
	entries, total, err := s.repo.FindEntries(ctx, repoFilter, filter.Limit, offset)
	if err != nil {
		return modelDto.PaginatedTimesheetReport{}, err
	}

	var reportEntries []modelDto.TimesheetEntryDTO
	for _, entry := range entries {
		taskName := "-"
		if entry.Task != nil {
			taskName = entry.Task.Name
		} else if entry.TaskID != nil {
			task, err := s.repo.FindTaskByID(ctx, *entry.TaskID)
			if err == nil {
				taskName = task.Name
			}
		}

		projectName := "-"
		if entry.Project != nil {
			projectName = entry.Project.Name
		}

		reportEntries = append(reportEntries, modelDto.TimesheetEntryDTO{
			ID:            entry.ID.String(),
			ProjectName:   projectName,
			TaskName:      taskName,
			Description:   entry.Notes,
			DurationHours: entry.DurationHours,
			Date:          entry.Date,
			CreatedAt:     entry.CreatedAt,
		})
	}

	return modelDto.PaginatedTimesheetReport{
		Entries:    reportEntries,
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalHours: totalHours,
	}, nil
}
