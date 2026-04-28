package postgres

import (
	"database/sql"
	"order-service/internal/domain/model"
	"order-service/internal/repository"
)

type orderRepo struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) repository.OrderRepository {
	return &orderRepo{db: db}
}

func (r *orderRepo) Create(order *model.Order) error {
	query := `
	INSERT INTO orders (id, customer_id, customer_email, item_name, amount, status, created_at, idempotency_key)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(query,
		order.ID,
		order.CustomerID,
		order.CustomerEmail,
		order.ItemName,
		order.Amount,
		order.Status,
		order.CreatedAt,
		order.IdempotencyKey,
	)
	return err
}

func (r *orderRepo) GetByID(id string) (*model.Order, error) {
	row := r.db.QueryRow(`
	SELECT id, customer_id, customer_email, item_name, amount, status, created_at, idempotency_key
	FROM orders WHERE id = $1
	`, id)

	var o model.Order
	err := row.Scan(
		&o.ID,
		&o.CustomerID,
		&o.CustomerEmail,
		&o.ItemName,
		&o.Amount,
		&o.Status,
		&o.CreatedAt,
		&o.IdempotencyKey,
	)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *orderRepo) UpdateStatus(id string, status string) error {
	_, err := r.db.Exec(`UPDATE orders SET status=$1 WHERE id=$2`, status, id)
	return err
}

func (r *orderRepo) FindByIdempotencyKey(key string) (*model.Order, error) {
	query := `
	SELECT id, customer_id, customer_email, item_name, amount, status, created_at, idempotency_key
	FROM orders WHERE idempotency_key = $1
	`
	row := r.db.QueryRow(query, key)

	var o model.Order
	err := row.Scan(
		&o.ID,
		&o.CustomerID,
		&o.CustomerEmail,
		&o.ItemName,
		&o.Amount,
		&o.Status,
		&o.CreatedAt,
		&o.IdempotencyKey,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &o, nil
}
