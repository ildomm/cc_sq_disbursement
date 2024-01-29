package order_process

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/ildomm/cc_sq_disbursement/entities"
	"github.com/ildomm/cc_sq_disbursement/system"
	"github.com/ildomm/cc_sq_disbursement/test_helpers"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
	"time"
)

func TestPipelineHappyPath(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Mocked data setup
	dailyDisbursements := []entities.MerchantDisbursement{
		{
			ID:                    uuid.New(),
			MerchantID:            uuid.New(),
			DisbursementFrequency: entities.DailyDisbursementFrequency,
			OrdersStartAt:         time.Now(),
			OrdersEndAt:           time.Now(),
			FeeAmount:             10.0,
			FeeAmountCorrection:   0.5,
			OrdersSumAmount:       100.0,
			OrdersTotalEntries:    1,
			CreatedAt:             time.Now(),
		},
	}

	weeklyDisbursements := []entities.MerchantDisbursement{
		{
			ID:                    uuid.New(),
			MerchantID:            uuid.New(),
			DisbursementFrequency: entities.WeeklyDisbursementFrequency,
			OrdersStartAt:         time.Now(),
			OrdersEndAt:           time.Now(),
			FeeAmount:             10.0,
			FeeAmountCorrection:   0.5,
			OrdersSumAmount:       100.0,
			OrdersTotalEntries:    1,
			CreatedAt:             time.Now(),
		},
	}

	// Set up mock querier
	mockQuerier := test_helpers.NewMockQuerier()

	// Return mocked data for daily disbursements
	mockQuerier.On("SelectSumOrders", ctx, mock.Anything).Return(dailyDisbursements, nil)

	// Return mocked data for weekly disbursements
	mockQuerier.On("SelectSumDisbursements", ctx, mock.Anything, mock.Anything, entities.WeeklyDisbursementFrequency).Return(weeklyDisbursements, nil)

	mockQuerier.On("InsertDisbursement", ctx, mock.Anything).Return(nil)
	mockQuerier.On("MarkOrdersAsDisbursed", ctx, mock.Anything).Return(nil)

	// Initialize the pipeline
	p := NewPipeline(ctx, mockQuerier)

	// Define a test day
	testDay := time.Now()

	// Run the pipeline for the test day
	p.Run(testDay)

	// Assert that the expected methods were called on the mock querier
	mockQuerier.AssertExpectations(t)
}

func TestPipelineOnDatabaseError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a simulated database error
	dbError := errors.New("database error")

	// Set up mock querier
	mockQuerier := test_helpers.NewMockQuerier()

	// Simulate an error during SelectSumOrders
	mockQuerier.On("SelectSumOrders", ctx, mock.Anything).Return([]entities.MerchantDisbursement{}, dbError)

	// Initialize the pipeline
	p := NewPipeline(ctx, mockQuerier)

	// Define a test day
	testDay := time.Now()

	// Mock logger to capture log output
	mockLog := test_helpers.NewLogMocker()
	log.SetOutput(mockLog)

	// Run the pipeline for the test day
	p.Run(testDay)

	// Assert that the error was logged
	mockLog.AssertContains(t, "error creating daily disbursements: database error")

	// Assert that the expected method was called on the mock querier
	mockQuerier.AssertExpectations(t)
}

func TestPipelineSingleDailyDisbursement(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up a day for testing
	testDay := time.Now()

	// Mocked data for daily disbursement
	dailyDisbursement := []entities.MerchantDisbursement{
		{
			ID:                    uuid.New(),
			MerchantID:            uuid.New(),
			DisbursementFrequency: entities.DailyDisbursementFrequency,
			OrdersStartAt:         testDay,
			OrdersEndAt:           testDay,
			FeeAmount:             10.0,
			FeeAmountCorrection:   0.0,
			OrdersSumAmount:       100.0,
			OrdersTotalEntries:    1,
			CreatedAt:             testDay,
		},
	}

	// Set up mock querier
	mockQuerier := test_helpers.NewMockQuerier()
	mockQuerier.On("SelectSumOrders", ctx, testDay).Return(dailyDisbursement, nil)

	// Set up expectations for SelectSumDisbursements
	mockQuerier.On("SelectSumDisbursements", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]entities.MerchantDisbursement{}, nil)

	mockQuerier.On("InsertDisbursement", ctx, mock.Anything).Return(nil)
	mockQuerier.On("MarkOrdersAsDisbursed", ctx, testDay).Return(nil)

	// Initialize the pipeline
	p := NewPipeline(ctx, mockQuerier)

	// Mock logger to capture log output
	mockLog := test_helpers.NewLogMocker()
	log.SetOutput(mockLog)

	// Run the pipeline for the test day
	p.Run(testDay)

	// Assert that the expected methods were called on the mock querier
	mockQuerier.AssertCalled(t, "SelectSumOrders", ctx, testDay)
	mockQuerier.AssertNumberOfCalls(t, "InsertDisbursement", len(dailyDisbursement))
	mockQuerier.AssertCalled(t, "MarkOrdersAsDisbursed", ctx, testDay)

	// Assert that the log contains expected messages
	mockLog.AssertContains(t, "start processing orders from day")
	mockLog.AssertContains(t, "finish processing orders from day")
}

