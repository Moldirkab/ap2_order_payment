package repository

import (
	"context"
	"payment-service/internal/domain/model"
)

type PaymentRepository interface {
	Create(payment *model.Payment) error
	GetByOrderID(orderID string) (*model.Payment, error)
	GetStats(ctx context.Context) (*model.PaymentStats, error)
}
