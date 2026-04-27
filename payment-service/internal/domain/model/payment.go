package model

type Payment struct {
	ID            string
	OrderID       string
	TransactionID string
	Amount        int64  // Amount in cents
	Status        string // "Authorized", "Declined"
}

type PaymentStats struct {
	TotalPayments    int64
	SuccessfulCounts int64
	FailedCounts     int64
	TotalAmount      int64
}
