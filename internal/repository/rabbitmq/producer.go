package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

// Publish publishes a message to a specified routing key
func (c *Client) Publish(ctx context.Context, routingKey string, msg Message) error {
	// Context handling for cancellation/timeout
	if ctx.Err() != nil {
		return ctx.Err()
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	log.Printf("Publishing message to routing key %s: %s", routingKey, body)

	err = c.channel.Publish(
		c.ExchangeName, // exchange
		routingKey,     // routing key
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		})
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}
