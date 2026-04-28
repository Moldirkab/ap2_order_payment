package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
)

type PaymentCompletedEvent struct {
	EventID       string `json:"event_id"`
	OrderID       string `json:"order_id"`
	Amount        int64  `json:"amount"`
	CustomerEmail string `json:"customer_email"`
	Status        string `json:"status"`
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	queueName := os.Getenv("PAYMENT_EVENTS_QUEUE")
	if queueName == "" {
		queueName = "payment.completed"
	}

	_, err = ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	err = ch.Qos(1, 0, false)
	if err != nil {
		log.Fatal(err)
	}

	msgs, err := ch.Consume(
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	processed := make(map[string]bool)
	var mu sync.Mutex

	log.Println("Notification Service started. Waiting for payment.completed events...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Notification Service shutting down...")
			return

		case msg, ok := <-msgs:
			if !ok {
				log.Println("Message channel closed")
				return
			}

			var event PaymentCompletedEvent
			err := json.Unmarshal(msg.Body, &event)
			if err != nil {
				log.Println("Invalid message:", err)
				msg.Nack(false, false)
				continue
			}

			mu.Lock()
			if processed[event.EventID] {
				mu.Unlock()
				log.Printf("[Duplicate] Event %s already processed\n", event.EventID)
				msg.Ack(false)
				continue
			}

			processed[event.EventID] = true
			mu.Unlock()

			log.Printf(
				"[Notification] Sent email to %s for Order #%s. Amount: $%.2f\n",
				event.CustomerEmail,
				event.OrderID,
				float64(event.Amount)/100,
			)

			err = msg.Ack(false)
			if err != nil {
				log.Println("ACK failed:", err)
			}
		}
	}
}
