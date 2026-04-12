package repository

import "order-service/internal/domain/model"

type OrderRepository interface {
	Create(order *model.Order) error
	GetByID(id string) (*model.Order, error)
	UpdateStatus(id string, status string) error
	FindByIdempotencyKey(key string) (*model.Order, error)
}
