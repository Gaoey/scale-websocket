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
	ctx         context.Context
	cancelFunc  context.CancelFunc
}

func NewWSChannel(client *rabbitmq.Client, channelName string, queueName string, routingKeys []string, store *stores.ConnectionStorage) *WSChannel {
	ctx, cancel := context.WithCancel(context.Background())

	return &WSChannel{
		Client:      client,
		ChannelName: channelName,
		QueueName:   queueName,
		RoutingKeys: routingKeys,
		store:       store,
		ctx:         ctx,
		cancelFunc:  cancel,
	}
}

func (ws *WSChannel) StartConsumer() error {
	return ws.Client.StartConsumer(ws.ctx, ws.QueueName, ws.RoutingKeys, ws.MessageHandler)
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

	log.Printf("Broadcasting message to %d connections in channel %s", len(store), ws.ChannelName)

	res := NewSuccessMessage("update", msg)
	// Handle incoming messages from RabbitMQ
	data, err := json.Marshal(res)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return fmt.Errorf("cannot marshaling message")
	}

	// Track connections that need to be removed
	var brokenConnections []string

	for _, c := range store {
		if err := c.Conn.Write(c.Ctx, websocket.MessageText, data); err != nil {
			log.Printf("Failed to send message to client=%s, %v", c.ClientID, err)
			brokenConnections = append(brokenConnections, c.ConnectionID)
			continue
		}
	}

	// Clean up broken connections
	if len(brokenConnections) > 0 {
		log.Printf("Removing %d broken connections", len(brokenConnections))
		for _, connID := range brokenConnections {
			// Get the user ID for this connection
			userID := ws.store.GetUserForConnection(connID)
			if userID != "" {
				ws.store.RemoveByConnID(userID, connID)
			}
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

func (c *WSChannel) Stop() {
	if c.cancelFunc != nil {
		c.cancelFunc()
	}
}
