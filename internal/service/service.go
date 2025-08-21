package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/escape-ship/ordersrv/internal/infra/sqlc/postgresql"
	"github.com/escape-ship/ordersrv/pkg/postgres"
	pb "github.com/escape-ship/protos/gen"
	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

type OrderController struct {
	pb.UnimplementedOrderServiceServer
	pg     postgres.DBEngine
	logger *slog.Logger
}

func NewOrderController(pg postgres.DBEngine) *OrderController {
	return &OrderController{
		pg:     pg,
		logger: slog.New(slog.NewJSONHandler(os.Stdout, nil)),
	}
}

func (s *OrderController) InsertOrder(ctx context.Context, req *pb.InsertOrderRequest) (*pb.InsertOrderResponse, error) {
	s.logger.Info("InsertOrder: Starting order insertion",
		"user_id", req.UserId,
		"order_number", req.OrderNumber,
		"total_price", req.TotalPrice,
		"items_count", len(req.Items))

	db := s.pg.GetDB()
	if db == nil {
		s.logger.Error("InsertOrder: Failed to get database connection")
		return nil, fmt.Errorf("database connection is nil")
	}

	querier := postgresql.New(db)

	tx, err := db.Begin()
	if err != nil {
		s.logger.Error("InsertOrder: Failed to begin transaction", "error", err)
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	qtx := querier.WithTx(tx)
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				s.logger.Error("InsertOrder: Failed to rollback transaction", "error", rollbackErr)
			} else {
				s.logger.Info("InsertOrder: Transaction rolled back successfully")
			}
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				s.logger.Error("InsertOrder: Failed to commit transaction", "error", commitErr)
				err = commitErr
			} else {
				s.logger.Info("InsertOrder: Transaction committed successfully")
			}
		}
	}()

	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		s.logger.Error("InsertOrder: Invalid user ID format", "user_id", req.UserId, "error", err)
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	orderID := uuid.New()
	s.logger.Info("InsertOrder: Generated order ID", "order_id", orderID.String())

	orderParams := postgresql.InsertOrderParams{
		ID:              orderID,
		UserID:          userId,
		OrderNumber:     req.OrderNumber,
		Status:          string(OrderStateReceived),
		TotalPrice:      req.TotalPrice,
		Quantity:        req.Quantity,
		PaymentMethod:   req.PaymentMethod,
		ShippingFee:     req.ShippingFee,
		ShippingAddress: req.ShippingAddress,
		Column10:        nil, // ordered_at (nil이면 CURRENT_TIMESTAMP)
		PaidAt:          parseNullTime(req.PaidAt),
		Memo:            parseNullString(req.Memo),
	}

	s.logger.Info("InsertOrder: Inserting order into database", "order_id", orderID.String())
	_, err = qtx.InsertOrder(ctx, orderParams)
	if err != nil {
		s.logger.Error("InsertOrder: Failed to insert order",
			"order_id", orderID.String(),
			"user_id", req.UserId,
			"error", err)
		return nil, fmt.Errorf("failed to insert order: %w", err)
	}
	s.logger.Info("InsertOrder: Order inserted successfully", "order_id", orderID.String())

	s.logger.Info("InsertOrder: Processing order items", "items_count", len(req.Items))
	for i, item := range req.Items {
		s.logger.Info("InsertOrder: Processing item",
			"item_index", i,
			"product_id", item.ProductId,
			"product_name", item.ProductName,
			"quantity", item.Quantity)

		itemID := uuid.New()

		var options map[string]interface{}
		if err := json.Unmarshal([]byte(item.ProductOptions), &options); err != nil {
			s.logger.Error("InsertOrder: Invalid product_options for item",
				"item_index", i,
				"product_id", item.ProductId,
				"product_options", item.ProductOptions,
				"error", err)
			continue
		}

		// 👉 map → JSON → RawMessage
		rawOptions, err := json.Marshal(options)
		if err != nil {
			s.logger.Error("InsertOrder: Failed to marshal product_options for item",
				"item_index", i,
				"product_id", item.ProductId,
				"error", err)
			continue
		}

		productUUID, err := uuid.Parse(item.ProductId)
		if err != nil {
			s.logger.Error("InsertOrder: Invalid product ID format",
				"item_index", i,
				"product_id", item.ProductId,
				"error", err)
			return nil, fmt.Errorf("invalid product ID for item %d: %w", i, err)
		}

		itemParams := postgresql.InsertOrderItemParams{
			ID:           itemID,
			OrderID:      orderID,
			ProductID:    productUUID,
			ProductName:  item.ProductName,
			ProductPrice: item.ProductPrice,
			ProductOptions: pqtype.NullRawMessage{
				RawMessage: rawOptions,
				Valid:      true,
			},
			Quantity: item.Quantity,
		}

		s.logger.Info("InsertOrder: Inserting order item",
			"item_id", itemID.String(),
			"order_id", orderID.String(),
			"product_id", item.ProductId)

		err = qtx.InsertOrderItem(ctx, itemParams)
		if err != nil {
			s.logger.Error("InsertOrder: Failed to insert order item",
				"item_index", i,
				"item_id", itemID.String(),
				"product_id", item.ProductId,
				"order_id", orderID.String(),
				"error", err)
			return nil, fmt.Errorf("failed to insert order item %v: %w", item.ProductId, err)
		}
		s.logger.Info("InsertOrder: Order item inserted successfully",
			"item_id", itemID.String(),
			"product_id", item.ProductId)
	}

	s.logger.Info("InsertOrder: Order creation completed successfully",
		"order_id", orderID.String(),
		"user_id", req.UserId,
		"items_processed", len(req.Items))

	return &pb.InsertOrderResponse{Id: orderID.String()}, nil
}

