package messaging

import (
	"context"
	"encoding/json"
	"log"
	"sync"
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

type Consumer struct {
	rabbitURL string
	queueName string
}

func NewConsumer(rabbitURL, queueName string) *Consumer {
	return &Consumer{
		rabbitURL: rabbitURL,
		queueName: queueName,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	dlxName := "payment.dlx"
	dlqName := "payment.completed.dlq"
	dlqRoutingKey := "payment.failed"

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
		return err
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
		return err
	}

	err = ch.QueueBind(
		dlqName,
		dlqRoutingKey,
		dlxName,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	_, err = ch.QueueDeclare(
		c.queueName,
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
		return err
	}

	err = ch.Qos(1, 0, false)
	if err != nil {
		return err
	}

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

	processed := make(map[string]bool)
	attempts := make(map[string]int)
	var mu sync.Mutex

	log.Println("Notification Service started. Waiting for payment.completed events...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Notification Service shutting down...")
			return nil

		case msg, ok := <-msgs:
			if !ok {
				log.Println("Message channel closed")
				return nil
			}

			c.handleMessage(msg, processed, attempts, &mu)
		}
	}
}

func (c *Consumer) handleMessage(
	msg amqp.Delivery,
	processed map[string]bool,
	attempts map[string]int,
	mu *sync.Mutex,
) {
	var event PaymentCompletedEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Println("Invalid message:", err)
		msg.Nack(false, false)
		return
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
			return
		}

		time.Sleep(2 * time.Second)
		msg.Nack(false, true)
		return
	}

	mu.Lock()
	if processed[event.EventID] {
		mu.Unlock()
		log.Printf("[Duplicate] Event %s already processed", event.EventID)
		msg.Ack(false)
		return
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
