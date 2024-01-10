CREATE INDEX IF NOT EXISTS orders_pxt_merchant ON orders (merchant_id);
CREATE INDEX IF NOT EXISTS orders_pxt_creation ON orders (created_at);