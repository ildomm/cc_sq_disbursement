package database

import (
	"context"
	"github.com/google/uuid"
	"github.com/ildomm/cc_sq_disbursement/entities"
	"time"
)

type Querier interface {
	Close()

	CountMerchants(ctx context.Context) (int64, error)
	SelectMerchantByReference(ctx context.Context, reference string) (*entities.Merchant, error)
	SelectMerchant(ctx context.Context, id uuid.UUID) (*entities.Merchant, error)
	InsertOrder(ctx context.Context, order entities.Order) error
	CountOrders(ctx context.Context) (int64, error)
	SelectOrder(ctx context.Context, id string) (*entities.Order, error)

	SelectSumOrders(ctx context.Context, day time.Time) ([]entities.MerchantDisbursement, error)
	InsertDisbursement(ctx context.Context, disbursement entities.MerchantDisbursement) error

	SelectSumDisbursements(ctx context.Context, from, to time.Time, frequency entities.DisbursementFrequencies) ([]entities.MerchantDisbursement, error)

	MarkOrdersAsDisbursed(ctx context.Context, day time.Time) error
	SelectSumDisbursementsForMerchant(ctx context.Context, merchantId uuid.UUID, from, to time.Time, frequency entities.DisbursementFrequencies) (*entities.MerchantDisbursement, error)
}