func TestPipelineSingleDailyDisbursementWithFeeAmountCorrection(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up a day for testing
	testDay := time.Now()

	// Mocked data for daily disbursement
	dailyDisbursement := []entities.MerchantDisbursement{
		{
			ID:                    uuid.New(),
			MerchantID:            uuid.New(),
			DisbursementFrequency: entities.DailyDisbursementFrequency,
			OrdersStartAt:         testDay,
			OrdersEndAt:           testDay,
			FeeAmount:             10.0,
			FeeAmountCorrection:   0.5,
			OrdersSumAmount:       100.0,
			OrdersTotalEntries:    1,
			CreatedAt:             testDay,
		},
	}

	// Set up mock querier
	mockQuerier := test_helpers.NewMockQuerier()
	mockQuerier.On("SelectSumOrders", ctx, testDay).Return(dailyDisbursement, nil)

	// Set up expectations for SelectSumDisbursements
	mockQuerier.On("SelectSumDisbursements", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]entities.MerchantDisbursement{}, nil)

	mockQuerier.On("InsertDisbursement", ctx, mock.Anything).Return(nil)
	mockQuerier.On("MarkOrdersAsDisbursed", ctx, testDay).Return(nil)

	// Initialize the pipeline
	p := NewPipeline(ctx, mockQuerier)

	// Mock logger to capture log output
	mockLog := test_helpers.NewLogMocker()
	log.SetOutput(mockLog)

	// Run the pipeline for the test day
	p.Run(testDay)

	// Assert that the expected methods were called on the mock querier
	mockQuerier.AssertCalled(t, "SelectSumOrders", ctx, testDay)
	mockQuerier.AssertNumberOfCalls(t, "InsertDisbursement", len(dailyDisbursement))
	mockQuerier.AssertCalled(t, "MarkOrdersAsDisbursed", ctx, testDay)

	// Assert that the log contains expected messages
	mockLog.AssertContains(t, "start processing orders from day")
	mockLog.AssertContains(t, "finish processing orders from day")
}

func TestPipelineWeeklyDailyDisbursement(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up a day for testing
	testDay := time.Now()

	// Mocked data for daily disbursement
	dailyDisbursement := []entities.MerchantDisbursement{
		{
			ID:                    uuid.New(),
			MerchantID:            uuid.New(),
			DisbursementFrequency: entities.DailyDisbursementFrequency,
			OrdersStartAt:         testDay,
			OrdersEndAt:           testDay,
			FeeAmount:             10.0,
			FeeAmountCorrection:   0.0,
			OrdersSumAmount:       100.0,
			OrdersTotalEntries:    1,
			CreatedAt:             testDay,
		},
	}

	// Mocked data for weekly disbursement
	weeklyDisbursement := []entities.MerchantDisbursement{
		{
			ID:                    uuid.New(),
			MerchantID:            uuid.New(),
			DisbursementFrequency: entities.WeeklyDisbursementFrequency,
			OrdersStartAt:         testDay,
			OrdersEndAt:           testDay,
			FeeAmount:             10.0,
			FeeAmountCorrection:   0.0,
			OrdersSumAmount:       100.0,
			OrdersTotalEntries:    1,
			CreatedAt:             testDay,
		},
	}

	// Set up mock querier
	mockQuerier := test_helpers.NewMockQuerier()
	mockQuerier.On("SelectSumOrders", ctx, testDay).Return(dailyDisbursement, nil)

	// Set up expectations for SelectSumDisbursements
	mockQuerier.On("SelectSumDisbursements", ctx, mock.Anything, mock.Anything, mock.Anything).Return(weeklyDisbursement, nil)

	mockQuerier.On("InsertDisbursement", ctx, mock.Anything).Return(nil)
	mockQuerier.On("MarkOrdersAsDisbursed", ctx, testDay).Return(nil)

	// Initialize the pipeline
	p := NewPipeline(ctx, mockQuerier)

	// Mock logger to capture log output
	mockLog := test_helpers.NewLogMocker()
	log.SetOutput(mockLog)

	// Run the pipeline for the test day
	p.Run(testDay)

	// Assert that the expected methods were called on the mock querier
	mockQuerier.AssertCalled(t, "SelectSumOrders", ctx, testDay)
}
func TestPipelineSingleMonthlyDisbursement(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up a day for testing
	testDay := time.Now()

	// Mocked data for daily disbursement
	dailyDisbursement := []entities.MerchantDisbursement{
		{
			ID:                    uuid.New(),
			MerchantID:            uuid.New(),
			DisbursementFrequency: entities.DailyDisbursementFrequency,
			OrdersStartAt:         testDay,
			OrdersEndAt:           testDay,
			FeeAmount:             10.0,
			FeeAmountCorrection:   0.0,
			OrdersSumAmount:       100.0,
			OrdersTotalEntries:    1,
			CreatedAt:             testDay,
		},
	}

	// Mocked data for monthly disbursement
	monthlyDisbursement := []entities.MerchantDisbursement{
		{
			ID:                    uuid.New(),
			MerchantID:            uuid.New(),
			DisbursementFrequency: entities.MonthlyDisbursementFrequency,
			OrdersStartAt:         testDay,
			OrdersEndAt:           testDay,
			FeeAmount:             10.0,
			FeeAmountCorrection:   0.0,
			OrdersSumAmount:       100.0,
			OrdersTotalEntries:    1,
			CreatedAt:             testDay,
		},
	}

	// Set up mock querier
	mockQuerier := test_helpers.NewMockQuerier()
	mockQuerier.On("SelectSumOrders", ctx, testDay).Return(dailyDisbursement, nil)

	// Set up expectations for SelectSumDisbursements
	mockQuerier.On("SelectSumDisbursements", ctx, mock.Anything, mock.Anything, mock.Anything).Return(monthlyDisbursement, nil)

	mockQuerier.On("InsertDisbursement", ctx, mock.Anything).Return(nil)
	mockQuerier.On("MarkOrdersAsDisbursed", ctx, testDay).Return(nil)

	// Initialize the pipeline
	p := NewPipeline(ctx, mockQuerier)

	// Mock logger to capture log output
	mockLog := test_helpers.NewLogMocker()
	log.SetOutput(mockLog)

	// Run the pipeline for the test day
	p.Run(testDay)

	// Assert that the expected methods were called on the mock querier
	mockQuerier.AssertCalled(t, "SelectSumOrders", ctx, testDay)
}

