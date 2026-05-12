package usecase

import (
	"encoding/json"
	"errors"
	"time"

	"order-service/internal/domain/model"
	"order-service/internal/repository"

	"github.com/google/uuid"
)

var (
	ErrIdempotencyRequired       = errors.New("idempotency key required")
	ErrInvalidAmount             = errors.New("amount must be > 0")
	ErrCustomerEmailRequired     = errors.New("customer email required")
	ErrPaymentServiceUnavailable = errors.New("payment service unavailable")
)

type Cache interface {
	Get(key string) (string, error)
	Set(key string, value string, ttl int) error
	Delete(key string) error
}

type OrderUsecase struct {
	repo    repository.OrderRepository
	payment PaymentClient
	cache   Cache
}

func NewOrderUsecase(r repository.OrderRepository, p PaymentClient, c Cache) *OrderUsecase {
	return &OrderUsecase{
		repo:    r,
		payment: p,
		cache:   c,
	}
}

func (uc *OrderUsecase) CreateOrder(customerID, customerEmail, item string, amount int64, idempotencyKey string) (*model.Order, bool, error) {
	if idempotencyKey == "" {
		return nil, false, ErrIdempotencyRequired
	}
	if customerEmail == "" {
		return nil, false, ErrCustomerEmailRequired
	}
	if amount <= 0 {
		return nil, false, ErrInvalidAmount
	}

	existing, err := uc.repo.FindByIdempotencyKey(idempotencyKey)
	if err != nil {
		return nil, false, err
	}
	if existing != nil {
		return existing, false, nil
	}

	order := &model.Order{
		ID:             uuid.New().String(),
		CustomerID:     customerID,
		CustomerEmail:  customerEmail,
		ItemName:       item,
		Amount:         amount,
		Status:         "Pending",
		CreatedAt:      time.Now(),
		IdempotencyKey: idempotencyKey,
	}

	if err := uc.repo.Create(order); err != nil {
		if existing, _ := uc.repo.FindByIdempotencyKey(idempotencyKey); existing != nil {
			return existing, false, nil
		}
		return nil, false, err
	}

	time.Sleep(2 * time.Second)

	status, err := uc.payment.ProcessPayment(order.ID, order.Amount, order.CustomerEmail)
	if err != nil {
		_ = uc.repo.UpdateStatus(order.ID, "Failed")
		order.Status = "Failed"

		uc.cache.Delete("order:" + order.ID)

		return order, true, ErrPaymentServiceUnavailable
	}

	if status == "Authorized" {
		_ = uc.repo.UpdateStatus(order.ID, "Paid")
		order.Status = "Paid"
	} else {
		_ = uc.repo.UpdateStatus(order.ID, "Failed")
		order.Status = "Failed"
	}

	uc.cache.Delete("order:" + order.ID)

	return order, true, nil
}
func (uc *OrderUsecase) GetOrder(id string) (*model.Order, error) {
	key := "order:" + id

	cached, err := uc.cache.Get(key)
	if err == nil {
		var order model.Order
		if json.Unmarshal([]byte(cached), &order) == nil {
			return &order, nil
		}
	}

	order, err := uc.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	data, _ := json.Marshal(order)
	_ = uc.cache.Set(key, string(data), 300)

	return order, nil
}
func (uc *OrderUsecase) CancelOrder(id string) error {
	order, err := uc.repo.GetByID(id)
	if err != nil {
		return err
	}

	if order.Status != "Pending" {
		return errors.New("only pending orders can be cancelled")
	}

	err = uc.repo.UpdateStatus(id, "Cancelled")
	if err != nil {
		return err
	}

	uc.cache.Delete("order:" + id)

	return nil
}
