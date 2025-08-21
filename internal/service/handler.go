package service

import (
	"context"
	"log"

	"github.com/escape-ship/ordersrv/internal/infra/sqlc/postgresql"
	"github.com/escape-ship/ordersrv/pkg/postgres"
	"github.com/google/uuid"
)

func PaymentSucceededHandler(ctx context.Context, key, value []byte, pg postgres.DBEngine) error {
	log.Printf("Processing payment succeeded message: key=%s, value=%s", string(key), string(value))
	db := pg.GetDB()
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
	request := postgresql.UpdateOrderStatusParams{
		ID:     uuid.MustParse(string(key)),
		Status: string(OrderStatePaid),
	}
	err = qtx.UpdateOrderStatus(ctx, request)
	if err != nil {
		log.Printf("failed to update order status: %v", err)
		return err
	}

	log.Printf("Order status updated successfully for order ID: %s", string(key))
	return nil
}
