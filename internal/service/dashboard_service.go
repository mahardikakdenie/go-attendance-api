package service

import (
	"context"
	"encoding/json"
	"fmt"
	modelDto "go-attendance-api/internal/dto"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
	"math"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type DashboardService interface {
	GetAdminDashboard(ctx context.Context, currentUserID uint) (modelDto.AdminDashboardResponse, error)
	GetHrDashboard(ctx context.Context, tenantID uint, currentUserID uint) (modelDto.HrDashboardResponse, error)
	GetFinanceDashboard(ctx context.Context, tenantID uint, currentUserID uint) (modelDto.FinanceDashboardResponse, error)
	GetHeatmapData(ctx context.Context, tenantID uint, query modelDto.HeatmapQuery) ([]modelDto.HeatmapItem, error)
	GetDailyPulse(ctx context.Context, tenantID uint) (modelDto.DailyPulseResponse, error)
}

type dashboardService struct {
	tenantRepo     repository.TenantRepository
	userRepo       repository.UserRepository
	attendanceRepo repository.AttendanceRepository
	leaveRepo      repository.LeaveRepository
	overtimeRepo   repository.OvertimeRepository
	redis          *redis.Client
}

func NewDashboardService(
	tenantRepo repository.TenantRepository,
	userRepo repository.UserRepository,
	attendanceRepo repository.AttendanceRepository,
	leaveRepo repository.LeaveRepository,
	overtimeRepo repository.OvertimeRepository,
	redis *redis.Client,
) DashboardService {
	return &dashboardService{
		tenantRepo:     tenantRepo,
		userRepo:       userRepo,
		attendanceRepo: attendanceRepo,
		leaveRepo:      leaveRepo,
		overtimeRepo:   overtimeRepo,
		redis:          redis,
	}
}

// Helper: Standardized Avatar Fallback
func (s *dashboardService) getAvatar(name, avatar string) string {
	if avatar != "" {
		return avatar
	}
	return fmt.Sprintf("https://ui-avatars.com/api/?name=%s&background=random&color=fff", url.QueryEscape(name))
}

func (s *dashboardService) GetAdminDashboard(ctx context.Context, currentUserID uint) (modelDto.AdminDashboardResponse, error) {
	tenants, _ := s.tenantRepo.FindAll(ctx)
	totalUsers, _ := s.userRepo.CountByTenantID(ctx, 0)

	// Fetch current user
	currentUser, _ := s.userRepo.FindByID(ctx, currentUserID, []string{"role"})
	var userRes interface{}
	if currentUser != nil {
		currentUser.MediaUrl = s.getAvatar(currentUser.Name, currentUser.MediaUrl)
		userRes = mapToUserResponse(currentUser, []string{"role"})
	}

	// Calculate Monthly Growth & Tenant Growth Data
	now := time.Now()
	thisMonth := now.Month()
	thisYear := now.Year()
	lastMonth := now.AddDate(0, -1, 0).Month()
	lastMonthYear := now.AddDate(0, -1, 0).Year()

	thisMonthCount := int64(0)
	lastMonthCount := int64(0)

	growthData := make(map[string]int64)
	planDistributionMap := make(map[string]*modelDto.PlanDistributionItem)

	for _, t := range tenants {
		if t.CreatedAt.Month() == thisMonth && t.CreatedAt.Year() == thisYear {
			thisMonthCount++
		}
		if t.CreatedAt.Month() == lastMonth && t.CreatedAt.Year() == lastMonthYear {
			lastMonthCount++
		}

		monthKey := t.CreatedAt.Format("Jan")
		growthData[monthKey]++

		plan := t.Plan
		if plan == "" {
			plan = "Basic"
		}

		if _, ok := planDistributionMap[plan]; !ok {
			planDistributionMap[plan] = &modelDto.PlanDistributionItem{
				Label: plan,
				Value: 0,
			}
		}
		item := planDistributionMap[plan]
		item.Value++
		item.Users = append(item.Users, modelDto.MappedUser{
			ID:   t.ID,
			Name: t.Name,
		})
	}

	monthlyGrowth := 0.0
	if lastMonthCount > 0 {
		monthlyGrowth = float64(thisMonthCount-lastMonthCount) / float64(lastMonthCount) * 100
	} else if thisMonthCount > 0 {
		monthlyGrowth = 100.0
	}

	tenantGrowth := make([]modelDto.TenantGrowthItem, 0)
	for i := 5; i >= 0; i-- {
		m := now.AddDate(0, -i, 0).Format("Jan")
		tenantGrowth = append(tenantGrowth, modelDto.TenantGrowthItem{
			Month: m,
			Count: growthData[m],
		})
	}

	planDistribution := make([]modelDto.PlanDistributionItem, 0)
	for _, item := range planDistributionMap {
		planDistribution = append(planDistribution, *item)
	}

	return modelDto.AdminDashboardResponse{
		User: userRes,
		Stats: modelDto.AdminDashboardStats{
			TotalTenants:  int64(len(tenants)),
			TotalUsers:    totalUsers,
			ActiveSubs:    int64(len(tenants)),
			MonthlyGrowth: math.Round(monthlyGrowth*10) / 10,
		},
		TenantGrowth:     tenantGrowth,
		PlanDistribution: planDistribution,
	}, nil
}

func (s *dashboardService) GetHrDashboard(ctx context.Context, tenantID uint, currentUserID uint) (modelDto.HrDashboardResponse, error) {
	// 1. Try Cache First
	cacheKey := fmt.Sprintf("cache:dashboard:hr:%d", tenantID)
	cachedData, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil && cachedData != "" {
		var res modelDto.HrDashboardResponse
		if err := json.Unmarshal([]byte(cachedData), &res); err == nil {
			return res, nil
		}
	}

	now := time.Now().In(WIB)
	last30Days := now.AddDate(0, 0, -30)
	last6Months := now.AddDate(0, -6, 0)

	// 2. Fetch Data in Parallel
	var (
		users        []model.User
		attendances  []model.Attendance
		overtimes    []model.Overtime
		leaves       []model.Leave
		trendLeaves  []model.Leave
		pendingLeave int64
		wg           sync.WaitGroup
		mu           sync.Mutex
	)

	wg.Add(6)
	go func() {
		defer wg.Done()
		u, _, _ := s.userRepo.FindAll(ctx, model.UserFilter{TenantID: tenantID}, []string{"role", "position"})
		mu.Lock(); users = u; mu.Unlock()
	}()
	go func() {
		defer wg.Done()
		a, _, _ := s.attendanceRepo.FindAll(ctx, model.AttendanceFilter{TenantID: tenantID, DateFrom: &last30Days, DateTo: &now}, []string{}, 0, 0)
		mu.Lock(); attendances = a; mu.Unlock()
	}()
	go func() {
		defer wg.Done()
		o, _, _ := s.overtimeRepo.FindAll(ctx, model.OvertimeFilter{TenantID: tenantID, DateFrom: &last30Days, DateTo: &now, Status: model.OvertimeStatusApproved})
		mu.Lock(); overtimes = o; mu.Unlock()
	}()
	go func() {
		defer wg.Done()
		// Broaden: Include upcoming approved leaves too
		l, _, _ := s.leaveRepo.FindAll(ctx, model.LeaveFilter{TenantID: tenantID, DateFrom: &last30Days, Status: model.LeaveStatusApproved})
		mu.Lock(); leaves = l; mu.Unlock()
	}()
	go func() {
		defer wg.Done()
		p, _ := s.leaveRepo.GetPendingCount(ctx, tenantID)
		mu.Lock(); pendingLeave = p; mu.Unlock()
	}()
	go func() {
		defer wg.Done()
		// For trends, we need 6 months. Include future approved for the current month.
		l, _, _ := s.leaveRepo.FindAll(ctx, model.LeaveFilter{TenantID: tenantID, DateFrom: &last6Months, Status: model.LeaveStatusApproved})
		mu.Lock(); trendLeaves = l; mu.Unlock()
	}()
	wg.Wait()

	totalUsersCount := int64(len(users))
	if totalUsersCount == 0 {
		return modelDto.HrDashboardResponse{}, nil
	}

	currentUser, _ := s.userRepo.FindByID(ctx, currentUserID, []string{"role", "tenant"})
	var userRes interface{}
	if currentUser != nil {
		currentUser.MediaUrl = s.getAvatar(currentUser.Name, currentUser.MediaUrl)
		userRes = mapToUserResponse(currentUser, []string{"role", "tenant"})
	}

	userMap := make(map[uint]model.User)
	userPerformanceInfo := make(map[uint]struct {
		Score      int
		Department string
	})

	for i := range users {
		users[i].MediaUrl = s.getAvatar(users[i].Name, users[i].MediaUrl)
		userMap[users[i].ID] = users[i]
	}

	// 3. Process Attendance & Overtime
	userAttendances := make(map[uint][]model.Attendance)
	for _, a := range attendances {
		userAttendances[a.UserID] = append(userAttendances[a.UserID], a)
	}
	userOvertimes := make(map[uint][]model.Overtime)
	for _, o := range overtimes {
		userOvertimes[o.UserID] = append(userOvertimes[o.UserID], o)
	}

	performanceMatrix := make([]modelDto.EmployeePerformanceItem, 0)
	totalPresenceCount := 0
	totalOTHours := 0.0

	for _, u := range users {
		atts := userAttendances[u.ID]
		lateCount := 0
		totalClockInMinutes := 0
		userOTHours := 0.0
		for _, a := range atts {
			if a.Status == model.StatusLate {
				lateCount++
			}
			totalClockInMinutes += a.ClockInTime.Hour()*60 + a.ClockInTime.Minute()
			totalPresenceCount++
		}
		for _, o := range userOvertimes[u.ID] {
			start, _ := time.Parse("15:04", o.StartTime)
			end, _ := time.Parse("15:04", o.EndTime)
			diff := end.Sub(start).Hours()
			if diff < 0 {
				diff += 24
			}
			userOTHours += diff
		}
		totalOTHours += userOTHours

		score := 100 - (lateCount * 5)
		if score < 0 {
			score = 0
		}
		status := "Excellent"
		if score < 75 {
			status = "At Risk"
		} else if score < 90 {
			status = "Good"
		}

		userPerformanceInfo[u.ID] = struct {
			Score      int
			Department string
		}{Score: score, Department: u.Department}

		avgClockIn := "08:00 AM"
		if len(atts) > 0 {
			avgMin := totalClockInMinutes / len(atts)
			hour := (avgMin / 60)
			ampm := "AM"
			if hour >= 12 {
				ampm = "PM"
				if hour > 12 {
					hour -= 12
				}
			}
			if hour == 0 {
				hour = 12
			}
			avgClockIn = fmt.Sprintf("%02d:%02d %s", hour, avgMin%60, ampm)
		}

		performanceMatrix = append(performanceMatrix, modelDto.EmployeePerformanceItem{
			ID: u.ID, Name: u.Name, Avatar: u.MediaUrl, Department: u.Department,
			Score: score, TotalLate: lateCount, AvgClockIn: avgClockIn, Status: status, OvertimeHours: userOTHours,
		})
	}

	// 4. Leave Distribution & Trends (Dynamic)
	type userLeaveStats struct {
		Count int
		Days  int
	}
	leaveStatsMap := make(map[string]map[uint]*userLeaveStats)
	for _, l := range leaves {
		if l.LeaveType != nil {
			typeName := l.LeaveType.Name
			if _, ok := leaveStatsMap[typeName]; !ok {
				leaveStatsMap[typeName] = make(map[uint]*userLeaveStats)
			}
			if _, ok := leaveStatsMap[typeName][l.UserID]; !ok {
				leaveStatsMap[typeName][l.UserID] = &userLeaveStats{}
			}
			stats := leaveStatsMap[typeName][l.UserID]
			stats.Count++
			duration := int(l.EndDate.Sub(l.StartDate).Hours()/24) + 1
			if duration < 1 {
				duration = 1
			}
			stats.Days += duration
		}
	}

	leaveDistribution := make([]modelDto.PlanDistributionItem, 0)
	for typeName, userStats := range leaveStatsMap {
		item := modelDto.PlanDistributionItem{Label: typeName, Value: 0, Users: make([]modelDto.MappedUser, 0)}
		totalTypeRequests := 0
		for uID, stats := range userStats {
			totalTypeRequests += stats.Count
			if u, ok := userMap[uID]; ok && len(item.Users) < 10 {
				perf := userPerformanceInfo[uID]
				item.Users = append(item.Users, modelDto.MappedUser{
					ID:           u.ID,
					Name:         u.Name,
					Avatar:       u.MediaUrl,
					Department:   perf.Department,
					Score:        perf.Score,
					RequestCount: stats.Count,
					TotalDays:    stats.Days,
				})
			}
		}
		item.Value = int64(totalTypeRequests)
		leaveDistribution = append(leaveDistribution, item)
	}

	// Dynamic Leave Trends
	monthLabels := make([]string, 0)
	for i := 5; i >= 0; i-- {
		monthLabels = append(monthLabels, now.AddDate(0, -i, 0).Format("Jan"))
	}

	// map[TypeName]map[Month]TotalDays
	dynamicTrendsMap := make(map[string]map[string]int)
	typeNames := make([]string, 0)

	for _, l := range trendLeaves {
		if l.LeaveType == nil {
			continue
		}
		typeName := l.LeaveType.Name
		month := l.StartDate.Format("Jan")
		duration := int(l.EndDate.Sub(l.StartDate).Hours()/24) + 1
		if duration < 1 {
			duration = 1
		}

		if _, ok := dynamicTrendsMap[typeName]; !ok {
			dynamicTrendsMap[typeName] = make(map[string]int)
			typeNames = append(typeNames, typeName)
		}
		dynamicTrendsMap[typeName][month] += duration
	}

	leaveTrends := make([]modelDto.LeaveTrendSeries, 0)
	sort.Strings(typeNames) // Sort for consistency

	for _, name := range typeNames {
		series := modelDto.LeaveTrendSeries{
			Name: name,
			Data: make([]int, 6),
		}
		for i, m := range monthLabels {
			series.Data[i] = dynamicTrendsMap[name][m]
		}
		leaveTrends = append(leaveTrends, series)
	}

	// If no data, provide at least the month labels with empty data for common types
	if len(leaveTrends) == 0 {
		leaveTrends = []modelDto.LeaveTrendSeries{
			{Name: "Annual Leave", Data: make([]int, 6)},
			{Name: "Sick Leave", Data: make([]int, 6)},
		}
	}

	// 5. Build Top/Bottom Lists
	atRiskUsers := make([]modelDto.MappedUser, 0)
	for _, p := range performanceMatrix {
		if p.Score < 75 && len(atRiskUsers) < 10 {
			atRiskUsers = append(atRiskUsers, modelDto.MappedUser{
				ID: p.ID, Name: p.Name, Avatar: p.Avatar, Department: p.Department, Score: p.Score,
			})
		}
	}

	topPerformers := make([]modelDto.EmployeePerformanceItem, len(performanceMatrix))
	copy(topPerformers, performanceMatrix)
	sort.Slice(topPerformers, func(i, j int) bool { return topPerformers[i].Score > topPerformers[j].Score })
	if len(topPerformers) > 5 {
		topPerformers = topPerformers[:5]
	}

	needAttention := make([]modelDto.EmployeePerformanceItem, len(performanceMatrix))
	copy(needAttention, performanceMatrix)
	sort.Slice(needAttention, func(i, j int) bool { return needAttention[i].Score < needAttention[j].Score })
	if len(needAttention) > 5 {
		needAttention = needAttention[:5]
	}

	presenceRate := (float64(totalPresenceCount) / float64(totalUsersCount*22)) * 100
	if presenceRate > 100 {
		presenceRate = 100
	}

	finalRes := modelDto.HrDashboardResponse{
		User: userRes,
		Stats: modelDto.HrDashboardStats{
			PresenceRate: math.Round(presenceRate*10) / 10,
			AvgOvertime:  math.Round((totalOTHours/float64(totalUsersCount))*10) / 10,
			PendingLeave: pendingLeave,
			AtRiskStaff:  int64(len(atRiskUsers)),
			AtRiskUsers:  atRiskUsers,
		},
		TopPerformers:     topPerformers,
		NeedAttention:     needAttention,
		PerformanceMatrix: performanceMatrix,
		LeaveDistribution: leaveDistribution,
		LeaveTrends:       leaveTrends,
	}

	if jsonData, err := json.Marshal(finalRes); err == nil {
		s.redis.Set(ctx, cacheKey, string(jsonData), 5*time.Minute)
	}

	return finalRes, nil
}

func (s *dashboardService) GetHeatmapData(ctx context.Context, tenantID uint, query modelDto.HeatmapQuery) ([]modelDto.HeatmapItem, error) {
	if query.Type == "" {
		query.Type = "clockin"
	}

	cacheKey := fmt.Sprintf("cache:dashboard:heatmap:%d:%s:%d:%s:%s", tenantID, query.Type, query.UserID, query.DateFrom, query.DateTo)
	if cached, err := s.redis.Get(ctx, cacheKey).Result(); err == nil && cached != "" {
		var res []modelDto.HeatmapItem
		if err := json.Unmarshal([]byte(cached), &res); err == nil {
			return res, nil
		}
	}

	var dateFrom, dateTo *time.Time
	if query.DateFrom != "" {
		if t, err := time.Parse("2006-01-02", query.DateFrom); err == nil {
			dateFrom = &t
		}
	} else {
		t := time.Now().In(WIB).AddDate(0, 0, -30)
		dateFrom = &t
	}
	if query.DateTo != "" {
		if t, err := time.Parse("2006-01-02", query.DateTo); err == nil {
			dateTo = &t
		}
	} else {
		t := time.Now().In(WIB)
		dateTo = &t
	}

	heatmapUserMap := make(map[string][]modelDto.MappedUser)
	userMap := make(map[uint]model.User)
	var users []model.User

	users, _, _ = s.userRepo.FindAll(ctx, model.UserFilter{TenantID: tenantID}, nil)
	for i := range users {
		users[i].MediaUrl = s.getAvatar(users[i].Name, users[i].MediaUrl)
		userMap[users[i].ID] = users[i]
	}

	if query.Type == "leave" {
		leaves, _, _ := s.leaveRepo.FindAll(ctx, model.LeaveFilter{TenantID: tenantID, UserID: query.UserID, DateFrom: dateFrom, DateTo: dateTo, Status: model.LeaveStatusApproved})
		for _, l := range leaves {
			curr := l.StartDate
			note := ""
			if l.LeaveType != nil {
				note = l.LeaveType.Name
			}
			for !curr.After(l.EndDate) {
				day := curr.Format("Mon")
				key := day + "-09:00"
				s.addToHeatmapMap(heatmapUserMap, key, userMap[l.UserID], note)
				curr = curr.AddDate(0, 0, 1)
			}
		}
	} else {
		atts, _, _ := s.attendanceRepo.FindAll(ctx, model.AttendanceFilter{TenantID: tenantID, UserID: query.UserID, DateFrom: dateFrom, DateTo: dateTo}, nil, 0, 0)
		for _, a := range atts {
			var t *time.Time
			if query.Type == "clockin" {
				t = &a.ClockInTime
			} else {
				t = a.ClockOutTime
			}
			if t == nil {
				continue
			}

			day := t.Format("Mon")
			snappedMin := "00"
			if t.Minute() >= 30 {
				snappedMin = "30"
			}
			key := fmt.Sprintf("%s-%02d:%s", day, t.Hour(), snappedMin)
			s.addToHeatmapMap(heatmapUserMap, key, userMap[a.UserID], "")
		}
	}

	// 4. Build Dynamic Response (Only include slots with data)
	heatmap := make([]modelDto.HeatmapItem, 0)
	isFilteredByUser := query.UserID != 0
	totalUsersInTenant := int64(len(users))
	dayPriority := map[string]int{"Mon": 1, "Tue": 2, "Wed": 3, "Thu": 4, "Fri": 5, "Sat": 6, "Sun": 7}

	for key, muList := range heatmapUserMap {
		if len(muList) == 0 {
			continue
		}

		parts := strings.Split(key, "-")
		if len(parts) != 2 {
			continue
		}
		d := parts[0]
		t := parts[1]

		intensity := 0
		if isFilteredByUser {
			intensity = 100
		} else if totalUsersInTenant > 0 {
			intensity = int(math.Min(float64(len(muList)*100)/float64(totalUsersInTenant), 100))
		}

		heatmap = append(heatmap, modelDto.HeatmapItem{
			Day:   d,
			Time:  t,
			Value: intensity,
			Users: muList,
		})
	}

	// Sort heatmap chronologically
	sort.Slice(heatmap, func(i, j int) bool {
		if heatmap[i].Day != heatmap[j].Day {
			return dayPriority[heatmap[i].Day] < dayPriority[heatmap[j].Day]
		}
		return heatmap[i].Time < heatmap[j].Time
	})

	if jsonData, err := json.Marshal(heatmap); err == nil {
		s.redis.Set(ctx, cacheKey, string(jsonData), 5*time.Minute)
	}

	return heatmap, nil
}

func (s *dashboardService) addToHeatmapMap(m map[string][]modelDto.MappedUser, key string, u model.User, note string) {
	if u.ID == 0 {
		return
	}
	if len(m[key]) >= 10 {
		return
	}
	for _, mu := range m[key] {
		if mu.ID == u.ID {
			return
		}
	}
	m[key] = append(m[key], modelDto.MappedUser{
		ID:     u.ID,
		Name:   u.Name,
		Avatar: u.MediaUrl,
		Note:   note,
	})
}

func (s *dashboardService) GetFinanceDashboard(ctx context.Context, tenantID uint, currentUserID uint) (modelDto.FinanceDashboardResponse, error) {
	now := time.Now().In(WIB)
	last6Months := now.AddDate(0, -6, 0)

	// Fetch current user
	currentUser, _ := s.userRepo.FindByID(ctx, currentUserID, []string{"role"})
	var userRes interface{}
	if currentUser != nil {
		currentUser.MediaUrl = s.getAvatar(currentUser.Name, currentUser.MediaUrl)
		userRes = mapToUserResponse(currentUser, []string{"role"})
	}

	users, _, _ := s.userRepo.FindAll(ctx, model.UserFilter{TenantID: tenantID}, []string{})
	totalBaseSalary := 0.0
	userMap := make(map[uint]model.User)
	for i := range users {
		users[i].MediaUrl = s.getAvatar(users[i].Name, users[i].MediaUrl)
		totalBaseSalary += users[i].BaseSalary
		userMap[users[i].ID] = users[i]
	}

	overtimes, _, _ := s.overtimeRepo.FindAll(ctx, model.OvertimeFilter{
		TenantID: tenantID,
		DateFrom: &last6Months,
		Status:   model.OvertimeStatusApproved,
	})

	monthlyOTCosts := make(map[string]float64)
	totalOTCosts := 0.0
	for _, o := range overtimes {
		var userSalary float64
		for _, u := range users {
			if u.ID == o.UserID {
				userSalary = u.BaseSalary
				break
			}
		}

		hourlyRate := userSalary / 173.0
		start, _ := time.Parse("15:04", o.StartTime)
		end, _ := time.Parse("15:04", o.EndTime)
		diff := end.Sub(start).Hours()
		if diff < 0 {
			diff += 24
		}

		cost := diff * hourlyRate * 1.5
		totalOTCosts += cost

		monthKey := o.Date.Format("Jan")
		monthlyOTCosts[monthKey] += cost
	}

	trends := make([]modelDto.PayrollTrendItem, 0)
	for i := 5; i >= 0; i-- {
		m := now.AddDate(0, -i, 0).Format("Jan")
		trends = append(trends, modelDto.PayrollTrendItem{
			Month:         m,
			BaseSalary:    totalBaseSalary,
			OvertimeCosts: math.Round(monthlyOTCosts[m]*100) / 100,
		})
	}

	salariesPct := 70.0
	overtimePct := (totalOTCosts / (totalBaseSalary + totalOTCosts)) * 100
	if overtimePct > 30 {
		overtimePct = 30
	}

	overtimeUsers := make([]modelDto.MappedUser, 0)
	overtimeUserMap := make(map[uint]bool)
	for _, o := range overtimes {
		if !overtimeUserMap[o.UserID] {
			overtimeUserMap[o.UserID] = true
			if u, ok := userMap[o.UserID]; ok {
				overtimeUsers = append(overtimeUsers, modelDto.MappedUser{
					ID:     u.ID,
					Name:   u.Name,
					Avatar: u.MediaUrl,
				})
			}
		}
	}

	breakdown := []modelDto.PlanDistributionItem{
		{Label: "Salaries", Value: int64(salariesPct)},
		{Label: "Overtime", Value: int64(overtimePct), Users: overtimeUsers},
		{Label: "Taxes", Value: 10},
		{Label: "Benefits", Value: 10},
	}

	return modelDto.FinanceDashboardResponse{
		User: userRes,
		Stats: modelDto.FinanceDashboardStats{
			TotalPayroll:      totalBaseSalary + totalOTCosts,
			OvertimeCosts:     math.Round(totalOTCosts*100) / 100,
			PendingDisbursals: 0,
			CostReduction:     0,
		},
		PayrollTrends: trends,
		CostBreakdown: breakdown,
	}, nil
}

func (s *dashboardService) GetDailyPulse(ctx context.Context, tenantID uint) (modelDto.DailyPulseResponse, error) {
	now := time.Now().In(WIB)
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, WIB)
	todayEnd := todayStart.Add(24 * time.Hour)

	var (
		users           []model.User
		todayAttendance []model.Attendance
		pendingLeaves   []model.Leave
		pendingOTsActual []model.Overtime
		wg              sync.WaitGroup
		mu              sync.Mutex
	)

	wg.Add(4)
	go func() {
		defer wg.Done()
		u, _, _ := s.userRepo.FindAll(ctx, model.UserFilter{TenantID: tenantID}, []string{"role", "position"})
		mu.Lock(); users = u; mu.Unlock()
	}()
	go func() {
		defer wg.Done()
		a, _, _ := s.attendanceRepo.FindAll(ctx, model.AttendanceFilter{TenantID: tenantID, DateFrom: &todayStart, DateTo: &todayEnd}, nil, 0, 0)
		mu.Lock(); todayAttendance = a; mu.Unlock()
	}()
	go func() {
		defer wg.Done()
		l, _, _ := s.leaveRepo.FindAll(ctx, model.LeaveFilter{TenantID: tenantID, Status: model.LeaveStatusPending})
		mu.Lock(); pendingLeaves = l; mu.Unlock()
	}()
	go func() {
		defer wg.Done()
		o, _, _ := s.overtimeRepo.FindAll(ctx, model.OvertimeFilter{TenantID: tenantID, Status: model.OvertimeStatusPending})
		mu.Lock(); pendingOTsActual = o; mu.Unlock()
	}()
	wg.Wait()

	totalUsers := int64(len(users))
	presentCount := int64(len(todayAttendance))
	presentPercentage := 0.0
	if totalUsers > 0 {
		presentPercentage = (float64(presentCount) / float64(totalUsers)) * 100
	}

	totalOTHours := 0.0
	for _, a := range todayAttendance {
		if a.ClockOutTime != nil {
			// Simplified OT calculation: any work > 8 hours
			duration := a.ClockOutTime.Sub(a.ClockInTime).Hours()
			if duration > 8 {
				totalOTHours += (duration - 8)
			}
		}
	}
	avgOT := 0.0
	if presentCount > 0 {
		avgOT = totalOTHours / float64(presentCount)
	}

	// Hotline Requests (Leaves and OTs)
	hotline := make([]modelDto.HotlineRequest, 0)
	for _, l := range pendingLeaves {
		priority := "Medium"
		if l.EndDate.Sub(l.StartDate).Hours() > 48 {
			priority = "High"
		}
		hotline = append(hotline, modelDto.HotlineRequest{
			ID:          fmt.Sprintf("LV-%d", l.ID),
			UserName:    l.User.Name,
			Avatar:      s.getAvatar(l.User.Name, l.User.MediaUrl),
			Department:  l.User.Department,
			RequestType: "Leave",
			Priority:    priority,
		})
	}
	for _, o := range pendingOTsActual {
		hotline = append(hotline, modelDto.HotlineRequest{
			ID:          fmt.Sprintf("OT-%d", o.ID),
			UserName:    o.User.Name,
			Avatar:      s.getAvatar(o.User.Name, o.User.MediaUrl),
			Department:  o.User.Department,
			RequestType: "Overtime",
			Priority:    "Normal",
		})
	}

	// Performance Logic (Simulated for pulse)
	performanceMatrix := make([]modelDto.EmployeePerformanceItem, 0)
	for _, u := range users {
		score := 90 + (int(u.ID) % 10) // Simulate score
		performanceMatrix = append(performanceMatrix, modelDto.EmployeePerformanceItem{
			Name: u.Name, Avatar: s.getAvatar(u.Name, u.MediaUrl), Department: u.Department, Score: score,
		})
	}
	sort.Slice(performanceMatrix, func(i, j int) bool { return performanceMatrix[i].Score > performanceMatrix[j].Score })
	topPerformers := performanceMatrix
	if len(topPerformers) > 5 {
		topPerformers = topPerformers[:5]
	}

	return modelDto.DailyPulseResponse{
		Stats: modelDto.DailyPulseStats{
			PresentPercentage:     math.Round(presentPercentage*10) / 10,
			AvgOvertimeHours:      math.Round(avgOT*10) / 10,
			PendingApprovalsCount: int64(len(pendingLeaves) + len(pendingOTsActual)),
			AtRiskCount:           int64(totalUsers - presentCount), // Crude "at risk" for pulse: not present
		},
		HotlineRequests: hotline,
		TopPerformers:   topPerformers,
	}, nil
}
