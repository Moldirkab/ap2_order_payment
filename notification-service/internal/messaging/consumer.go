package messaging

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"notification-service/internal/provider"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

type PaymentCompletedEvent struct {
	EventID       string `json:"event_id"`
	OrderID       string `json:"order_id"`
	Amount        int64  `json:"amount"`
	CustomerEmail string `json:"customer_email"`
	Status        string `json:"status"`
}

type Consumer struct {
	rabbitURL string
	queueName string
	redis     *redis.Client
	sender    provider.EmailSender
}

func NewConsumer(rabbitURL string, queueName string, redisClient *redis.Client) *Consumer {
	return &Consumer{
		rabbitURL: rabbitURL,
		queueName: queueName,
		redis:     redisClient,
		sender:    provider.NewEmailSender(),
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	conn, err := amqp.Dial(c.rabbitURL)
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	msgs, err := ch.Consume(
		c.queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	log.Println("Notification worker started...")

	for {
		select {
		case <-ctx.Done():
			return nil

		case msg, ok := <-msgs:
			if !ok {
				return nil
			}

			c.handleMessage(ctx, msg)
		}
	}
}

func (c *Consumer) handleMessage(ctx context.Context, msg amqp.Delivery) {
	var event PaymentCompletedEvent

	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Println("invalid message:", err)
		msg.Nack(false, false)
		return
	}

	key := "notification:" + event.EventID

	status, _ := c.redis.Get(ctx, key).Result()

	if status == "done" {
		log.Println("already processed:", event.EventID)
		msg.Ack(false)
		return
	}

	if status == "processing" {
		log.Println("already processing:", event.EventID)
		msg.Nack(false, true)
		return
	}

	c.redis.Set(ctx, key, "processing", 10*time.Minute)

	maxRetries := 3

	for attempt := 1; attempt <= maxRetries; attempt++ {

		err := c.sender.Send(
			event.CustomerEmail,
			"Payment Successful",
			"Your payment was successful",
		)

		if err == nil {
			log.Printf("email sent to %s (order %s)", event.CustomerEmail, event.OrderID)

			c.redis.Set(ctx, key, "done", 24*time.Hour)

			msg.Ack(false)
			return
		}

		log.Printf("attempt %d failed for %s", attempt, event.CustomerEmail)

		backoff := time.Duration(2<<attempt) * time.Second
		time.Sleep(backoff)
	}

	log.Println("FAILED permanently:", event.EventID)

	msg.Nack(false, false)
}
