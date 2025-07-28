START TRANSACTION;

CREATE SCHEMA IF NOT EXISTS orders;

CREATE TABLE orders.order (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    order_number VARCHAR NOT NULL,
    status TEXT NOT NULL,
    total_price BIGINT NOT NULL,
    quantity INT NOT NULL,
    payment_method VARCHAR NOT NULL,
    shipping_fee INT NOT NULL,
    shipping_address TEXT NOT NULL,
    ordered_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    paid_at TIMESTAMP,
    memo TEXT
);

CREATE TABLE orders.order_items (
    id UUID PRIMARY KEY,
    order_id UUID NOT NULL REFERENCES orders.order(id) ON DELETE CASCADE,
    product_id UUID NOT NULL,
    product_name TEXT NOT NULL,
    product_price BIGINT NOT NULL,
    product_options JSONB,
    quantity INT NOT NULL
);

COMMIT;