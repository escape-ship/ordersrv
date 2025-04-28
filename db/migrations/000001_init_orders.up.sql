START TRANSACTION;

CREATE SCHEMA IF NOT EXISTS orders;

CREATE TABLE orders.order (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    user_email 
    product_id UUID NOT NULL,
    product_name VARCHAR(255) UNIQUE NOT NULL,
    product_price  BIGINT NOT NULL,
    product_imageUrl VARCHAR NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
);  

COMMIT;