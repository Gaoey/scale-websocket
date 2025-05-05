package routes

import (
	"github.com/Gaoey/scale-websocket/internal/handlers"
	"github.com/labstack/echo/v4"
)

func SetupRoutes(e *echo.Echo) {
	e.GET("/health", handlers.HealthCheck)
}