func TestPipelineFeeAmountCorrectionWithoutLastMonth(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up a day for testing
	testDay := time.Now()

	// Mocked data for daily disbursement
	dailyDisbursement := []entities.MerchantDisbursement{
		{
			ID:                    uuid.New(),
			MerchantID:            uuid.New(),
			DisbursementFrequency: entities.DailyDisbursementFrequency,
			OrdersStartAt:         testDay,
			OrdersEndAt:           testDay,
			FeeAmount:             10.0,
			FeeAmountCorrection:   0.0,
			OrdersSumAmount:       100.0,
			OrdersTotalEntries:    1,
			CreatedAt:             testDay,
		},
	}

	// Mocked data for monthly disbursement
	monthlyDisbursement := []entities.MerchantDisbursement{}

	// Set up mock querier
	mockQuerier := test_helpers.NewMockQuerier()
	mockQuerier.On("SelectSumOrders", ctx, testDay).Return(dailyDisbursement, nil)

	// Set up expectations for SelectSumDisbursements
	mockQuerier.On("SelectSumDisbursements", ctx, mock.Anything, mock.Anything, mock.Anything).Return(monthlyDisbursement, nil)

	mockQuerier.On("InsertDisbursement", ctx, mock.Anything).Return(nil)
	mockQuerier.On("MarkOrdersAsDisbursed", ctx, testDay).Return(nil)

	// Initialize the pipeline
	p := NewPipeline(ctx, mockQuerier)

	// Mock logger to capture log output
	mockLog := test_helpers.NewLogMocker()
	log.SetOutput(mockLog)

	// Run the pipeline for the test day
	p.Run(testDay)

	// Assert that the expected methods were called on the mock querier
	mockQuerier.AssertCalled(t, "SelectSumOrders", ctx, testDay)
}

func TestPipelineMonthlyDisbursements(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Mock querier setup
	mockQuerier := test_helpers.NewMockQuerier()

	// Mocked data for monthly disbursement
	monthlyDisbursement := []entities.MerchantDisbursement{
		{
			ID:                    uuid.New(),
			MerchantID:            uuid.New(),
			DisbursementFrequency: entities.MonthlyDisbursementFrequency,
			OrdersStartAt:         system.FirstDayOfLastMonth(time.Now()),
			OrdersEndAt:           system.LastDayOfLastMonth(time.Now()),
			FeeAmount:             10.0,
			OrdersSumAmount:       100.0,
			OrdersTotalEntries:    1,
			CreatedAt:             time.Now(),
		},
	}

	// Initialize the pipeline
	p := NewPipeline(ctx, mockQuerier)

	// Test 1: Run on the first day of the month
	firstDayOfMonth := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	mockQuerier.On("SelectSumDisbursements", ctx, mock.Anything, mock.Anything, entities.MonthlyDisbursementFrequency).Return(monthlyDisbursement, nil)
	mockQuerier.On("InsertDisbursement", ctx, mock.Anything).Return(nil)

	err := p.monthlyDisbursements(firstDayOfMonth)
	require.NoError(t, err)
	mockQuerier.AssertNumberOfCalls(t, "InsertDisbursement", len(monthlyDisbursement))

	// Test 2: Run on a day other than the first of the month
	nonFirstDayOfMonth := time.Date(2024, 2, 2, 0, 0, 0, 0, time.UTC)
	err = p.monthlyDisbursements(nonFirstDayOfMonth)
	require.NoError(t, err)
	mockQuerier.AssertNumberOfCalls(t, "InsertDisbursement", len(monthlyDisbursement)) // Assert that no new calls were made
}
