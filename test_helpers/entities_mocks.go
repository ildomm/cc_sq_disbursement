package test_helpers

import (
	"github.com/google/uuid"
	"github.com/ildomm/cc_sq_disbursement/entities"
	"time"
)

func SetupOrderTemplate() entities.Order {
	return entities.Order{
		ID:         "an_order_id",
		MerchantID: uuid.New(),
		Amount:     1.01,
		CreatedAt:  time.Now(),
	}
}

func SetupMerchantTemplate() entities.Merchant {
	return entities.Merchant{
		ID:                    uuid.New(),
		Reference:             "a_merchant_reference",
		LiveAt:                time.Now(),
		DisbursementFrequency: "weekly",
		MinimumMonthlyFee:     1.01,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}
}
