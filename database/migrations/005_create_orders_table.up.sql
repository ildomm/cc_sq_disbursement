CREATE TABLE IF NOT EXISTS orders (
    id           VARCHAR PRIMARY KEY,
    merchant_id  UUID NOT NULL, /* DO NOT make it a Referential Integrity Constraint for performance reasons ONLY */
    amount       DECIMAL(10,2) NOT NULL,
    created_at   DATE NOT NULL
);