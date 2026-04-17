package model

import (
	"time"
)

type GoalType string

const (
	GoalTypeKPI GoalType = "KPI"
	GoalTypeOKR GoalType = "OKR"
)

type GoalStatus string

const (
	GoalStatusInProgress GoalStatus = "IN_PROGRESS"
	GoalStatusCompleted  GoalStatus = "COMPLETED"
	GoalStatusCancelled  GoalStatus = "CANCELLED"
)

type PerformanceGoal struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	TenantID        uint       `gorm:"not null;index" json:"tenant_id"`
	UserID          uint       `gorm:"not null;index" json:"user_id"`
	Title           string     `gorm:"type:varchar(255);not null" json:"title"`
	Description     string     `gorm:"type:text" json:"description"`
	Type            GoalType   `gorm:"type:varchar(20);not null" json:"type"`
	TargetValue     float64    `gorm:"type:decimal(15,2)" json:"target_value"`
	CurrentProgress float64    `gorm:"type:decimal(15,2);default:0" json:"current_progress"`
	Unit            string     `gorm:"type:varchar(50)" json:"unit"`
	StartDate       time.Time  `json:"start_date"`
	EndDate         time.Time  `json:"end_date"`
	Status          GoalStatus `gorm:"type:varchar(20);default:'IN_PROGRESS'" json:"status"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

type CycleStatus string

const (
	CycleStatusDraft  CycleStatus = "DRAFT"
	CycleStatusActive CycleStatus = "ACTIVE"
	CycleStatusClosed CycleStatus = "CLOSED"
)

type PerformanceCycle struct {
	ID        uint        `gorm:"primaryKey" json:"id"`
	TenantID  uint        `gorm:"not null;index" json:"tenant_id"`
	Name      string      `gorm:"type:varchar(255);not null" json:"name"`
	StartDate time.Time   `json:"start_date"`
	EndDate   time.Time   `json:"end_date"`
	Status    CycleStatus `gorm:"type:varchar(20);default:'DRAFT'" json:"status"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type AppraisalStatus string

const (
	AppraisalStatusPending        AppraisalStatus = "PENDING"
	AppraisalStatusSelfReview     AppraisalStatus = "SELF_REVIEW"
	AppraisalStatusManagerReview  AppraisalStatus = "MANAGER_REVIEW"
	AppraisalStatusCompleted      AppraisalStatus = "COMPLETED"
)

type Appraisal struct {
	ID           uint            `gorm:"primaryKey" json:"id"`
	TenantID     uint            `gorm:"not null;index" json:"tenant_id"`
	CycleID      uint            `gorm:"not null;index" json:"cycle_id"`
	UserID       uint            `gorm:"not null;index" json:"user_id"`
	SelfScore    float64         `gorm:"type:decimal(5,2)" json:"self_score"`
	ManagerScore float64         `gorm:"type:decimal(5,2)" json:"manager_score"`
	FinalScore   float64         `gorm:"type:decimal(5,2)" json:"final_score"`
	FinalRating  string          `gorm:"type:varchar(100)" json:"final_rating"`
	Status       AppraisalStatus `gorm:"type:varchar(20);default:'PENDING'" json:"status"`
	Comments     string          `gorm:"type:text" json:"comments"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`

	Cycle PerformanceCycle `gorm:"foreignKey:CycleID" json:"cycle"`
	User  User             `gorm:"foreignKey:UserID" json:"user"`
}
