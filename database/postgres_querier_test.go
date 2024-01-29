package database

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/ildomm/cc_sq_disbursement/entities"
	"github.com/ildomm/cc_sq_disbursement/test_helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestPostgresQuerier(t *testing.T) {
	testDB := test_helpers.NewTestDatabase(t)
	dbURL := testDB.ConnectionString(t) + "?sslmode=disable"

	ctx := context.Background()

	t.Run("NewPostgresQuerier_Success", func(t *testing.T) {
		querier, err := NewPostgresQuerier(ctx, dbURL)
		require.NoError(t, err)
		require.NotNil(t, querier)

		defer querier.Close()

		assert.NotNil(t, querier.dbConn)

		// Check if at least one migration has run by querying the database
		var extensionExists bool
		err = querier.dbConn.Get(&extensionExists, "SELECT EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'uuid-ossp')")
		require.NoError(t, err)
		assert.True(t, extensionExists, "uuid-ossp extension should exist")

		// Check the number of migration files in the folder
		migrationFiles, err := fs.ReadDir("migrations")
		require.NoError(t, err)

		// Count the number of migration files
		expectedNumMigrations := len(migrationFiles) / 2 // Each migration file has a corresponding .down.sql file

		// Query the "schema_migrations_service_file_loader" table to get the version
		var version string
		err = querier.dbConn.Get(&version, "SELECT version FROM schema_migrations")
		require.NoError(t, err)

		// Convert the version to an integer
		versionInt, err := strconv.Atoi(strings.TrimSpace(version))
		require.NoError(t, err)

		// Compare the number of migrations with the version in the database
		assert.Equal(t, expectedNumMigrations, versionInt, fmt.Sprintf("Number of migrations should match the version in the database. Expected: %d, Actual: %d", expectedNumMigrations, versionInt))
	})

	t.Run("NewPostgresQuerier_InvalidURL", func(t *testing.T) {
		_, err := NewPostgresQuerier(ctx, "invalid-url")
		require.Error(t, err)
	})
}

func setupTestQuerier(t *testing.T) (context.Context, func(t *testing.T), *PostgresQuerier) {
	testDB := test_helpers.NewTestDatabase(t)
	ctx := context.Background()
	q, err := NewPostgresQuerier(ctx, testDB.ConnectionString(t)+"?sslmode=disable")
	require.NoError(t, err)

	return ctx, func(t *testing.T) {
		testDB.Close(t)
	}, q
}

func TestDatabaseBasicOperations(t *testing.T) {
	ctx, teardownTest, q := setupTestQuerier(t)
	defer teardownTest(t)

	// Check that the table counts are 50, which implies that the migrations have been run
	count, err := q.CountMerchants(ctx)
	require.Equal(t, int64(50), count, "Rows counted")
	require.NoError(t, err)

	// Select an existing merchant
	merchant, err := q.SelectMerchantByReference(ctx, "reichert_group")
	require.NoError(t, err)
	require.NotNil(t, merchant)

	// File order
	order := test_helpers.SetupOrderTemplate()
	err = q.InsertOrder(ctx, order)
	require.NoError(t, err)

	count, err = q.CountOrders(ctx)
	require.Equal(t, int64(1), count, "Rows counted")
	require.NoError(t, err)

	orderReloaded, err := q.SelectOrder(ctx, order.ID)
	require.NoError(t, err)
	require.NotNil(t, orderReloaded)

	// Select an DailyDisbursement per merchants
	disbursements, err := q.SelectSumOrders(ctx, order.CreatedAt)
	require.NoError(t, err)
	require.NotNil(t, disbursements)
	require.Equal(t, 1, len(disbursements))

	err = q.InsertDisbursement(ctx, disbursements[0])
	require.NoError(t, err)

	// Select an DailyDisbursement per merchants, from previous insert
	disbursements, err = q.SelectSumDisbursements(ctx, order.CreatedAt, order.CreatedAt, entities.WeeklyDisbursementFrequency)
	require.NoError(t, err)
	require.NotNil(t, disbursements)
	require.Equal(t, 1, len(disbursements))

	// Mark orders as disbursed
	err = q.MarkOrdersAsDisbursed(ctx, order.CreatedAt)
	require.NoError(t, err)

	// Check that the order is marked as disbursed
	orderReloaded, err = q.SelectOrder(ctx, order.ID)
	require.NoError(t, err)
	require.NotNil(t, orderReloaded)
	require.Equal(t, true, orderReloaded.Disbursed)
}

