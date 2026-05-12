package app

import (
	"database/sql"
	"order-service/internal/cache"
	"order-service/internal/repository/postgres"
	grpcTransport "order-service/internal/transport/grpc"
	httpTransport "order-service/internal/transport/http"
	"order-service/internal/usecase"

	"github.com/redis/go-redis/v9"
)

type App struct {
	HTTPHandler         *httpTransport.Handler
	OrderTrackingServer *grpcTransport.OrderTrackingServer
	PaymentClient       *grpcTransport.PaymentGRPCClient
}

func NewApp(db *sql.DB, paymentGRPCAddr string, redisClient *redis.Client) (*App, error) {
	orderRepo := postgres.NewOrderRepository(db)

	paymentClient, err := grpcTransport.NewPaymentGRPCClient(paymentGRPCAddr)
	if err != nil {
		return nil, err
	}

	redisCache := cache.NewRedisCache(redisClient)

	orderUC := usecase.NewOrderUsecase(orderRepo, paymentClient, redisCache)

	httpHandler := httpTransport.NewHandler(orderUC)
	orderTrackingServer := grpcTransport.NewOrderTrackingServer(db)

	return &App{
		HTTPHandler:         httpHandler,
		OrderTrackingServer: orderTrackingServer,
		PaymentClient:       paymentClient,
	}, nil
}
