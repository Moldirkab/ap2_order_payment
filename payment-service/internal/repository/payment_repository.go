package repository

import "payment-service/internal/domain/model"

type PaymentRepository interface {
	Create(payment *model.Payment) error
	GetByOrderID(orderID string) (*model.Payment, error)
}
