package ws

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Gaoey/scale-websocket/internal/repository/rabbitmq"
	"github.com/coder/websocket"
)

// WebSocketConsumer consumes messages from RabbitMQ and forwards them to WebSocket
func WebSocketConsumer(client *rabbitmq.Client, topic []string, wsConn *websocket.Conn) error {
	ctx, cancel := context.WithCancel(context.Background())

	// Handler function for incoming RabbitMQ messages
	handler := func(msg rabbitmq.Message) error {
		// Convert message to JSON
		data, err := json.Marshal(msg)
		if err != nil {
			return err
		}

		// Send to WebSocket
		return wsConn.Write(ctx, websocket.MessageText, data)
	}

	// Start consuming messages
	err := client.Consume(ctx, "", topic, handler)
	if err != nil {
		cancel()
		return err
	}

	// The consumer will continue running until context is cancelled
	go func() {
		wsConn.CloseRead(ctx)
		log.Printf("WebSocket connection closed for user  stopping consumer")
		cancel()
	}()

	return nil
}
