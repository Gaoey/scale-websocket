package main

import (
	"log"

	"github.com/Gaoey/scale-websocket/internal/routes"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	routes.SetupRoutes(e)

	if err := e.Start(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
