package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
)

// ConsumeFunc is a callback function type for consuming messages
type ConsumeFunc func(Message) error

// Consume starts consuming messages from a queue with the given routing keys
func (c *Client) StartConsumer(ctx context.Context, queueName string, routingKeys []string, handler ConsumeFunc) error {
	log.Printf("Starting consumer for queue: %s with routing keys: %v", queueName, routingKeys)

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

	log.Printf("Queue declared: %s", q.Name)

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
		log.Printf("Bound queue %s to exchange %s with routing key %s", q.Name, c.ExchangeName, key)
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

	log.Printf("Registered consumer for queue: %s", q.Name)

	// Use a waitgroup to keep the consumer goroutine running
	var wg sync.WaitGroup
	wg.Add(1)

	// Create a new context that's cancelled when the parent context is
	consumerCtx, cancel := context.WithCancel(ctx)

	go func() {
		defer wg.Done()
		defer log.Println("RabbitMQ consumer stopped for queue:", queueName)
		defer cancel() // Ensure context is cancelled when goroutine exits

		for {
			select {
			case <-consumerCtx.Done():
				log.Printf("Context cancelled for consumer %s", queueName)
				return
			case d, ok := <-msgs:
				if !ok {
					log.Printf("RabbitMQ channel closed for queue %s", queueName)
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
					log.Printf("Error handling message from %s: %v", queueName, err)
					d.Nack(false, false) // Negative acknowledgement, requeue
					continue
				}

				// Acknowledge messageno clients connected
				err = d.Ack(false)
				if err != nil {
					log.Printf("Error acknowledging message from %s: %v", queueName, err)
				} else {
					log.Printf("Successfully processed and acknowledged message from %s", queueName)
				}
			}
		}
	}()

	// This will prevent the consumer from stopping when StartConsumer returns
	go func() {
		// Wait for the consumer to finish
		wg.Wait()
		log.Printf("Consumer for queue %s has fully stopped", queueName)
	}()

	return nil
}