func TestSelectSumDisbursementsForMerchant(t *testing.T) {
	ctx, teardownTest, querier := setupTestQuerier(t)
	defer teardownTest(t)

	// Create a merchant
	merchant := entities.Merchant{
		ID: uuid.MustParse("66312006-4d7e-45c4-9c28-788f4aa68a62"),
	}

	err := insertTestMerchants(ctx, querier)
	require.NoError(t, err)

	// Create a disbursement for the merchant
	disbursement := entities.MerchantDisbursement{
		MerchantID:            merchant.ID,
		DisbursementFrequency: entities.DailyDisbursementFrequency,
		OrdersStartAt:         time.Now().Add(-24 * time.Hour),
		OrdersEndAt:           time.Now(),
		FeeAmount:             0.0,
		FeeAmountCorrection:   0.0,
		OrdersSumAmount:       100.0,
		OrdersTotalEntries:    1,
	}
	err = querier.InsertDisbursement(ctx, disbursement)
	require.NoError(t, err)

	// Define the time range for the test
	startTime := time.Now().Add(-24 * time.Hour) // 1 day ago
	endTime := time.Now()

	// Call the method under test
	result, err := querier.SelectSumDisbursementsForMerchant(ctx, merchant.ID, startTime, endTime, entities.DailyDisbursementFrequency)
	require.NoError(t, err)

	// Assert that the result is as expected
	assert.NotNil(t, result)
}

func TestSelectMerchant(t *testing.T) {
	ctx, teardownTest, querier := setupTestQuerier(t)
	defer teardownTest(t)

	// Insert test merchant data
	err := insertTestMerchants(ctx, querier)
	require.NoError(t, err)

	t.Run("SelectExistingMerchant", func(t *testing.T) {
		// UUID of one of the predefined merchants
		testUUID := uuid.MustParse("66312006-4d7e-45c4-9c28-788f4aa68a62")

		// Call the method under test
		merchant, err := querier.SelectMerchant(ctx, testUUID)
		require.NoError(t, err)
		require.NotNil(t, merchant)
		assert.Equal(t, testUUID, merchant.ID)
		assert.Equal(t, "apadberg_group", merchant.Reference)
	})

	t.Run("SelectNonExistingMerchant", func(t *testing.T) {
		// UUID that does not exist in the database
		nonExistingUUID := uuid.New()

		// Call the method under test
		merchant, err := querier.SelectMerchant(ctx, nonExistingUUID)
		require.NoError(t, err)
		assert.Nil(t, merchant)
	})
}

func insertTestMerchants(ctx context.Context, querier *PostgresQuerier) error {
	const insertMerchantsSQL = `
		INSERT INTO merchants (id, reference, email, live_at, disbursement_frequency, minimum_monthly_fee, created_at, updated_at)
		VALUES
		    ('66312006-4d7e-45c4-9c28-788f4aa68a62', 'apadberg_group', 'ainfo@padberg-group.com', '2023-02-01', 'daily', 0.0, NOW(), NOW()),
		    ('61649242-a612-46ba-82d8-225542bb9576', 'adeckow_gibson', 'ainfo@deckow-gibson.com', '2022-12-14', 'daily', 0.0, NOW(), NOW()),
		    ('6616488f-c8b2-45dd-b29f-364d12a20238', 'aromaguera_and_sons', 'ainfo@romaguera-and-sons.com', '2022-12-10', 'daily', 0.0, NOW(), NOW()),
		    ('6b6d2b8a-f06c-4298-8f27-f33545eb5899', 'arosenbaum_parisian', 'ainfo@rosenbaum-parisian.com', '2022-11-09', 'weekly', 15.0, NOW(), NOW())`

	_, err := querier.dbConn.ExecContext(ctx, insertMerchantsSQL)
	return err
}
