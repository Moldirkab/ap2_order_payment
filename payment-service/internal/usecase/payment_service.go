package usecase

import (
	"context"
	"payment-service/internal/domain/model"
)

type PaymentService interface {
	ProcessPayment(orderID string, amount int64, customerEmail string) (*model.Payment, error)
	GetPayment(orderID string) (*model.Payment, error)
	GetPaymentStats(ctx context.Context) (*model.PaymentStats, error)
}
