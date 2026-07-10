package handler

import (
	"net/http"

	"github.com/florentsorel/watchue/internal/db"
	"github.com/labstack/echo/v5"
)

type settingsResponse struct {
	NotifyEnabled    bool   `json:"notify_enabled"`
	NotifyConfigured bool   `json:"notify_configured"`
	NotifyProvider   string `json:"notify_provider"`
	HueBridgeHost    string `json:"hue_bridge_host"`
	BridgeOnline     bool   `json:"bridge_online"`
	Version          string `json:"version"`
}

// GetSettings returns the current notification-channel toggle plus
// non-secret config status (never credentials themselves).
func (h *Handler) GetSettings(c *echo.Context) error {
	ctx := c.Request().Context()
	enabled, err := h.db.GetBoolSetting(ctx, db.NotifyEnabledKey, true)
	if err != nil {
		return jsonInternalError(c, err)
	}
	provider, err := h.notifyProvider(ctx)
	if err != nil {
		return jsonInternalError(c, err)
	}
	return c.JSON(http.StatusOK, settingsResponse{
		NotifyEnabled:    enabled,
		NotifyConfigured: provider != "",
		NotifyProvider:   provider,
		HueBridgeHost:    h.cfg.HueBridgeHost,
		BridgeOnline:     h.bridgeOnline.Load(),
		Version:          h.version,
	})
}

type putNotifyEnabledRequest struct {
	Enabled *bool `json:"enabled"`
}

// PutNotifyEnabled turns the active notification channel on or off.
func (h *Handler) PutNotifyEnabled(c *echo.Context) error {
	var req putNotifyEnabledRequest
	if err := c.Bind(&req); err != nil {
		return jsonError(c, http.StatusBadRequest, "invalid request body")
	}
	if req.Enabled == nil {
		return jsonError(c, http.StatusBadRequest, "enabled is required")
	}

	if err := h.db.SetBoolSetting(c.Request().Context(), db.NotifyEnabledKey, *req.Enabled); err != nil {
		return jsonInternalError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}