func (s *OrderController) GetAllOrders(ctx context.Context, req *pb.GetAllOrdersRequest) (*pb.GetAllOrdersResponse, error) {
	s.logger.Info("GetAllOrders: Starting to fetch all orders")

	querier := postgresql.New(s.pg.GetDB())

	orders, err := querier.GetAllOrders(ctx)
	if err != nil {
		s.logger.Error("GetAllOrders: Failed to fetch orders from database", "error", err)
		return nil, fmt.Errorf("failed to fetch orders: %w", err)
	}

	s.logger.Info("GetAllOrders: Fetched orders from database", "orders_count", len(orders))

	var respOrders []*pb.Order
	for i, o := range orders {
		s.logger.Info("GetAllOrders: Processing order",
			"order_index", i,
			"order_id", o.ID.String(),
			"user_id", o.UserID.String(),
			"status", o.Status)

		items, err := querier.GetOrderItems(ctx, o.ID)
		if err != nil {
			s.logger.Error("GetAllOrders: Failed to fetch order items",
				"order_id", o.ID.String(),
				"error", err)
			return nil, fmt.Errorf("failed to fetch order items for order %s: %w", o.ID.String(), err)
		}

		s.logger.Info("GetAllOrders: Fetched order items",
			"order_id", o.ID.String(),
			"items_count", len(items))

		var pbItems []*pb.OrderItem
		for j, it := range items {
			s.logger.Info("GetAllOrders: Processing order item",
				"order_id", o.ID.String(),
				"item_index", j,
				"item_id", it.ID.String(),
				"product_id", it.ProductID.String())

			pbItems = append(pbItems, &pb.OrderItem{
				Id:           it.ID.String(),
				OrderId:      it.OrderID.String(),
				ProductId:    it.ProductID.String(),
				ProductName:  it.ProductName,
				ProductPrice: it.ProductPrice,
				Quantity:     it.Quantity,
			})
		}

		respOrders = append(respOrders, &pb.Order{
			Id:              o.ID.String(),
			UserId:          o.UserID.String(),
			OrderNumber:     o.OrderNumber,
			Status:          o.Status,
			TotalPrice:      o.TotalPrice,
			Quantity:        o.Quantity,
			PaymentMethod:   o.PaymentMethod,
			ShippingFee:     o.ShippingFee,
			ShippingAddress: o.ShippingAddress,
			OrderedAt:       o.OrderedAt.Format(time.RFC3339),
			PaidAt:          o.PaidAt.Time.Format(time.RFC3339),
			Memo:            o.Memo.String,
			Items:           pbItems,
		})
	}

	s.logger.Info("GetAllOrders: Successfully processed all orders",
		"total_orders", len(respOrders))

	return &pb.GetAllOrdersResponse{Orders: respOrders}, nil
}

// kafka 메시지를 받았을때 order의 status를 변경하는 함수
func (s *OrderController) UpdateOrderStatus(ctx context.Context, orderID string, status OrderStatus) error {
	s.logger.Info("UpdateOrderStatus: Starting order status update",
		"order_id", orderID,
		"new_status", status)

	db := s.pg.GetDB()
	if db == nil {
		s.logger.Error("UpdateOrderStatus: Failed to get database connection")
		return fmt.Errorf("database connection is nil")
	}

	querier := postgresql.New(db)

	tx, err := db.Begin()
	if err != nil {
		s.logger.Error("UpdateOrderStatus: Failed to begin transaction",
			"order_id", orderID,
			"error", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	qtx := querier.WithTx(tx)
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				s.logger.Error("UpdateOrderStatus: Failed to rollback transaction",
					"order_id", orderID,
					"error", rollbackErr)
			} else {
				s.logger.Info("UpdateOrderStatus: Transaction rolled back successfully",
					"order_id", orderID)
			}
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				s.logger.Error("UpdateOrderStatus: Failed to commit transaction",
					"order_id", orderID,
					"error", commitErr)
				err = commitErr
			} else {
				s.logger.Info("UpdateOrderStatus: Transaction committed successfully",
					"order_id", orderID)
			}
		}
	}()

	orderUUID, err := uuid.Parse(orderID)
	if err != nil {
		s.logger.Error("UpdateOrderStatus: Invalid order ID format",
			"order_id", orderID,
			"error", err)
		return fmt.Errorf("invalid order ID: %w", err)
	}

	// 주문 상태 업데이트
	s.logger.Info("UpdateOrderStatus: Updating order status in database",
		"order_id", orderID,
		"status", status)

	err = qtx.UpdateOrderStatus(ctx, postgresql.UpdateOrderStatusParams{
		ID:     orderUUID,
		Status: string(status),
	})
	if err != nil {
		s.logger.Error("UpdateOrderStatus: Failed to update order status",
			"order_id", orderID,
			"status", status,
			"error", err)
		return fmt.Errorf("failed to update order status: %w", err)
	}

	s.logger.Info("UpdateOrderStatus: Order status updated successfully",
		"order_id", orderID,
		"new_status", status)

	return nil
}

func parseNullTime(s string) sql.NullTime {
	if s == "" {
		return sql.NullTime{Valid: false}
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Valid: true, Time: t}
}

func parseNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{Valid: true, String: s}
}
