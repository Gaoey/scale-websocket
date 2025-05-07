package store

import (
	"net/http"

	"github.com/Gaoey/scale-websocket/internal/stores"
	"github.com/labstack/echo/v4"
)

type StoreHandler struct {
	Store *stores.ConnectionStorage
}

func NewStoreHandler(store *stores.ConnectionStorage) *StoreHandler {
	return &StoreHandler{
		Store: store,
	}
}

func (h *StoreHandler) GetAllConnections(c echo.Context) error {
	allStores := h.Store.GetAll()
	return c.JSON(http.StatusOK, allStores)
}
