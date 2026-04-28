package model

type Payment struct {
	ID            string
	OrderID       string
	CustomerEmail string
	TransactionID string
	Amount        int64
	Status        string
}

type PaymentStats struct {
	TotalPayments    int64
	SuccessfulCounts int64
	FailedCounts     int64
	TotalAmount      int64
}
