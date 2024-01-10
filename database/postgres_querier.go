package database

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"github.com/ildomm/cc_sq_disbursement/entities"
	"github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
	"net/url"
	"time"
)

type PostgresQuerier struct {
	dbURL  string
	dbConn *sqlx.DB
	ctx    context.Context
}

func NewPostgresQuerier(ctx context.Context, url string) (*PostgresQuerier, error) {
	querier := PostgresQuerier{dbURL: url, ctx: ctx}

	_, err := pgx.ParseConfig(url)
	if err != nil {
		return &querier, err
	}

	// Open the connection using the DataDog wrapper methods
	querier.dbConn, err = sqlx.Open("pgx", url)
	if err != nil {
		return &querier, err
	}
	log.Print("opened database connection")

	// Ping the database to check that the connection is actually working
	err = querier.dbConn.Ping()
	if err != nil {
		return &querier, err
	}

	// Migrate the database
	err = querier.migrate()
	if err != nil {
		return &querier, err
	}
	log.Print("database migration complete")

	return &querier, nil
}

func (q *PostgresQuerier) Close() {
	q.dbConn.Close()
	log.Print("closed database connection")
}

var (
	//go:embed migrations/*.sql
	fs           embed.FS
	ErrorNilUUID = errors.New("UUID is nil")
)

