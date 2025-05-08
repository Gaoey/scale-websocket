package example

import (
	"net/http"

	"github.com/Gaoey/scale-websocket/internal/repository/rabbitmq"
	"github.com/labstack/echo/v4"
)

type BodyPayload struct {
	RoutingKey string      `json:"routing_key"`
	Message    interface{} `json:"message"`
}

type ExampleHandler struct {
	Client *rabbitmq.Client
}

func NewExampleHandler(client *rabbitmq.Client) *ExampleHandler {
	return &ExampleHandler{
		Client: client,
	}
}

func (h *ExampleHandler) PublishMessage(c echo.Context) error {
	// Parse request body
	var payload BodyPayload
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	// Publish message to RabbitMQ
	err := h.Client.Publish(c.Request().Context(), payload.RoutingKey, payload.Message)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to publish message",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status": "message published successfully",
	})
}
