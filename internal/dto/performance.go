package modelDto

import (
	"go-attendance-api/internal/model"
)

type CreateGoalRequest struct {
	UserID         uint           `json:"user_id" binding:"required"`
	Title          string         `json:"title" binding:"required"`
	Description    string         `json:"description"`
	Type           model.GoalType `json:"type" binding:"required,oneof=KPI OKR"`
	TargetValue    float64        `json:"target_value" binding:"required"`
	Unit           string         `json:"unit"`
	StartDate      string         `json:"start_date" binding:"required" example:"2026-01-01"`
	EndDate        string         `json:"end_date" binding:"required" example:"2026-12-31"`
}

type UpdateGoalProgressRequest struct {
	CurrentProgress float64 `json:"current_progress" binding:"required"`
}

type SubmitSelfReviewRequest struct {
	SelfScore float64 `json:"self_score" binding:"required"`
	Comments  string  `json:"comments"`
}

type GoalResponse struct {
	ID              uint             `json:"id"`
	UserID          uint             `json:"user_id"`
	Title           string           `json:"title"`
	Description     string           `json:"description"`
	Type            model.GoalType   `json:"type"`
	TargetValue     float64          `json:"target_value"`
	CurrentProgress float64          `json:"current_progress"`
	Unit            string           `json:"unit"`
	StartDate       string           `json:"start_date"`
	EndDate         string           `json:"end_date"`
	Status          model.GoalStatus `json:"status"`
}

type CycleResponse struct {
	ID        uint              `json:"id"`
	Name      string            `json:"name"`
	StartDate string            `json:"start_date"`
	EndDate   string            `json:"end_date"`
	Status    model.CycleStatus `json:"status"`
}

type AppraisalResponse struct {
	ID           uint                  `json:"id"`
	CycleID      uint                  `json:"cycle_id"`
	UserID       uint                  `json:"user_id"`
	UserName     string                `json:"user_name"`
	SelfScore    float64               `json:"self_score"`
	ManagerScore float64               `json:"manager_score"`
	FinalScore   float64               `json:"final_score"`
	FinalRating  string                `json:"final_rating"`
	Status       model.AppraisalStatus `json:"status"`
	Comments     string                `json:"comments"`
}
