package usecase

import (
	"context"
	"errors"
	"payment-service/internal/domain/model"
	"payment-service/internal/repository"
)

type paymentUsecase struct {
	repo repository.PaymentRepository
}

func NewPaymentUsecase(repo repository.PaymentRepository) PaymentService {
	return &paymentUsecase{repo: repo}
}

func (u *paymentUsecase) ProcessPayment(orderID string, amount int64) (*model.Payment, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be > 0")
	}

	status := "Authorized"
	if amount > 100000 {
		status = "Declined"
	}

	payment := &model.Payment{
		OrderID: orderID,
		Amount:  amount,
		Status:  status,
	}

	if err := u.repo.Create(payment); err != nil {
		return nil, err
	}

	return payment, nil
}

func (u *paymentUsecase) GetPayment(orderID string) (*model.Payment, error) {
	return u.repo.GetByOrderID(orderID)
}
func (u *paymentUsecase) GetPaymentStats(ctx context.Context) (*model.PaymentStats, error) {
	return u.repo.GetStats(ctx)
}
