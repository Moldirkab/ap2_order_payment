package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"payment-service/internal/domain/model"
	"payment-service/internal/messaging"
	"payment-service/internal/repository"
)

type paymentUsecase struct {
	repo      repository.PaymentRepository
	publisher messaging.EventPublisher
}

func NewPaymentUsecase(repo repository.PaymentRepository, publisher messaging.EventPublisher) PaymentService {
	return &paymentUsecase{
		repo:      repo,
		publisher: publisher,
	}
}

func (u *paymentUsecase) ProcessPayment(orderID string, amount int64, customerEmail string) (*model.Payment, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be > 0")
	}
	if customerEmail == "" {
		return nil, errors.New("customer email required")
	}

	status := "Authorized"
	if amount > 100000 {
		status = "Declined"
	}

	payment := &model.Payment{
		OrderID:       orderID,
		CustomerEmail: customerEmail,
		Amount:        amount,
		Status:        status,
	}

	if err := u.repo.Create(payment); err != nil {
		return nil, err
	}

	if status == "Authorized" && u.publisher != nil {
		event := messaging.PaymentCompletedEvent{
			EventID:       uuid.NewString(),
			OrderID:       orderID,
			Amount:        amount,
			CustomerEmail: customerEmail,
			Status:        "completed",
		}

		if err := u.publisher.PublishPaymentCompleted(context.Background(), event); err != nil {
			return nil, err
		}
	}

	return payment, nil
}

func (u *paymentUsecase) GetPayment(orderID string) (*model.Payment, error) {
	return u.repo.GetByOrderID(orderID)
}

func (u *paymentUsecase) GetPaymentStats(ctx context.Context) (*model.PaymentStats, error) {
	return u.repo.GetStats(ctx)
}
