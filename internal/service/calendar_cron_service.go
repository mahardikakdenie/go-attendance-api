package service

import (
	"context"
	"fmt"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
	"go-attendance-api/internal/utils"
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

type CalendarCronService interface {
	Run()
	SendUpcomingEventReminders() error
}

type calendarCronService struct {
	hrRepo   repository.HrOpsRepository
	userRepo repository.UserRepository
	cron     *cron.Cron
}

func NewCalendarCronService(hrRepo repository.HrOpsRepository, userRepo repository.UserRepository) CalendarCronService {
	return &calendarCronService{
		hrRepo:   hrRepo,
		userRepo: userRepo,
		cron:     cron.New(),
	}
}

func (s *calendarCronService) Run() {
	// Schedule to run every day at 08:00 AM WIB
	// WIB is UTC+7, so 08:00 WIB is 01:00 UTC
	_, err := s.cron.AddFunc("0 8 * * *", func() {
		log.Println("[Cron] Starting Daily Calendar Event Reminders...")
		if err := s.SendUpcomingEventReminders(); err != nil {
			log.Printf("[Cron] Error sending reminders: %v\n", err)
		}
	})

	if err != nil {
		log.Fatalf("[Cron] Failed to schedule calendar reminders: %v\n", err)
	}

	s.cron.Start()
	log.Println("[Cron] Calendar Cron Service started successfully.")
}

func (s *calendarCronService) SendUpcomingEventReminders() error {
	ctx := context.Background()
	// Tomorrow date
	tomorrow := time.Now().In(time.FixedZone("WIB", 7*3600)).AddDate(0, 0, 1)
	
	events, err := s.hrRepo.FindUpcomingEvents(ctx, tomorrow)
	if err != nil {
		return err
	}

	if len(events) == 0 {
		log.Println("[Cron] No events found for tomorrow.")
		return nil
	}

	for _, event := range events {
		var targetUsers []model.User

		if event.IsAllUsers {
			// Fetch all active users in the tenant
			users, _, err := s.userRepo.FindAll(ctx, model.UserFilter{TenantID: event.TenantID}, []string{})
			if err != nil {
				log.Printf("[Cron] Error fetching users for tenant %d: %v\n", event.TenantID, err)
				continue
			}
			targetUsers = users
		} else {
			targetUsers = event.Users
		}

		if len(targetUsers) == 0 {
			continue
		}

		// Compose Email
		subject := fmt.Sprintf("Reminder: %s tomorrow!", event.Name)
		dateStr := event.Date.Format("Monday, 02 January 2006")
		
		for _, user := range targetUsers {
			html := utils.GetCalendarEventReminderTemplate(
				user.Name,
				event.Name,
				dateStr,
				string(event.Type),
				event.Description,
			)

			// Send Email asynchronously
			go func(email string, sub, body string) {
				emailCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				if err := utils.SendEmail(emailCtx, []string{email}, sub, body); err != nil {
					log.Printf("[Cron] Failed to send email to %s: %v\n", email, err)
				}
			}(user.Email, subject, html)
		}
		
		log.Printf("[Cron] Sent reminders for event '%s' to %d users.\n", event.Name, len(targetUsers))
	}

	return nil
}
