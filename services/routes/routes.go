package routes

import (
	"net/http"
	"strings"

	"github.com/Gaoey/scale-websocket/services/auth"
	"github.com/Gaoey/scale-websocket/services/healthcheck"
	"github.com/Gaoey/scale-websocket/services/ws"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func SetupRoutes(e *echo.Echo) {
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/health", healthcheck.HealthCheckHandler)
	e.POST("/login", auth.LoginHandler)
	e.GET("/auth-ws", ws.AuthWebSocketHandler)

	auth := e.Group("/api")
	auth.Use(JWTAuth())
	// API auth list
}

// JWTAuth middleware for JWT authentication
func JWTAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get the Authorization header
			authHeader := c.Request().Header.Get("Authorization")

			// Check if the header exists and has the correct format
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing or invalid Authorization header")
			}

			// Extract the token
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			// Validate the token
			claims, err := auth.ValidateToken(tokenString)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired token")
			}

			// Set user information in context
			c.Set("userID", claims.UserID)
			c.Set("username", claims.Username)

			// Continue to the next middleware/handler
			return next(c)
		}
	}
}
