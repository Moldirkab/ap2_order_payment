package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

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

	queueName := os.Getenv("PAYMENT_EVENTS_QUEUE")
	if queueName == "" {
		queueName = "payment.completed"
	}

	dlxName := "payment.dlx"
	dlqName := "payment.completed.dlq"
	dlqRoutingKey := "payment.failed"

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

	err = ch.ExchangeDeclare(
		dlxName,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = ch.QueueDeclare(
		dlqName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	err = ch.QueueBind(
		dlqName,
		dlqRoutingKey,
		dlxName,
		false,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-dead-letter-exchange":    dlxName,
			"x-dead-letter-routing-key": dlqRoutingKey,
		},
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
	attempts := make(map[string]int)
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
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				log.Println("Invalid message:", err)
				msg.Nack(false, false)
				continue
			}

			if event.CustomerEmail == "fail@example.com" {
				mu.Lock()
				attempts[event.EventID]++
				currentAttempt := attempts[event.EventID]
				mu.Unlock()

				log.Printf("[Error] Simulated permanent failure for event %s. Attempt %d/3",
					event.EventID,
					currentAttempt,
				)

				if currentAttempt >= 3 {
					log.Printf("[DLQ] Event %s failed 3 times. Moving to DLQ", event.EventID)
					msg.Nack(false, false)
					continue
				}

				time.Sleep(2 * time.Second)
				msg.Nack(false, true)
				continue
			}

			mu.Lock()
			if processed[event.EventID] {
				mu.Unlock()
				log.Printf("[Duplicate] Event %s already processed", event.EventID)
				msg.Ack(false)
				continue
			}

			processed[event.EventID] = true
			mu.Unlock()

			log.Printf(
				"[Notification] Sent email to %s for Order #%s. Amount: $%.2f",
				event.CustomerEmail,
				event.OrderID,
				float64(event.Amount)/100,
			)

			if err := msg.Ack(false); err != nil {
				log.Println("ACK failed:", err)
			}
		}
	}
}
