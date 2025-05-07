package ws

// Message represents a WebSocket message
type Message struct {
	Event   string      `json:"event"`
	Status  string      `json:"status"`
	Data    interface{} `json:"data"`
	Channel string      `json:"channel,omitempty"`
}

func NewSuccessMessage(event string, data interface{}) Message {
	return Message{
		Event:  event,
		Status: "1001",
		Data:   data,
	}
}

func NewErrorMessage(event string, status string, errorMsg string) Message {
	return Message{
		Event:  event,
		Status: status,
		Data: map[string]string{
			"error": errorMsg,
		},
	}
}
