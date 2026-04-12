package usecase

import "order-service/internal/domain/model"

type OrderService interface {
	CreateOrder(customerID, item string, amount int64, idempotencyKey string) (*model.Order, bool, error)
	GetOrder(id string) (*model.Order, error)
	CancelOrder(id string) error
}
