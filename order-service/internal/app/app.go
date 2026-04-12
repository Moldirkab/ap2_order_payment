package app

import (
	"database/sql"
	"order-service/internal/repository/postgres"
	grpcTransport "order-service/internal/transport/grpc"
	httpTransport "order-service/internal/transport/http"
	"order-service/internal/usecase"
)

type App struct {
	HTTPHandler         *httpTransport.Handler
	OrderTrackingServer *grpcTransport.OrderTrackingServer
	PaymentClient       *grpcTransport.PaymentGRPCClient
}

func NewApp(db *sql.DB, paymentGRPCAddr string) (*App, error) {
	orderRepo := postgres.NewOrderRepository(db)

	paymentClient, err := grpcTransport.NewPaymentGRPCClient(paymentGRPCAddr)
	if err != nil {
		return nil, err
	}

	orderUC := usecase.NewOrderUsecase(orderRepo, paymentClient)
	httpHandler := httpTransport.NewHandler(orderUC)
	orderTrackingServer := grpcTransport.NewOrderTrackingServer(db)

	return &App{
		HTTPHandler:         httpHandler,
		OrderTrackingServer: orderTrackingServer,
		PaymentClient:       paymentClient,
	}, nil
}
