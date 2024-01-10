package database

import (
	"context"
	"fmt"
	"github.com/ildomm/cc_sq_disbursement/entities"
	"github.com/ildomm/cc_sq_disbursement/test_helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strconv"
	"strings"
	"testing"
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
