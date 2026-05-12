package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"notification-service/internal/messaging"

	"github.com/redis/go-redis/v9"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	rabbitURL := os.Getenv("RABBITMQ_URL")
	queue := os.Getenv("PAYMENT_EVENTS_QUEUE")

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	consumer := messaging.NewConsumer(rabbitURL, queue, redisClient)

	if err := consumer.Start(ctx); err != nil {
		log.Fatal(err)
	}
}
