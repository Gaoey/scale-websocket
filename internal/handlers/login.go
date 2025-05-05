package handlers

import (
	"net/http"

	"github.com/Gaoey/scale-websocket/internal/auth"
	"github.com/labstack/echo/v4"
)

// LoginRequest represents the login request body
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Token string `json:"token"`
}

// Login handles user authentication and JWT token generation
func Login(c echo.Context) error {
	// Parse request
	req := new(LoginRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}

	// In a real app, validate credentials against a database
	// This is just a simple example
	user, exists := auth.MockUsers[req.Username]
	if !exists {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid credentials")
	}

	if req.Password == user.Password {
		// Generate JWT token
		token, err := auth.GenerateToken(user.UserID, req.Username)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Could not generate token")
		}

		// Return the token
		return c.JSON(http.StatusOK, LoginResponse{
			Token: token,
		})
	}

	return echo.NewHTTPError(http.StatusUnauthorized, "Invalid credentials")
}
