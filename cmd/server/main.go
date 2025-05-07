package main

import (
	"log"
	"os"

	"github.com/Gaoey/scale-websocket/internal/repository/rabbitmq"
	"github.com/Gaoey/scale-websocket/internal/stores"
	"github.com/Gaoey/scale-websocket/services/example"
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

	// dependency injection
	stores := stores.NewConnectionStorage()
	rabbitmqClient, err := rabbitmq.NewClient(rabbitmq.Config{
		URL:          os.Getenv("RABBITMQ_URL"),
		ExchangeName: "websocket_events",
		ExchangeType: "fanout",
	})
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ client: %v", err)
	}
	e := echo.New()

	exampleHandler := example.NewExampleHandler(rabbitmqClient)
	wsHandler := ws.NewWebSocketHandler(stores)

	routes.SetupRoutes(e, wsHandler, exampleHandler)

	// Consumer
	wsOrderUpdateChannel := ws.NewWSChannel(
		rabbitmqClient,
		"order_update",
		"ws_order_queue",
		[]string{"ws_order.#"},
		stores,
	)
	if err := wsOrderUpdateChannel.StartConsumer(); err != nil {
		log.Fatalf("Failed to start order_update consumer: %v", err)
	}

	if err := e.Start(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
