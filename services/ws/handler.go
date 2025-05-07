package ws

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/Gaoey/scale-websocket/services/auth"
	"github.com/labstack/echo/v4"

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

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

type AuthWebSocket struct {
	Conn   *websocket.Conn
	Claims *auth.Claims
}

func NewAuthWebSocket(conn *websocket.Conn, claims *auth.Claims) *AuthWebSocket {
	return &AuthWebSocket{
		Conn:   conn,
		Claims: claims,
	}
}

func AuthWebSocketHandler(c echo.Context) error {
	token := c.QueryParam("token")
	if token == "" {
		return echo.NewHTTPError(401, "Unauthorized: Missing token")
	}

	// Validate token
	claims, err := auth.ValidateToken(token)
	if err != nil {
		return echo.NewHTTPError(401, "Unauthorized: Invalid token")
	}

	// Create a context with the user information
	ctx := context.WithValue(c.Request().Context(), ContextKey("userID"), claims.UserID)
	ctx = context.WithValue(ctx, ContextKey("username"), claims.Username)

	r := c.Request().WithContext(ctx)

	conn, err := websocket.Accept(c.Response().Writer, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		log.Println("Accept error:", err)
		return echo.NewHTTPError(1000, "Cannot connect to WebSocket")
	}

	log.Printf("WebSocket connection established for user: %s", claims.Username)
	defer conn.Close(websocket.StatusNormalClosure, "Connection closed")

	// Set read timeout and message size limit
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ws := NewAuthWebSocket(conn, claims)

	// Send welcome message
	log.Printf("Sending welcome message to user: %s", claims.Username)
	welcomeMsg := NewSuccessMessage("auth", map[string]interface{}{
		"message":   "success",
		"username":  claims.Username,
		"timestamp": time.Now().Unix(),
	})
	if err := ws.SendMessage(ctx, welcomeMsg); err != nil {
		log.Printf("Error sending welcome message: %v", err)
		return err
	}

	ws.EventHandler(ctx)

	return nil
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

		default:
			log.Printf("Received message of type: %s", msg.Event)
			// Echo the message back to the client
			// if err := conn.Write(ctx, websocket.MessageText, data); err != nil {
			// 	log.Printf("Error echoing message: %v", err)
			// }
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
