package messaging

import (
	"encoding/json"
	"log"

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
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   amqp.Queue

	processed map[string]bool
}

func NewConsumer(url, queueName string) (*Consumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	q, err := ch.QueueDeclare(
		queueName,
		true, // durable
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		conn:      conn,
		channel:   ch,
		queue:     q,
		processed: make(map[string]bool),
	}, nil
}

func (c *Consumer) Start() error {
	msgs, err := c.channel.Consume(
		c.queue.Name,
		"",
		false, // manual ACK
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	log.Println("Notification service started...")

	for msg := range msgs {
		var event PaymentCompletedEvent

		if err := json.Unmarshal(msg.Body, &event); err != nil {
			log.Println("Invalid message:", err)
			msg.Nack(false, false)
			continue
		}

		// Idempotency check
		if c.processed[event.EventID] {
			log.Println("Duplicate event ignored:", event.EventID)
			msg.Ack(false)
			continue
		}

		// Simulate email
		log.Printf("[Notification] Sent email to %s for Order #%s. Amount: %d\n",
			event.CustomerEmail,
			event.OrderID,
			event.Amount,
		)

		c.processed[event.EventID] = true

		msg.Ack(false)
	}

	return nil
}

func (c *Consumer) Close() {
	c.channel.Close()
	c.conn.Close()
}
