package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/escape-ship/ordersrv/internal/infra/sqlc/postgresql"
	"github.com/escape-ship/ordersrv/pkg/kafka"
	"github.com/escape-ship/ordersrv/pkg/postgres"
	pb "github.com/escape-ship/ordersrv/proto/gen"
	"github.com/google/uuid"
)

type OrderController struct {
	pb.UnimplementedOrderServiceServer
	pg    postgres.DBEngine
	kafka kafka.Engine
}

func NewOrderController(pg postgres.DBEngine, kafkaEngine kafka.Engine) *OrderController {
	return &OrderController{
		pg:    pg,
		kafka: kafkaEngine,
	}
}

func (s *OrderController) InsertOrder(ctx context.Context, req *pb.InsertOrderRequest) (*pb.InsertOrderResponse, error) {
	db := s.pg.GetDB()
	querier := postgresql.New(db)

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	qtx := querier.WithTx(tx)
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	orderID := uuid.New()
	orderParams := postgresql.InsertOrderParams{
		ID:              orderID,
		UserID:          req.UserId,
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
	_, err = qtx.InsertOrder(ctx, orderParams)
	if err != nil {
		return nil, err
	}
	for _, item := range req.Items {
		itemID := uuid.New()
		itemParams := postgresql.InsertOrderItemParams{
			ID:           itemID,
			OrderID:      orderID,
			ProductID:    uuid.MustParse(item.ProductId),
			ProductName:  item.ProductName,
			ProductPrice: item.ProductPrice,
			Quantity:     item.Quantity,
		}
		err = qtx.InsertOrderItem(ctx, itemParams)
		if err != nil {
			return nil, err
		}
	}
	return &pb.InsertOrderResponse{Id: orderID.String()}, nil
}

func (s *OrderController) GetAllOrders(ctx context.Context, req *pb.GetAllOrdersRequest) (*pb.GetAllOrdersResponse, error) {
	querier := postgresql.New(s.pg.GetDB())

	orders, err := querier.GetAllOrders(ctx)
	if err != nil {
		return nil, err
	}
	var respOrders []*pb.Order
	for _, o := range orders {
		items, err := querier.GetOrderItems(ctx, o.ID)
		if err != nil {
			return nil, err
		}
		var pbItems []*pb.OrderItem
		for _, it := range items {
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
			UserId:          o.UserID,
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
	return &pb.GetAllOrdersResponse{Orders: respOrders}, nil
}

// kafka 메시지를 받았을때 order의 status를 변경하는 함수
func (s *OrderController) UpdateOrderStatus(ctx context.Context, orderID string, status OrderStatus) error {
	db := s.pg.GetDB()
	querier := postgresql.New(db)

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	qtx := querier.WithTx(tx)
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	orderUUID, err := uuid.Parse(orderID)
	if err != nil {
		return err
	}

	// 상품 id 불러오기
	productId, err := qtx.GetProductIDsByOrderID(ctx, orderUUID)
	if err != nil {
		return err
	}

	// Kafka 메시지 생성
	msgValue, err := json.Marshal(productId)
	if err != nil {
		slog.Error("Failed to marshal product IDs", "error", err)
		return err
	}

	// 주문 상태 업데이트
	err = qtx.UpdateOrderStatus(ctx, postgresql.UpdateOrderStatusParams{
		ID:     orderUUID,
		Status: string(status),
	})
	if err != nil {
		return err
	}

	// Kafka 메시지 전송
	if s.kafka != nil {
		producer := s.kafka.Producer()
		if producer != nil {
			err := producer.Publish(ctx, []byte("inventory-discount"), msgValue)
			if err != nil {
				slog.Error("Failed to publish kafka message", "error", err)
			}
			slog.Info("Published kakao-approve message to Kafka", "order_id", orderID)
		}
	}

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
