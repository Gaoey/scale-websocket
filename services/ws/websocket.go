package ws

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/Gaoey/scale-websocket/internal/stores"
	"github.com/Gaoey/scale-websocket/services/auth"
	"github.com/coder/websocket"
)

var (
	PingEvent      = "ping"
	AuthEvent      = "auth"
	SubscribeEvent = "subscribe"
)

var (
	OrderUpdateChannel = "order_update"
)

type AuthWebSocket struct {
	Conn   *websocket.Conn
	Claims *auth.Claims
	Store  *stores.ConnectionStorage
}

func NewAuthWebSocket(conn *websocket.Conn, claims *auth.Claims, store *stores.ConnectionStorage) *AuthWebSocket {
	connId := stores.GenerateConnectionID()
	store.Add(claims.UserID, connId, conn, true)

	return &AuthWebSocket{
		Conn:   conn,
		Claims: claims,
		Store:  store,
	}
}

func (ws AuthWebSocket) EventHandler(ctx context.Context) {
	// Message handling loop
	for {
		msgType, data, err := ws.Conn.Read(ctx)
		if err != nil {
			if websocket.CloseStatus(err) != websocket.StatusNormalClosure {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		// Only process text messages
		if msgType != websocket.MessageText {
			continue
		}

		// Parse message
		msg, err := ValidateMessage(ctx, data)
		if err != nil {
			ws.SendMessage(ctx, msg)
			continue
		}

		// Process message based on type
		switch msg.Event {
		case PingEvent:
			response := NewSuccessMessage("pong", map[string]interface{}{
				"timestamp": time.Now().Unix(),
			})
			ws.SendMessage(ctx, response)

		case SubscribeEvent:
			if msg.Channel == "" {
				response := NewErrorMessage("subscribe", "1002", "Channel name is required")
				ws.SendMessage(ctx, response)
			}
			ws.Store.AddChannel(ws.Claims.UserID, msg.Channel)

		default:
			log.Printf("Received message of type: %s", msg.Event)
		}
	}
}

func (ws AuthWebSocket) SendMessage(ctx context.Context, msg Message) error {
	result, err := json.Marshal(msg)
	if err != nil {
		log.Printf("cannot read msg data: %v", err)
		return err
	}

	if err := ws.Conn.Write(ctx, websocket.MessageText, result); err != nil {
		log.Printf("Error sending message: %v", err)
	}
	return nil
}

func ValidateMessage(ctx context.Context, data []byte) (Message, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("Invalid message format: %v", err)
		errorMsg := NewErrorMessage("auth", "1001", "Invalid message format")
		return errorMsg, err
	}

	return msg, nil
}
