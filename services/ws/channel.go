package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/Gaoey/scale-websocket/internal/repository/rabbitmq"
	"github.com/Gaoey/scale-websocket/internal/stores"
	"github.com/coder/websocket"
)

type ChannelQueue struct {
	ChannelName string
	QueueName   string
	RoutingKeys []string
}

var (
	CHANNELS = []ChannelQueue{{
		ChannelName: "order_update",
		QueueName:   "ws_order_queue",
		RoutingKeys: []string{"ws_order.#"},
	}}
)

type WSChannel struct {
	Client      *rabbitmq.Client
	ChannelName string
	QueueName   string
	RoutingKeys []string
	store       *stores.ConnectionStorage
}

func NewWSChannel(client *rabbitmq.Client, channelName string, queueName string, routingKeys []string, store *stores.ConnectionStorage) *WSChannel {
	return &WSChannel{
		Client:      client,
		ChannelName: channelName,
		QueueName:   queueName,
		RoutingKeys: routingKeys,
		store:       store,
	}
}

func (ws *WSChannel) StartConsumer() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure cancel is called when the function returns

	// Start consuming messages
	err := ws.Client.StartConsumer(ctx, ws.QueueName, ws.RoutingKeys, ws.MessageHandler)
	if err != nil {
		return err
	}

	return nil
}

func (ws *WSChannel) MessageHandler(msg rabbitmq.Message) error {
	// transform message to type Message
	store, err := ws.store.GetByChannel(ws.ChannelName)
	if err != nil {
		return err
	}

	if len(store) == 0 {
		log.Printf("No clients connected to channel=%s", ws.ChannelName)
		return fmt.Errorf("no clients connected to channel=%s", ws.ChannelName)
	}

	// Handle incoming messages from RabbitMQ
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	for _, c := range store {
		if err := c.Conn.Write(c.Ctx, websocket.MessageText, data); err != nil {
			log.Printf("Failed to send message to client=%s, %v", c.ClientID, err)
			continue
		}
	}

	return nil
}

func ValidateChannel(msg Message) error {
	if msg.Channel == "" {
		return fmt.Errorf("channel name is required")
	}

	for _, channel := range CHANNELS {
		if msg.Channel == channel.ChannelName {
			return nil
		}
	}

	return fmt.Errorf("invalid channel name: %s", msg.Channel)
}
