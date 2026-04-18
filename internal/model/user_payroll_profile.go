package model

import "time"

type PtkpStatus string

const (
	PtkpTK0 PtkpStatus = "TK/0"
	PtkpTK1 PtkpStatus = "TK/1"
	PtkpTK2 PtkpStatus = "TK/2"
	PtkpTK3 PtkpStatus = "TK/3"
	PtkpK0  PtkpStatus = "K/0"
	PtkpK1  PtkpStatus = "K/1"
	PtkpK2  PtkpStatus = "K/2"
	PtkpK3  PtkpStatus = "K/3"
)

type UserPayrollProfile struct {
	ID                    uint       `gorm:"primaryKey" json:"id"`
	UserID                uint       `gorm:"uniqueIndex;not null" json:"user_id"`
	BankName              string     `gorm:"type:varchar(100)" json:"bank_name"`
	BankAccountNumber     string     `gorm:"type:varchar(50)" json:"bank_account_number"`
	BankAccountHolder     string     `gorm:"type:varchar(100)" json:"bank_account_holder"`
	BpjsHealthNumber      string     `gorm:"type:varchar(50)" json:"bpjs_health_number"`
	BpjsEmploymentNumber  string     `gorm:"type:varchar(50)" json:"bpjs_employment_number"`
	NpwpNumber            string     `gorm:"type:varchar(50)" json:"npwp_number"`
	PtkpStatus            PtkpStatus `gorm:"type:varchar(10);default:'TK/0'" json:"ptkp_status"`
	BasicSalary           float64    `gorm:"type:decimal(15,2);default:0" json:"basic_salary"`
	FixedAllowance        float64    `gorm:"type:decimal(15,2);default:0" json:"fixed_allowance"`
	DailyMealAllowance      float64  `gorm:"type:decimal(15,2);default:0" json:"daily_meal_allowance"`
	DailyTransportAllowance float64  `gorm:"type:decimal(15,2);default:0" json:"daily_transport_allowance"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
