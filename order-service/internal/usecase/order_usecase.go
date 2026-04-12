package usecase

import (
	"errors"
	"time"

	"order-service/internal/domain/model"
	"order-service/internal/repository"

	"github.com/google/uuid"
)

var (
	ErrIdempotencyRequired       = errors.New("idempotency key required")
	ErrInvalidAmount             = errors.New("amount must be > 0")
	ErrPaymentServiceUnavailable = errors.New("payment service unavailable")
)

type OrderUsecase struct {
	repo    repository.OrderRepository
	payment PaymentClient
}

func NewOrderUsecase(r repository.OrderRepository, p PaymentClient) *OrderUsecase {
	return &OrderUsecase{repo: r, payment: p}
}

func (uc *OrderUsecase) CreateOrder(customerID, item string, amount int64, idempotencyKey string) (*model.Order, bool, error) {
	if idempotencyKey == "" {
		return nil, false, ErrIdempotencyRequired
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

	status, err := uc.payment.ProcessPayment(order.ID, order.Amount)
	if err != nil {
		_ = uc.repo.UpdateStatus(order.ID, "Failed")
		order.Status = "Failed"
		return order, true, ErrPaymentServiceUnavailable
	}

	if status == "Authorized" {
		_ = uc.repo.UpdateStatus(order.ID, "Paid")
		order.Status = "Paid"
	} else {
		_ = uc.repo.UpdateStatus(order.ID, "Failed")
		order.Status = "Failed"
	}

	return order, true, nil
}

func (uc *OrderUsecase) GetOrder(id string) (*model.Order, error) {
	return uc.repo.GetByID(id)
}

func (uc *OrderUsecase) CancelOrder(id string) error {
	order, err := uc.repo.GetByID(id)
	if err != nil {
		return err
	}

	if order.Status != "Pending" {
		return errors.New("only pending orders can be cancelled")
	}

	return uc.repo.UpdateStatus(id, "Cancelled")
}
