-- name: GetAll :many

SELECT
    id,
    order_source,
    loyalty_member_id,
    order_status,
    updated
FROM escape.orders;

-- name: CreateOrder :execresult

INSERT INTO escape.orders (
    order_source,
    loyalty_member_id,
    order_status,
    updated
)
VALUES (?, ?, ?, ?);

-- name: UpdateOrderStatus :execresult

Update escape.orders
SET order_status = ?, updated = ?
WHERE id = ?;