package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/escape-ship/ordersrv/internal/infra/sqlc/postgresql"
	pb "github.com/escape-ship/ordersrv/proto/gen"
	"github.com/google/uuid"
)

type OrderController struct {
	pb.UnimplementedOrderServiceServer
	Queries *postgresql.Queries
	DB      *sql.DB
}

func NewOrderController(db *sql.DB, q *postgresql.Queries) *OrderController {
	return &OrderController{DB: db, Queries: q}
}

func (s *OrderController) InsertOrder(ctx context.Context, req *pb.InsertOrderRequest) (*pb.InsertOrderResponse, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	qtx := s.Queries.WithTx(tx)
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
		UserID:          uuid.MustParse(req.UserId),
		OrderNumber:     req.OrderNumber,
		Status:          req.Status,
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
	orders, err := s.Queries.GetAllOrders(ctx)
	if err != nil {
		return nil, err
	}
	var respOrders []*pb.Order
	for _, o := range orders {
		items, err := s.Queries.GetOrderItems(ctx, o.ID)
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
	return &pb.GetAllOrdersResponse{Orders: respOrders}, nil
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
