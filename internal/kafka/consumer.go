package kafka

import (
	"context"
	"log"
)

func PaymentSucceededHandler(ctx context.Context, key, value []byte) error {
	log.Printf("Processing payment succeeded message: key=%s, value=%s", string(key), string(value))
	// Handle the payment succeeded message
	//TODO: Database update
	return nil
}
