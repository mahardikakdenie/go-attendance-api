package model

import (
	"time"

	"github.com/google/uuid"
)

type TrialRequestStatus string

const (
	TrialRequestStatusNew       TrialRequestStatus = "NEW"
	TrialRequestStatusQualified TrialRequestStatus = "QUALIFIED"
	TrialRequestStatusApproved  TrialRequestStatus = "APPROVED"
	TrialRequestStatusRejected  TrialRequestStatus = "REJECTED"
)

type EmployeeCountRange string

const (
	EmployeeCountRange1To10   EmployeeCountRange = "1-10"
	EmployeeCountRange11To50  EmployeeCountRange = "11-50"
	EmployeeCountRange51To200 EmployeeCountRange = "51-200"
	EmployeeCountRange201Plus  EmployeeCountRange = "201+"
)

type TrialRequest struct {
	ID                 uuid.UUID          `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CompanyName        string             `gorm:"type:varchar(255);not null" json:"company_name"`
	ContactName        string             `gorm:"type:varchar(255);not null" json:"contact_name"`
	Email              string             `gorm:"type:varchar(255);unique;not null" json:"email"`
	PhoneNumber        string             `gorm:"type:varchar(50)" json:"phone_number"`
	EmployeeCountRange EmployeeCountRange `gorm:"type:varchar(20);not null" json:"employee_count_range"`
	Industry           string             `gorm:"type:varchar(100)" json:"industry"`
	Status             TrialRequestStatus `gorm:"type:varchar(20);default:'NEW'" json:"status"`
	CreatedAt          time.Time          `json:"created_at"`
}
