package main

import (
	"log"

	"github.com/Gaoey/scale-websocket/services/routes"
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

	routes.SetupRoutes(e)

	if err := e.Start(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
