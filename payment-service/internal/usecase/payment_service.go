package usecase

import "payment-service/internal/domain/model"

type PaymentService interface {
	ProcessPayment(orderID string, amount int64) (*model.Payment, error)
	GetPayment(orderID string) (*model.Payment, error)
}
