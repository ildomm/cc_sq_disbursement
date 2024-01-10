package entities

import (
	"time"

	"github.com/google/uuid"
)

// DisbursementFrequencies represents the disbursement frequencies enum type.
type DisbursementFrequencies string

const (
	DailyDisbursementFrequency   DisbursementFrequencies = "daily"
	WeeklyDisbursementFrequency  DisbursementFrequencies = "weekly"
	MonthlyDisbursementFrequency DisbursementFrequencies = "monthly"
)

// Merchant represents the merchants table in the database.
type Merchant struct {
	ID                    uuid.UUID               `db:"id"`
	Reference             string                  `db:"reference"`
	Email                 string                  `db:"email"`
	LiveAt                time.Time               `db:"live_at"`
	DisbursementFrequency DisbursementFrequencies `db:"disbursement_frequency"`
	MinimumMonthlyFee     float64                 `db:"minimum_monthly_fee"`
	CreatedAt             time.Time               `db:"created_at"`
	UpdatedAt             time.Time               `db:"updated_at"`
}
