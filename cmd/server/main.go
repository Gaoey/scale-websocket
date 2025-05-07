package main

import (
	"log"

	"github.com/Gaoey/scale-websocket/internal/stores"
	"github.com/Gaoey/scale-websocket/services/routes"
	"github.com/Gaoey/scale-websocket/services/ws"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// redis := os.Getenv("REDIS_URL")
	// rabbitmq := os.Getenv("RABBITMQ_URL")

	// dependency injection
	e := echo.New()
	wsHandler := ws.NewWebSocketHandler(stores.NewConnectionStorage())

	routes.SetupRoutes(e, wsHandler)

	if err := e.Start(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
