package usecase

type PaymentClient interface {
	ProcessPayment(orderID string, amount int64, customerEmail string) (string, error)
}
