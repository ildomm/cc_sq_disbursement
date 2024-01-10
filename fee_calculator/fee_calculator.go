package fee_calculator

import (
	"context"
	"github.com/ildomm/cc_sq_disbursement/database"
	"github.com/ildomm/cc_sq_disbursement/entities"
	"github.com/ildomm/cc_sq_disbursement/system"
	"math"
	"time"
)

const (

	// FeePercentage1 is the fee percentage for orders with an amount strictly smaller than 50 €.
	FeePercentage1 = 0.01

	// FeePercentage2 is the fee percentage for orders with an amount between 50 € and 300 €.
	FeePercentage2 = 0.0095

	// FeePercentage3 is the fee percentage for orders with an amount of 300 € or more.
	FeePercentage3 = 0.0085

	// AmountThreshold1 is the threshold for the first fee percentage.
	AmountThreshold1 = 50.0

	// AmountThreshold2 is the lower bound for the second fee percentage.
	AmountThreshold2 = 50.0

	// AmountThreshold3 is the upper bound for the second fee percentage and the lower bound for the third fee percentage.
	AmountThreshold3 = 300.0
)

// FeeCalculator is the struct that calculates the fee amount for an order and disbursement.
type FeeCalculator struct {
	ctx     context.Context
	querier database.Querier
}

func NewFeeCalculator(ctx context.Context, querier database.Querier) *FeeCalculator {
	calculator := FeeCalculator{
		ctx:     ctx,
		querier: querier,
	}

	return &calculator
}

// CalculateFeeAmount calculates the fee amount based on the order amount and fee percentages.
func (fc FeeCalculator) CalculateFeeAmount(amount float64) (feeAmount float64) {
	switch {
	case amount < AmountThreshold1:
		feeAmount = amount * FeePercentage1
	case amount >= AmountThreshold2 && amount <= AmountThreshold3:
		feeAmount = amount * FeePercentage2
	case amount > AmountThreshold3:
		feeAmount = amount * FeePercentage3
	}
	// Round the fee amount to two decimal places
	return math.Round(feeAmount*100) / 100
}

// CalculateFeeAmountCorrection calculates the fee amount correction for the disbursement
// It is calculated by comparing the fee amount for the last month with the minimum monthly fee
// Only the first disbursement of the month is corrected, and if the minimum monthly fee is greater than the fee amount
func (fc *FeeCalculator) CalculateFeeAmountCorrection(day time.Time, disbursement entities.MerchantDisbursement) (float64, error) {
	// Check if it is the first disbursement for the merchant in this month
	if day.Day() != 1 {
		return 0, nil
	}

	// Get the disbursement for the last month
	lastMonthDisbursement, err := fc.querier.SelectSumDisbursementsForMerchant(fc.ctx,
		disbursement.MerchantID,
		system.FirstDayOfLastMonth(day),
		system.LastDayOfLastMonth(day),
		entities.MonthlyDisbursementFrequency)
	if err != nil {
		return 0, err
	}

	// No disbursement for the last month, ignore
	if lastMonthDisbursement == nil {
		return 0, nil
	}

	// Get the merchant
	merchant, err := fc.querier.SelectMerchant(fc.ctx, disbursement.MerchantID)
	if err != nil {
		return 0, err
	}

	// Check if the fee amount needs to be corrected
	if merchant.MinimumMonthlyFee > lastMonthDisbursement.FeeAmount {
		return merchant.MinimumMonthlyFee - lastMonthDisbursement.FeeAmount, nil
	}

	return 0, nil
}
