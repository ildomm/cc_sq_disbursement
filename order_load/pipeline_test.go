package order_load

import (
	"context"
	"errors"
	"github.com/ildomm/cc_sq_disbursement/system"
	"github.com/ildomm/cc_sq_disbursement/test_helpers"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
	"time"
)

func TestPipelineHappyPath(t *testing.T) {
	err := system.SetGlobalTimezoneUTC()
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockQuerier := test_helpers.NewMockQuerier()
	merchant := test_helpers.SetupMerchantTemplate()
	mockQuerier.On("SelectMerchantByReference", mock.Anything, mock.Anything).Return(
		&merchant, nil)
	mockQuerier.On("SelectOrder", mock.Anything, mock.Anything)
	mockQuerier.On("InsertOrder", mock.Anything, mock.Anything)

	mockLog := test_helpers.NewLogMocker()
	log.SetOutput(mockLog)

	test_helpers.SetupCSVOrder(WaitingPath)
	time.Sleep(time.Duration(1) * time.Second)

	p := NewPipeline(ctx, mockQuerier)
	go p.Run()

	time.Sleep(time.Duration(2) * time.Second)
	ctx.Done()

	// Check that the log contains the expected messages
	mockLog.AssertContains(t, "starting loading orders")
	mockLog.AssertContains(t, "files to process: 1")
	mockLog.AssertContains(t, "orders imported successfully from CSV file")

	// Check mock querier expectations
	mockQuerier.AssertExpectations(t)

	// Check the total number of orders in the database
	mockQuerier.On("CountOrders", mock.Anything)
	mockQuerier.On("SelectOrder", mock.Anything, mock.Anything)
	count, err := mockQuerier.CountOrders(ctx)
	require.Equal(t, int64(1), count, "Rows counted")
	require.NoError(t, err)

	// Check the order persistence
	orderReloaded, err := mockQuerier.SelectOrder(ctx, "any_order_id")
	require.NoError(t, err)
	require.NotNil(t, orderReloaded)
	require.Equal(t, 0.95, orderReloaded.FeeAmount)

	// Erase the example file
	test_helpers.RemoveCSVOrder(ImportedPath)
}

func TestPipelineOnWrongMerchant(t *testing.T) {
	err := system.SetGlobalTimezoneUTC()
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockQuerier := test_helpers.NewMockQuerier()
	mockQuerier.On("SelectMerchantByReference", mock.Anything, mock.Anything).Return(
		nil, errors.New("merchant not found"))

	mockLog := test_helpers.NewLogMocker()
	log.SetOutput(mockLog)

	test_helpers.SetupCSVOrder(WaitingPath)

	p := NewPipeline(ctx, mockQuerier)
	go p.Run()

	time.Sleep(time.Duration(1) * time.Second)
	ctx.Done()

	// Check that the log contains the expected messages
	mockLog.AssertContains(t, "starting loading orders")
	mockLog.AssertContains(t, "files to process: 1")
	mockLog.AssertContains(t, "error checking if merchant exists")
	mockLog.AssertNotContains(t, "orders imported successfully from CSV file")

	// Check mock querier expectations
	mockQuerier.AssertExpectations(t)

	// Check the total number of orders in the database
	mockQuerier.On("CountOrders", mock.Anything)
	count, err := mockQuerier.CountOrders(ctx)
	require.Equal(t, int64(0), count, "Rows counted")
	require.NoError(t, err)

	// Erase the example file
	test_helpers.RemoveCSVOrder(FailedPath)
}

func TestPipelineWhenOrderExists(t *testing.T) {
	err := system.SetGlobalTimezoneUTC()
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockQuerier := test_helpers.NewMockQuerier()
	merchant := test_helpers.SetupMerchantTemplate()
	mockQuerier.On("SelectMerchantByReference", mock.Anything, mock.Anything).Return(
		&merchant, nil)
	mockQuerier.On("SelectOrder", mock.Anything, mock.Anything)
	mockQuerier.On("InsertOrder", mock.Anything, mock.Anything)

	mockLog := test_helpers.NewLogMocker()
	log.SetOutput(mockLog)

	test_helpers.SetupCSVOrder(WaitingPath)

	p := NewPipeline(ctx, mockQuerier)
	// First execution
	go p.Run()

	time.Sleep(time.Duration(1) * time.Second)
	ctx.Done()

	// Basic check in logs
	mockLog.AssertContains(t, "orders imported successfully from CSV file")

	// Check mock querier expectations
	mockQuerier.AssertExpectations(t)

	// Check the total number of orders in the database
	mockQuerier.On("CountOrders", mock.Anything)
	count, err := mockQuerier.CountOrders(ctx)
	require.Equal(t, int64(1), count, "Rows counted")
	require.NoError(t, err)

	// Erase the example file
	test_helpers.RemoveCSVOrder(ImportedPath)

	// Second execution
	test_helpers.SetupCSVOrder(WaitingPath)
	go p.Run()

	time.Sleep(time.Duration(1) * time.Second)
	ctx.Done()

	// Basic check in logs
	mockLog.AssertContains(t, "order already exists")

	// Check the total number of orders in the database
	mockQuerier.On("CountOrders", mock.Anything)
	count, err = mockQuerier.CountOrders(ctx)
	require.Equal(t, int64(1), count, "Rows counted")
	require.NoError(t, err)

	// Erase the example file
	test_helpers.RemoveCSVOrder(ImportedPath)
}

func TestPipelineBuildOrderHappyPath(t *testing.T) {
	// TODO: Implement
	// Not doing right now because we are running out of time
}

func TestPipelineBuildOrderOnInvalidValues(t *testing.T) {
	// TODO: Implement
	// Not doing right now because we are running out of time
}
