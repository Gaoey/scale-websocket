package socket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Gaoey/scale-websocket/internal/auth"
	"github.com/labstack/echo/v4"

	"github.com/coder/websocket"
)

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

// Message represents a WebSocket message
type Message struct {
	Type    string      `json:"type"`
	Content interface{} `json:"content"`
}

// WebSocketHandler handles WebSocket connections with Echo
func WebSocketHandler(c echo.Context) error {
	// Verify JWT token from query parameter or Authorization header
	token := c.QueryParam("token")
	if token == "" {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return echo.NewHTTPError(401, "Unauthorized: Missing token")
		}
		token = strings.TrimPrefix(authHeader, "Bearer ")
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

	conn, err := websocket.Accept(c.Response().Writer, r, nil)
	if err != nil {
		log.Println("Accept error:", err)
		return echo.NewHTTPError(1000, "Cannot connect to WebSocket")
	}
	defer conn.Close(websocket.StatusInternalError, "websocket closed unexpectedly")

	// Handle the authenticated WebSocket connection in a goroutine
	go handleWebSocketConnection(conn, claims)

	return nil
}

// handleWebSocketConnection manages an individual WebSocket connection
func handleWebSocketConnection(conn *websocket.Conn, claims *auth.Claims) {
	defer conn.Close(websocket.StatusNormalClosure, "Connection closed")

	// Send welcome message
	welcomeMsg := Message{
		Type: "welcome",
		Content: map[string]string{
			"message":  "Successfully connected to secure WebSocket",
			"username": claims.Username,
			"userID":   claims.UserID,
		},
	}

	result, e := json.Marshal(welcomeMsg)
	if e != nil {
		log.Printf("cannot read msg data: %v", e)
		panic(e)
	}

	err := conn.Write(context.Background(), websocket.MessageText, result)
	if err != nil {
		log.Printf("Error sending welcome message: %v", err)
		return
	}

	// Set read timeout and message size limit
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Message handling loop
	for {
		msgType, data, err := conn.Read(ctx)
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
		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("Invalid message format: %v", err)
			errorMsg := Message{
				Type:    "error",
				Content: "Invalid message format",
			}

			result, err := json.Marshal(errorMsg)
			if err != nil {
				log.Printf("cannot read msg data: %v", err)
				continue
			}

			if err := conn.Write(ctx, websocket.MessageText, result); err != nil {
				log.Printf("Error sending error message: %v", err)
			}
			continue
		}

		// Process message based on type
		switch msg.Type {
		case "ping":
			response := Message{
				Type:    "pong",
				Content: fmt.Sprintf("Pong at %s", time.Now().Format(time.RFC3339)),
			}

			result, err := json.Marshal(response)
			if err != nil {
				log.Printf("cannot read msg data: %v", err)
				continue
			}

			if err := conn.Write(ctx, websocket.MessageText, result); err != nil {
				log.Printf("Error sending pong: %v", err)
			}
		// Add more message type handlers as needed
		default:
			log.Printf("Received message of type: %s", msg.Type)
			// Echo the message back to the client
			if err := conn.Write(ctx, websocket.MessageText, data); err != nil {
				log.Printf("Error echoing message: %v", err)
			}
		}
	}
}
