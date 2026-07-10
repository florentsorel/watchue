package handler

import (
	"net/http"

	"github.com/florentsorel/watchue/internal/db"
	"github.com/labstack/echo/v5"
)

type settingsResponse struct {
	TelegramEnabled    bool   `json:"telegram_enabled"`
	TelegramConfigured bool   `json:"telegram_configured"`
	HueBridgeHost      string `json:"hue_bridge_host"`
	BridgeOnline       bool   `json:"bridge_online"`
	Version            string `json:"version"`
}

// GetSettings returns the current notification-channel toggles plus
// non-secret config status (never the bot token/chat id themselves).
func (h *Handler) GetSettings(c *echo.Context) error {
	enabled, err := h.db.GetBoolSetting(c.Request().Context(), db.TelegramEnabledKey, true)
	if err != nil {
		return jsonInternalError(c, err)
	}
	return c.JSON(http.StatusOK, settingsResponse{
		TelegramEnabled:    enabled,
		TelegramConfigured: h.cfg.TelegramBotToken != "",
		HueBridgeHost:      h.cfg.HueBridgeHost,
		BridgeOnline:       h.bridgeOnline.Load(),
		Version:            h.version,
	})
}

type putTelegramEnabledRequest struct {
	Enabled *bool `json:"enabled"`
}

// PutTelegramEnabled turns the Telegram notification channel on or off.
func (h *Handler) PutTelegramEnabled(c *echo.Context) error {
	var req putTelegramEnabledRequest
	if err := c.Bind(&req); err != nil {
		return jsonError(c, http.StatusBadRequest, "invalid request body")
	}
	if req.Enabled == nil {
		return jsonError(c, http.StatusBadRequest, "enabled is required")
	}

	if err := h.db.SetBoolSetting(c.Request().Context(), db.TelegramEnabledKey, *req.Enabled); err != nil {
		return jsonInternalError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}
