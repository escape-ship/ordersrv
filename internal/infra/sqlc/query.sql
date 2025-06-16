-- name: InsertOrder :one
INSERT INTO orders.order (
    id, user_id, order_number, status, total_price, quantity, payment_method, shipping_fee, shipping_address, ordered_at, paid_at, memo
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, COALESCE($10, CURRENT_TIMESTAMP), $11, $12
) RETURNING id;

-- name: InsertOrderItem :exec
INSERT INTO orders.order_items (
    id, order_id, product_id, product_name, product_price, quantity
) VALUES (
    $1, $2, $3, $4, $5, $6
);

-- name: GetOrderWithItems :one
SELECT * FROM orders.order WHERE id = $1;

-- name: GetOrderItems :many
SELECT * FROM orders.order_items WHERE order_id = $1;

-- name: GetAllOrders :many
SELECT * FROM orders.order;

-- name: UpdateOrderStatus :exec
UPDATE orders.order
SET status = $2,
    updated_at = NOW()
WHERE id = $1;