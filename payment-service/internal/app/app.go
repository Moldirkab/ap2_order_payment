package app

import (
	"database/sql"
	"log"
	"os"

	"payment-service/internal/messaging"
	"payment-service/internal/repository/postgres"
	grpcTransport "payment-service/internal/transport/grpc"
	"payment-service/internal/usecase"
)

type App struct {
	GRPCServer *grpcTransport.PaymentServer
	Publisher  messaging.EventPublisher
}

func NewApp(db *sql.DB) *App {
	repo := postgres.NewPaymentRepository(db)

	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	queueName := os.Getenv("PAYMENT_EVENTS_QUEUE")
	if queueName == "" {
		queueName = "payment.completed"
	}

	publisher, err := messaging.NewRabbitMQPublisher(rabbitURL, queueName)
	if err != nil {
		log.Fatal(err)
	}

	uc := usecase.NewPaymentUsecase(repo, publisher)
	grpcServer := grpcTransport.NewPaymentServer(uc)

	return &App{
		GRPCServer: grpcServer,
		Publisher:  publisher,
	}
}

func (a *App) Close() error {
	if a.Publisher != nil {
		return a.Publisher.Close()
	}
	return nil
}
