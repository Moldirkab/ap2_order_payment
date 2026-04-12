package postgres

import (
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
		`INSERT INTO payments (id, order_id, transaction_id, amount, status)
		 VALUES ($1, $2, $3, $4, $5)`,
		payment.ID, payment.OrderID, payment.TransactionID, payment.Amount, payment.Status,
	)
	return err
}

func (r *paymentRepo) GetByOrderID(orderID string) (*model.Payment, error) {
	row := r.db.QueryRow(`SELECT id, order_id, transaction_id, amount, status FROM payments WHERE order_id=$1`, orderID)
	var p model.Payment
	err := row.Scan(&p.ID, &p.OrderID, &p.TransactionID, &p.Amount, &p.Status)
	if err != nil {
		return nil, err
	}
	return &p, nil
}
