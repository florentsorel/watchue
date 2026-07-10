package handler

import (
	"net/http"

	"github.com/florentsorel/watchue/internal/catalog"
	"github.com/labstack/echo/v5"
)

// GetZones lists every zone with its lights and its own on/off state.
func (h *Handler) GetZones(c *echo.Context) error {
	cat, err := catalog.Build(c.Request().Context(), h.hue)
	if err != nil {
		return jsonBridgeError(c, err)
	}
	return c.JSON(http.StatusOK, cat.Zones)
}

// GetRooms lists every room with its lights and its own on/off state.
func (h *Handler) GetRooms(c *echo.Context) error {
	cat, err := catalog.Build(c.Request().Context(), h.hue)
	if err != nil {
		return jsonBridgeError(c, err)
	}
	return c.JSON(http.StatusOK, cat.Rooms)
}
