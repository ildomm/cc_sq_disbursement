package entities

import (
	"time"

	"github.com/google/uuid"
)

// MerchantDisbursement represents the merchant_disbursements table in the database.
type MerchantDisbursement struct {
	ID                    uuid.UUID               `db:"id"`
	MerchantID            uuid.UUID               `db:"merchant_id"`
	DisbursementFrequency DisbursementFrequencies `db:"disbursement_frequency"`
	OrdersStartAt         time.Time               `db:"orders_start_at"`
	OrdersEndAt           time.Time               `db:"orders_end_at"`
	FeeAmount             float64                 `db:"fee_amount"`
	FeeAmountCorrection   float64                 `db:"fee_amount_correction"`
	OrdersSumAmount       float64                 `db:"orders_sum_amount"`
	OrdersTotalEntries    int                     `db:"orders_total_entries"`
	CreatedAt             time.Time               `db:"created_at"`
}
