package app

import (
	"database/sql"
	"payment-service/internal/repository/postgres"
	grpcTransport "payment-service/internal/transport/grpc"
	"payment-service/internal/usecase"
)

type App struct {
	GRPCServer *grpcTransport.PaymentServer
}

func NewApp(db *sql.DB) *App {
	repo := postgres.NewPaymentRepository(db)
	uc := usecase.NewPaymentUsecase(repo)
	grpcServer := grpcTransport.NewPaymentServer(uc)

	return &App{
		GRPCServer: grpcServer,
	}
}