func (q *PostgresQuerier) migrate() error {

	// Amend the database URl with custom parameter so that we can specify the
	// table name to be used to hold database migration state
	migrationsURL, err := q.migrationsURL()
	if err != nil {
		return err
	}

	// Load the migrations from our embedded resources
	d, err := iofs.New(fs, "migrations")
	if err != nil {
		return err
	}

	// Use a custom table name for schema migrations
	m, err := migrate.NewWithSourceInstance("iofs", d, migrationsURL)
	if err != nil {
		return err
	}

	// Migrate all the way up ...
	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

const (
	CustomMigrationParam = "x-migrations-table"
	CustomMigrationValue = "schema_migrations"
)

func (q *PostgresQuerier) migrationsURL() (string, error) {
	url, err := url.Parse(q.dbURL)
	if err != nil {
		return "", err
	}

	// Add the new Query parameter that specifies the table name for the migrations
	values := url.Query()
	values.Add(CustomMigrationParam, CustomMigrationValue)

	// Replace the Query parameters in the original URL & return
	url.RawQuery = values.Encode()
	return url.String(), nil
}

////////////////////////////////// Database Querier operations /////////////////////////////////////////////////////////

const countMerchantsFilesSQL = `SELECT COUNT(*) FROM merchants`

func (q *PostgresQuerier) CountMerchants(ctx context.Context) (int64, error) {
	return q.count(ctx, countMerchantsFilesSQL)
}

const countOrdersFilesSQL = `SELECT COUNT(*) FROM orders`

func (q *PostgresQuerier) CountOrders(ctx context.Context) (int64, error) {
	return q.count(ctx, countOrdersFilesSQL)
}

func (q *PostgresQuerier) count(ctx context.Context, sql string) (int64, error) {
	row := q.dbConn.QueryRowContext(ctx, sql)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const selectMerchantByReferenceSQL = `SELECT * FROM merchants WHERE reference = $1`

func (q *PostgresQuerier) SelectMerchantByReference(ctx context.Context, reference string) (*entities.Merchant, error) {
	var merchant entities.Merchant

	err := q.dbConn.GetContext(
		ctx,
		&merchant,
		selectMerchantByReferenceSQL,
		reference)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return &merchant, nil
}

const selectMerchantSQL = `SELECT * FROM merchants WHERE id = $1`

func (q *PostgresQuerier) SelectMerchant(ctx context.Context, id uuid.UUID) (*entities.Merchant, error) {
	var merchant entities.Merchant

	err := q.dbConn.GetContext(
		ctx,
		&merchant,
		selectMerchantSQL,
		id)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return &merchant, nil
}

const insertOrderSQL = `
	INSERT INTO orders ( id, merchant_id, amount, created_at, fee_amount )
	VALUES             ( $1, $2,          $3,     $4,         $5 )`

func (q *PostgresQuerier) InsertOrder(ctx context.Context, order entities.Order) error {
	err := q.dbConn.GetContext(
		ctx,
		&order.ID,
		insertOrderSQL,
		order.ID,
		order.MerchantID,
		order.Amount,
		order.CreatedAt,
		order.FeeAmount)

	// False positive error, ignore
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	return nil
}

const selectOrderSQL = `SELECT * FROM orders WHERE id = $1`

func (q *PostgresQuerier) SelectOrder(ctx context.Context, id string) (*entities.Order, error) {
	var order entities.Order

	err := q.dbConn.GetContext(
		ctx,
		&order,
		selectOrderSQL,
		id)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return &order, nil
}

const selectSumOrdersSQL = `
SELECT 
    uuid_generate_v4() AS id,
    merchant_id,
    $1                AS disbursement_frequency,
    created_at        AS orders_start_at,
    created_at        AS orders_end_at,    
    SUM(fee_amount)   AS fee_amount,
    0                 AS fee_amount_correction,
    SUM(amount)       AS orders_sum_amount,
    COUNT(*)          AS orders_total_entries     
FROM
    orders
WHERE
    created_at = $2 AND disbursed = false
GROUP BY
    merchant_id,
    created_at;
`

func (q *PostgresQuerier) SelectSumOrders(ctx context.Context, day time.Time) ([]entities.MerchantDisbursement, error) {
	var disbursements []entities.MerchantDisbursement

	err := q.dbConn.SelectContext(
		ctx,
		&disbursements,
		selectSumOrdersSQL,
		entities.DailyDisbursementFrequency,
		day)

	return disbursements, err
}

const insertDisbursementSQL = `
	INSERT INTO merchant_disbursements ( merchant_id, disbursement_frequency, orders_start_at, orders_end_at, fee_amount, fee_amount_correction, orders_sum_amount, orders_total_entries, created_at)
	VALUES                             ( $1,          $2,                     $3,              $4, 		      $5,         $6,                    $7,                $8,                   $9)`

func (q *PostgresQuerier) InsertDisbursement(ctx context.Context, disbursement entities.MerchantDisbursement) error {
	disbursement.CreatedAt = time.Now()

	err := q.dbConn.GetContext(
		ctx,
		&disbursement.ID,
		insertDisbursementSQL,
		disbursement.MerchantID,
		disbursement.DisbursementFrequency,
		disbursement.OrdersStartAt,
		disbursement.OrdersEndAt,
		disbursement.FeeAmount,
		disbursement.FeeAmountCorrection,
		disbursement.OrdersSumAmount,
		disbursement.OrdersTotalEntries,
		disbursement.CreatedAt)

	// False positive error, ignore
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	return nil
}

const selectSumDisbursementsSQL = `
SELECT 
    uuid_generate_v4() AS id,
    merchant_id,
    $1                         AS disbursement_frequency,
    DATE($2::timestamp)        AS orders_start_at,
    DATE($3::timestamp)        AS orders_end_at,
    SUM(fee_amount)            AS fee_amount,
    SUM(fee_amount_correction) AS fee_amount_correction,
    SUM(orders_sum_amount)     AS orders_sum_amount,
    SUM(orders_total_entries)  AS orders_total_entries, 
    NOW()                      AS created_at
FROM
    merchant_disbursements
WHERE
    orders_start_at >= $4 and orders_end_at <= $5 
GROUP BY
    merchant_id;
`

func (q *PostgresQuerier) SelectSumDisbursements(ctx context.Context, from, to time.Time, frequency entities.DisbursementFrequencies) ([]entities.MerchantDisbursement, error) {
	var disbursements []entities.MerchantDisbursement

	err := q.dbConn.SelectContext(
		ctx,
		&disbursements,
		selectSumDisbursementsSQL,
		frequency,
		from, to,
		from, to)

	return disbursements, err
}

const selectSumDisbursementsPerMerchantSQL = `
SELECT 
    uuid_generate_v4() AS id,
    merchant_id,
    $1                         AS disbursement_frequency,
    DATE($2::timestamp)        AS orders_start_at,
    DATE($3::timestamp)        AS orders_end_at,
    SUM(fee_amount)            AS fee_amount,
    SUM(fee_amount_correction) AS fee_amount_correction,
    SUM(orders_sum_amount)     AS orders_sum_amount,
    SUM(orders_total_entries)  AS orders_total_entries, 
    NOW()                      AS created_at
FROM
    merchant_disbursements
WHERE
    merchant_id = $4 
    AND orders_start_at >= $5 and orders_end_at <= $6 
GROUP BY
    merchant_id;
`

func (q *PostgresQuerier) SelectSumDisbursementsForMerchant(ctx context.Context, merchantId uuid.UUID, from, to time.Time, frequency entities.DisbursementFrequencies) (*entities.MerchantDisbursement, error) {
	var disbursement entities.MerchantDisbursement

	err := q.dbConn.GetContext(
		ctx,
		&disbursement,
		selectSumDisbursementsPerMerchantSQL,
		frequency,
		from, to,
		merchantId,
		from, to)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return &disbursement, err
}

const updateOrdersSQL = `
	UPDATE orders
	SET disbursed = true
	WHERE created_at = $1`

func (q *PostgresQuerier) MarkOrdersAsDisbursed(ctx context.Context, day time.Time) error {
	_, err := q.dbConn.ExecContext(ctx, updateOrdersSQL, day)

	return err
}
