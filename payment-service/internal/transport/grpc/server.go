package grpc

import (
	"context"

	"payment-service/internal/usecase"

	"github.com/Moldirkab/ap2-generated/paymentpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PaymentServer struct {
	paymentpb.UnimplementedPaymentServiceServer
	usecase usecase.PaymentService
}

func NewPaymentServer(uc usecase.PaymentService) *PaymentServer {
	return &PaymentServer{usecase: uc}
}

func (s *PaymentServer) ProcessPayment(ctx context.Context, req *paymentpb.PaymentRequest) (*paymentpb.PaymentResponse, error) {
	if req.GetOrderId() == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
	}

	if req.GetCustomerEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "customer_email is required")
	}

	payment, err := s.usecase.ProcessPayment(
		req.GetOrderId(),
		req.GetAmount(),
		req.GetCustomerEmail(),
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &paymentpb.PaymentResponse{
		PaymentId:     payment.ID,
		OrderId:       payment.OrderID,
		TransactionId: payment.TransactionID,
		Amount:        payment.Amount,
		Status:        payment.Status,
	}, nil
}

func (s *PaymentServer) GetPaymentStats(ctx context.Context, req *paymentpb.GetPaymentStatsRequest) (*paymentpb.PaymentStats, error) {
	stats, err := s.usecase.GetPaymentStats(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &paymentpb.PaymentStats{
		TotalPayments:    stats.TotalPayments,
		SuccessfulCounts: stats.SuccessfulCounts,
		FailedCounts:     stats.FailedCounts,
		TotalAmount:      stats.TotalAmount,
	}, nil
}
