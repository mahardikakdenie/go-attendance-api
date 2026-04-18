package model

import "time"

type SubscriptionStatus string

const (
	SubscriptionStatusActive   SubscriptionStatus = "Active"
	SubscriptionStatusPastDue  SubscriptionStatus = "Past Due"
	SubscriptionStatusCanceled SubscriptionStatus = "Canceled"
	SubscriptionStatusTrial    SubscriptionStatus = "Trial"
)

type BillingCycle string

const (
	BillingCycleMonthly BillingCycle = "Monthly"
	BillingCycleYearly  BillingCycle = "Yearly"
)

type Subscription struct {
	ID              uint               `gorm:"primaryKey" json:"id"`
	TenantID        uint               `gorm:"not null;uniqueIndex" json:"tenant_id"`
	Tenant          *Tenant            `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	Plan            string             `gorm:"type:varchar(50);not null" json:"plan"`
	BillingCycle    BillingCycle       `gorm:"type:varchar(20);not null" json:"billing_cycle"`
	Amount          float64            `gorm:"type:decimal(15,2);not null" json:"amount"`
	Status          SubscriptionStatus `gorm:"type:varchar(20);not null;default:'Trial'" json:"status"`
	NextBillingDate time.Time          `json:"next_billing_date"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
}
