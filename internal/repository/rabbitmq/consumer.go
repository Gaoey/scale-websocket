package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
)

// ConsumeFunc is a callback function type for consuming messages
type ConsumeFunc func(Message) error

// Consume starts consuming messages from a queue with the given routing keys
func (c *Client) StartConsumer(ctx context.Context, queueName string, routingKeys []string, handler ConsumeFunc) error {
	// Declare a queue
	q, err := c.channel.QueueDeclare(
		queueName, // queue name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare a queue: %w", err)
	}

	// Bind the queue to the exchange with the specified routing keys
	for _, key := range routingKeys {
		err = c.channel.QueueBind(
			q.Name,         // queue name
			key,            // routing key
			c.ExchangeName, // exchange
			false,          // no-wait
			nil,            // arguments
		)
		if err != nil {
			return fmt.Errorf("failed to bind a queue: %w", err)
		}
	}

	// Set up the consumer
	msgs, err := c.channel.Consume(
		q.Name, // queue
		"",     // consumer tag
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %w", err)
	}

	go func() {
		defer log.Println("RabbitMQ consumer stopped")

		for {
			select {
			case <-ctx.Done():
				return
			case d, ok := <-msgs:
				if !ok {
					log.Println("RabbitMQ channel closed")
					return
				}

				var msg Message
				err := json.Unmarshal(d.Body, &msg)
				if err != nil {
					log.Printf("Error unmarshaling message: %v", err)
					d.Nack(false, false) // Negative acknowledgement, don't requeue
					continue
				}

				err = handler(msg)
				if err != nil {
					log.Printf("Error handling message: %v", err)
					d.Nack(false, true) // Negative acknowledgement, requeue
					continue
				}

				// Acknowledge message
				err = d.Ack(false)
				if err != nil {
					log.Printf("Error acknowledging message: %v", err)
				}
			}
		}
	}()

	return nil
}
