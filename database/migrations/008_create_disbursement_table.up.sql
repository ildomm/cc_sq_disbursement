CREATE TABLE IF NOT EXISTS merchant_disbursements (
    id                      UUID PRIMARY KEY DEFAULT UUID_GENERATE_V4(),
    merchant_id             UUID NOT NULL,

    disbursement_frequency  DISBURSEMENT_FREQUENCIES NOT NULL,
    orders_start_at         DATE NOT NULL,
    orders_end_at           DATE NOT NULL,

    fee_amount              DECIMAL(10,2) NOT NULL DEFAULT 0,
    fee_amount_correction   DECIMAL(10,2) NOT NULL DEFAULT 0,
    orders_sum_amount       DECIMAL(10,2) NOT NULL DEFAULT 0,
    orders_total_entries    INT NOT NULL DEFAULT 0,

    created_at              TIMESTAMP(6) WITHOUT TIME ZONE NOT NULL
);