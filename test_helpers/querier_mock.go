package test_helpers

import (
	"context"
	"github.com/google/uuid"
	"time"

	"github.com/ildomm/cc_sq_disbursement/entities"
	"github.com/stretchr/testify/mock"
)

type mockQuerier struct {
	mock.Mock
	keys map[string]map[string]interface{}
}

func NewMockQuerier() *mockQuerier {
	mocked := &mockQuerier{
		keys: make(map[string]map[string]interface{}),
	}

	mocked.keys["orders"] = make(map[string]interface{})

	return mocked
}

func (m *mockQuerier) Close() {
	m.Called()
}

func (m *mockQuerier) CountMerchants(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	if len(args) > 0 && args.Get(1) != nil {
		return 0, args.Error(0)
	}

	return 1, nil
}

func (m *mockQuerier) SelectMerchantByReference(ctx context.Context, reference string) (*entities.Merchant, error) {
	args := m.Called(ctx, reference)

	if len(args) > 0 && args.Get(1) != nil {
		return &entities.Merchant{}, args.Error(1)
	}
	if len(args) > 0 && args.Get(0) != nil {
		return args.Get(0).(*entities.Merchant), nil
	}

	for _, merchant := range m.keys["merchant"] {
		if merchant.(entities.Merchant).Reference == reference {
			_merchant := merchant.(entities.Merchant)
			return &_merchant, nil
		}
	}

	return nil, nil
}

func (m *mockQuerier) SelectMerchant(ctx context.Context, id uuid.UUID) (*entities.Merchant, error) {
	args := m.Called(ctx, id)

	if len(args) > 0 && args.Get(1) != nil {
		return &entities.Merchant{}, args.Error(1)
	}
	if len(args) > 0 && args.Get(0) != nil {
		return args.Get(0).(*entities.Merchant), nil
	}

	for _, merchant := range m.keys["merchant"] {
		if merchant.(entities.Merchant).ID == id {
			_merchant := merchant.(entities.Merchant)
			return &_merchant, nil
		}
	}

	return nil, nil
}

func (m *mockQuerier) InsertOrder(ctx context.Context, order entities.Order) error {
	args := m.Called(ctx, order)
	if len(args) > 0 && args.Get(0) != nil {
		return args.Error(0)
	}

	m.keys["orders"][order.ID] = order

	return nil
}

func (m *mockQuerier) CountOrders(ctx context.Context) (int64, error) {
	args := m.Called(ctx)

	if len(args) > 0 && args.Get(1) != nil {
		return 0, args.Error(0)
	}

	return int64(len(m.keys["orders"])), nil
}

func (m *mockQuerier) SelectOrder(ctx context.Context, id string) (*entities.Order, error) {
	args := m.Called(ctx, id)

	if len(args) > 0 && args.Get(1) != nil {
		return &entities.Order{}, args.Error(1)
	}
	if len(args) > 0 && args.Get(0) != nil {
		return args.Get(0).(*entities.Order), nil
	}

	for _, order := range m.keys["orders"] {
		if order.(entities.Order).ID == id {
			_order := order.(entities.Order)
			return &_order, nil
		}
	}

	return nil, nil
}

func (m *mockQuerier) InsertDisbursement(ctx context.Context, disbursement entities.MerchantDisbursement) error {
	args := m.Called(ctx, disbursement)
	if len(args) > 0 && args.Get(0) != nil {
		return args.Error(0)
	}

	m.keys["disbursement"][disbursement.ID.String()] = disbursement

	return nil
}

func (m *mockQuerier) SelectSumOrders(ctx context.Context, day time.Time) ([]entities.MerchantDisbursement, error) {
	args := m.Called(ctx, day)

	if len(args) > 0 && args.Get(1) != nil {
		return []entities.MerchantDisbursement{}, args.Error(1)
	}
	if len(args) > 0 && args.Get(0) != nil {
		return args.Get(0).([]entities.MerchantDisbursement), nil
	}

	return []entities.MerchantDisbursement{}, nil
}

func (m *mockQuerier) SelectSumDisbursements(ctx context.Context, from, to time.Time, frequency entities.DisbursementFrequencies) ([]entities.MerchantDisbursement, error) {
	args := m.Called(ctx, from, to, frequency)

	if len(args) > 0 && args.Get(1) != nil {
		return []entities.MerchantDisbursement{}, args.Error(1)
	}
	if len(args) > 0 && args.Get(0) != nil {
		return args.Get(0).([]entities.MerchantDisbursement), nil
	}

	return []entities.MerchantDisbursement{}, nil
}

func (m *mockQuerier) MarkOrdersAsDisbursed(ctx context.Context, day time.Time) error {
	args := m.Called(ctx, day)

	if len(args) > 0 && args.Get(0) != nil {
		return args.Error(0)
	}

	return nil
}

func (m *mockQuerier) SelectSumDisbursementsForMerchant(ctx context.Context, merchantId uuid.UUID, from, to time.Time, frequency entities.DisbursementFrequencies) (*entities.MerchantDisbursement, error) {
	args := m.Called(ctx, merchantId, merchantId, from, to, frequency)

	if len(args) > 0 && args.Get(1) != nil {
		return &entities.MerchantDisbursement{}, args.Error(1)
	}
	if len(args) > 0 && args.Get(0) != nil {
		return args.Get(0).(*entities.MerchantDisbursement), nil
	}

	return nil, nil
}
