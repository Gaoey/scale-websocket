package rabbitmq

import (
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

const (
	ExchangeDirect  = amqp.ExchangeDirect
	ExchangeFanout  = amqp.ExchangeFanout
	ExchangeTopic   = amqp.ExchangeTopic
	ExchangeHeaders = amqp.ExchangeHeaders
)

// Client represents a RabbitMQ client connection
type Client struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	exchangeName string
	url          string
}

// Message represents the structure of messages passed through RabbitMQ
type Message interface{}

// Config holds the configuration for RabbitMQ connection
type Config struct {
	URL          string
	ExchangeName string
	ExchangeType string // "direct", "topic", "fanout", etc.
}

// NewClient creates a new RabbitMQ client
func NewClient(cfg Config) (*Client, error) {
	if cfg.ExchangeName == "" {
		return nil, fmt.Errorf("exchange name required")
	}

	if cfg.ExchangeType == "" {
		cfg.ExchangeType = ExchangeTopic
	}

	client := &Client{
		url:          cfg.URL,
		exchangeName: cfg.ExchangeName,
	}

	// Connect to RabbitMQ
	var err error
	client.conn, err = amqp.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// Create a channel
	client.channel, err = client.conn.Channel()
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	// Declare the exchange
	err = client.channel.ExchangeDeclare(
		cfg.ExchangeName, // exchange name
		cfg.ExchangeType, // exchange type
		true,             // durable
		false,            // auto-deleted
		false,            // internal
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to declare an exchange: %w", err)
	}

	return client, nil
}

// Close closes the RabbitMQ connection and channel
func (c *Client) Close() error {
	var err error
	if c.channel != nil {
		err = c.channel.Close()
		c.channel = nil
	}
	if c.conn != nil {
		if closeErr := c.conn.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
		c.conn = nil
	}
	return err
}

// Reconnect attempts to reconnect to RabbitMQ with exponential backoff
func (c *Client) Reconnect(maxRetries int) error {
	var err error
	retries := 0
	backoff := 1 * time.Second

	for retries < maxRetries {
		log.Printf("Attempting to reconnect to RabbitMQ (attempt %d/%d)", retries+1, maxRetries)

		if c.conn, err = amqp.Dial(c.url); err == nil {
			if c.channel, err = c.conn.Channel(); err == nil {
				// Re-declare exchange
				err = c.channel.ExchangeDeclare(
					c.exchangeName, // exchange name
					"topic",        // exchange type
					true,           // durable
					false,          // auto-deleted
					false,          // internal
					false,          // no-wait
					nil,            // arguments
				)
				if err == nil {
					log.Println("Successfully reconnected to RabbitMQ")
					return nil
				}
			}
			// Close connection if channel creation failed
			c.conn.Close()
		}

		log.Printf("Failed to reconnect: %v. Retrying in %v", err, backoff)
		time.Sleep(backoff)

		// Exponential backoff with a cap
		backoff *= 2
		if backoff > 30*time.Second {
			backoff = 30 * time.Second
		}
		retries++
	}

	return fmt.Errorf("failed to reconnect to RabbitMQ after %d attempts: %w", maxRetries, err)
}
