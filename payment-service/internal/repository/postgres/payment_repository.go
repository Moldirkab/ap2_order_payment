package postgres

import (
	"context"
	"database/sql"
	"payment-service/internal/domain/model"
	"payment-service/internal/repository"

	"github.com/google/uuid"
)

type paymentRepo struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) repository.PaymentRepository {
	return &paymentRepo{db: db}
}

func (r *paymentRepo) Create(payment *model.Payment) error {
	if payment.ID == "" {
		payment.ID = uuid.New().String()
	}
	if payment.TransactionID == "" {
		payment.TransactionID = uuid.New().String()
	}

	_, err := r.db.Exec(
		`INSERT INTO payments (id, order_id, customer_email, transaction_id, amount, status)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		payment.ID,
		payment.OrderID,
		payment.CustomerEmail,
		payment.TransactionID,
		payment.Amount,
		payment.Status,
	)
	return err
}

func (r *paymentRepo) GetByOrderID(orderID string) (*model.Payment, error) {
	row := r.db.QueryRow(`
		SELECT id, order_id, customer_email, transaction_id, amount, status
		FROM payments
		WHERE order_id=$1
	`, orderID)

	var p model.Payment
	err := row.Scan(
		&p.ID,
		&p.OrderID,
		&p.CustomerEmail,
		&p.TransactionID,
		&p.Amount,
		&p.Status,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *paymentRepo) GetStats(ctx context.Context) (*model.PaymentStats, error) {
	query := `
		SELECT
			COUNT(*) AS total_payments,
			COUNT(*) FILTER (WHERE status = 'Authorized') AS successful_counts,
			COUNT(*) FILTER (WHERE status = 'Declined') AS failed_counts,
			COALESCE(SUM(amount) FILTER (WHERE status = 'Authorized'), 0) AS total_amount
		FROM payments
	`

	var stats model.PaymentStats

	err := r.db.QueryRowContext(ctx, query).Scan(
		&stats.TotalPayments,
		&stats.SuccessfulCounts,
		&stats.FailedCounts,
		&stats.TotalAmount,
	)
	if err != nil {
		return nil, err
	}

	return &stats, nil
}
