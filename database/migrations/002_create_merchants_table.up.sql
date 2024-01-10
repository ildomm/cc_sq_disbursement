DROP TYPE IF EXISTS disbursement_frequencies;
CREATE TYPE disbursement_frequencies AS ENUM ('weekly', 'daily', 'monthly');

CREATE TABLE IF NOT EXISTS merchants (
    id                      UUID PRIMARY KEY DEFAULT UUID_GENERATE_V4(),
    reference               VARCHAR NOT NULL,
    email                   VARCHAR NOT NULL,
    live_at                 DATE NOT NULL,
    disbursement_frequency  DISBURSEMENT_FREQUENCIES NOT NULL,
    minimum_monthly_fee     DECIMAL(10,2) NOT NULL,
    created_at              TIMESTAMP(6) WITHOUT TIME ZONE NOT NULL,
    updated_at              TIMESTAMP(6) WITHOUT TIME ZONE NOT NULL
);