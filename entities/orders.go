package entities

import (
	"time"

	"github.com/google/uuid"
)

// Order represents the orders table in the database.
type Order struct {
	ID         string    `db:"id"`
	MerchantID uuid.UUID `db:"merchant_id"`
	Amount     float64   `db:"amount"`
	CreatedAt  time.Time `db:"created_at"`
	Disbursed  bool      `db:"disbursed"`
	FeeAmount  float64   `db:"fee_amount"`
}
