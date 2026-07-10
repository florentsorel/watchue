package handler

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/florentsorel/watchue/internal/db"
	"github.com/florentsorel/watchue/internal/hue"
	"github.com/labstack/echo/v5"
)

type setupStatusResponse struct {
	Configured    bool   `json:"configured"`
	HueBridgeHost string `json:"hue_bridge_host"`
}

// GetSetupStatus reports whether a Hue app key is available (env or DB), so
// the web app knows whether to show the dashboard or the /setup flow. Always
// a live check, never cached: pairing can complete mid-poll within the same
// process.
func (h *Handler) GetSetupStatus(c *echo.Context) error {
	configured, err := h.hasAppKey(c)
	if err != nil {
		return jsonInternalError(c, err)
	}
	return c.JSON(http.StatusOK, setupStatusResponse{
		Configured:    configured,
		HueBridgeHost: h.cfg.HueBridgeHost,
	})
}

type pairResponse struct {
	Paired bool   `json:"paired"`
	Reason string `json:"reason,omitempty"`
}

// PostSetupPair attempts one pairing exchange against HUE_BRIDGE_HOST. On
// success, it stores the app key under db.HueAppKeyKey and triggers a
// graceful restart (h.stop) so cmd/web picks up the new key on the next
// boot — see the pairing-setup design notes for why this doesn't hot-reload
// the Hue client in place.
func (h *Handler) PostSetupPair(c *echo.Context) error {
	ctx := c.Request().Context()

	if h.cfg.HueAppKey != "" {
		return jsonError(c, http.StatusConflict, "already configured via HUE_APP_KEY")
	}

	key, err := h.pair(ctx, h.cfg.HueBridgeHost)
	if errors.Is(err, hue.ErrLinkButtonNotPressed) {
		return c.JSON(http.StatusOK, pairResponse{Paired: false, Reason: "waiting_for_button"})
	}
	if err != nil {
		return jsonBridgeError(c, err)
	}

	if err := h.db.SetSetting(ctx, db.SetSettingParams{Key: db.HueAppKeyKey, Value: key}); err != nil {
		return jsonInternalError(c, err)
	}

	if err := c.JSON(http.StatusOK, pairResponse{Paired: true}); err != nil {
		return err
	}
	h.stop()
	return nil
}

func (h *Handler) hasAppKey(c *echo.Context) (bool, error) {
	if h.cfg.HueAppKey != "" {
		return true, nil
	}
	_, err := h.db.GetSetting(c.Request().Context(), db.HueAppKeyKey)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
