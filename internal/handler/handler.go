// Package handler exposes the HTTP API the web app uses.
package handler

import (
	"context"
	"log/slog"
	"net/http"
	"sync/atomic"

	"github.com/florentsorel/watchue/internal/config"
	"github.com/florentsorel/watchue/internal/db"
	"github.com/florentsorel/watchue/internal/hue"
	"github.com/florentsorel/watchue/internal/stream"
	"github.com/labstack/echo/v5"
)

// HueClient is the subset of hue.Client operations the handler needs.
type HueClient interface {
	Lights(ctx context.Context) ([]hue.Light, error)
	Zones(ctx context.Context) ([]hue.Zone, error)
	Rooms(ctx context.Context) ([]hue.Room, error)
	GroupedLights(ctx context.Context) ([]hue.GroupedLight, error)
}

// PairFunc performs the bridge's link-button pairing exchange (see
// hue.Pair). A function type rather than an interface since there's exactly
// one operation and no other bridge-authenticated calls involved.
type PairFunc func(ctx context.Context, bridgeAddr string) (string, error)

type Handler struct {
	hue          HueClient
	db           *db.Queries
	cfg          *config.Config
	hub          *stream.Hub
	bridgeOnline *atomic.Bool
	version      string
	stop         context.CancelFunc
	pair         PairFunc
}

func New(hueClient HueClient, queries *db.Queries, cfg *config.Config, hub *stream.Hub, bridgeOnline *atomic.Bool, version string, stop context.CancelFunc, pair PairFunc) *Handler {
	return &Handler{hue: hueClient, db: queries, cfg: cfg, hub: hub, bridgeOnline: bridgeOnline, version: version, stop: stop, pair: pair}
}

type errorResponse struct {
	Error string `json:"error"`
}

func jsonError(c *echo.Context, status int, msg string) error {
	return c.JSON(status, errorResponse{Error: msg})
}

// jsonInternalError reports an unexpected server-side failure (DB, etc).
func jsonInternalError(c *echo.Context, err error) error {
	slog.Error("internal server error", "error", err, "path", c.Request().URL.Path)
	return jsonError(c, http.StatusInternalServerError, "internal server error")
}

// jsonBridgeError reports a failure to reach the Hue bridge specifically.
func jsonBridgeError(c *echo.Context, err error) error {
	slog.Error("failed to reach the Hue bridge", "error", err, "path", c.Request().URL.Path)
	return jsonError(c, http.StatusBadGateway, "failed to reach the Hue bridge")
}
