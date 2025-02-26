START TRANSACTION;

CREATE SCHEMA IF NOT EXISTS escape;

CREATE TABLE escape.orders (
    id CHAR(36) NOT NULL DEFAULT (UUID()),
    order_source INT NOT NULL,
    loyalty_member_id CHAR(36) NOT NULL,
    order_status INT NOT NULL,
    updated TIMESTAMP NULL,
    CONSTRAINT pk_orders PRIMARY KEY (id)
);

COMMIT;