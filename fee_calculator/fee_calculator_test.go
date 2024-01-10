package fee_calculator

import (
	"context"
	"github.com/ildomm/cc_sq_disbursement/test_helpers"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"

	"github.com/ildomm/cc_sq_disbursement/entities"
	"github.com/stretchr/testify/assert"
)

func TestCalculateFeeAmount(t *testing.T) {
	mockQuerier := test_helpers.NewMockQuerier()
	calculator := NewFeeCalculator(context.Background(), mockQuerier)

	tests := []struct {
		amount      float64
		expectedFee float64
		description string
	}{
		{25.0, 0.25, "Amount < 50, FeePercentage1"},
		{75.0, 0.71, "50 <= Amount <= 300, FeePercentage2"},
		{500.0, 4.25, "Amount > 300, FeePercentage3"},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result := calculator.CalculateFeeAmount(test.amount)
			assert.Equal(t, test.expectedFee, result, "Fee calculation mismatch")
		})
	}
}

func TestCalculateFeeAmountCorrection(t *testing.T) {
	mockQuerier := test_helpers.NewMockQuerier()
	mockQuerier.On("SelectSumDisbursementsForMerchant", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

	calculator := NewFeeCalculator(context.Background(), mockQuerier)

	tests := []struct {
		day                time.Time
		disbursement       entities.MerchantDisbursement
		expectedCorrection float64
		expectedError      bool
		description        string
	}{
		{time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
			entities.MerchantDisbursement{},
			0.0,
			false,
			"First disbursement of the month, no last month disbursement"},
		{time.Date(2022, 1, 15, 0, 0, 0, 0, time.UTC),
			entities.MerchantDisbursement{},
			0.0,
			false,
			"Not the first disbursement of the month"},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result, err := calculator.CalculateFeeAmountCorrection(test.day, test.disbursement)
			if test.expectedError {
				assert.Error(t, err, "Expected an error but got none")
			} else {
				assert.NoError(t, err, "Unexpected error")
				assert.Equal(t, test.expectedCorrection, result, "Fee amount correction mismatch")
			}
		})
	}
}

func TestCalculateFeeAmountCorrectionWithLastMonthPresent(t *testing.T) {
	mockQuerier := test_helpers.NewMockQuerier()
	mockQuerier.On("SelectSumDisbursementsForMerchant",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything).Return(&entities.MerchantDisbursement{FeeAmount: 0}, nil)

	mockQuerier.On("SelectMerchant",
		mock.Anything,
		mock.Anything).Return(&entities.Merchant{MinimumMonthlyFee: 10}, nil)

	calculator := NewFeeCalculator(context.Background(), mockQuerier)

	tests := []struct {
		day                time.Time
		disbursement       entities.MerchantDisbursement
		expectedCorrection float64
		expectedError      bool
		description        string
	}{
		{time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
			entities.MerchantDisbursement{},
			10.0,
			false,
			"First disbursement of the month, with last month disbursement present"},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result, err := calculator.CalculateFeeAmountCorrection(test.day, test.disbursement)
			if test.expectedError {
				assert.Error(t, err, "Expected an error but got none")
			} else {
				assert.NoError(t, err, "Unexpected error")
				assert.Equal(t, test.expectedCorrection, result, "Fee amount correction mismatch")
			}
		})
	}
}
