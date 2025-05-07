package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Gaoey/scale-websocket/internal/repository/rabbitmq"
	"github.com/Gaoey/scale-websocket/internal/stores"
	"github.com/Gaoey/scale-websocket/services/example"
	"github.com/Gaoey/scale-websocket/services/routes"
	"github.com/Gaoey/scale-websocket/services/store"
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
	fmt.Printf("config rabbit: %v\n", os.Getenv("RABBITMQ_URL"))

	rabbitmqClient, err := rabbitmq.NewClient(rabbitmq.Config{
		URL:          os.Getenv("RABBITMQ_URL"),
		ExchangeName: "ws_events",
		ExchangeType: "topic",
	})
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ client: %v", err)
	}
	e := echo.New()

	storeHandler := store.NewStoreHandler(stores)
	exampleHandler := example.NewExampleHandler(rabbitmqClient)
	wsHandler := ws.NewWebSocketHandler(stores)

	routes.SetupRoutes(e, wsHandler, exampleHandler, storeHandler)

	// Consumer
	wsOrderUpdateChannel := ws.NewWSChannel(
		rabbitmqClient,
		"order_update",
		"ws_order_queue",
		[]string{"ws_order.update"},
		stores,
	)

	if err := wsOrderUpdateChannel.StartConsumer(); err != nil {
		log.Fatalf("Failed to start order_update consumer: %v", err)
	}

	go func() {
		if err := e.Start(":8080"); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	// When shutting down, stop the channel properly
	log.Println("Stopping WebSocket channels...")
	wsOrderUpdateChannel.Stop()
	// Close RabbitMQ connections
	log.Println("Closing RabbitMQ connections...")
	rabbitmqClient.Close()

	log.Println("Server exited properly")
}
