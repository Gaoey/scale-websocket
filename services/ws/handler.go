package ws

import (
	"context"
	"log"
	"time"

	"github.com/Gaoey/scale-websocket/internal/stores"
	"github.com/Gaoey/scale-websocket/services/auth"
	"github.com/labstack/echo/v4"

	"github.com/coder/websocket"
)

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

type WebSocketHandler struct {
	store *stores.ConnectionStorage
}

func NewWebSocketHandler(store *stores.ConnectionStorage) *WebSocketHandler {
	return &WebSocketHandler{
		store: store,
	}
}

func (h WebSocketHandler) AuthWebSocketHandler(c echo.Context) error {
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
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
	defer cancel()

	ws := NewAuthWebSocket(ctx, conn, claims, h.store)

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

	ws.AuthEventHandler(ctx)

	return nil
}
