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

// Notifier sends a resource state-change notification through some external
// channel. Implemented by internal/telegram.Client and internal/discord.Client.
type Notifier interface {
	Send(ctx context.Context, resourceName string, on bool) error
	SendTest(ctx context.Context) error
}

// NotifyConfig is a request-bind target ONLY — never return it in a JSON response.
type NotifyConfig struct {
	Provider          string `json:"provider"`
	TelegramBotToken  string `json:"telegram_bot_token,omitempty"`
	TelegramChatID    string `json:"telegram_chat_id,omitempty"`
	DiscordWebhookURL string `json:"discord_webhook_url,omitempty"`
}

// NotifierFactory builds a Notifier for the given provider/credentials.
type NotifierFactory func(NotifyConfig) (Notifier, error)

type Handler struct {
	hue             HueClient
	db              *db.Queries
	cfg             *config.Config
	hub             *stream.Hub
	bridgeOnline    *atomic.Bool
	version         string
	stop            context.CancelFunc
	pair            PairFunc
	notifierFactory NotifierFactory
	notifierStore   *NotifierStore
}

func New(hueClient HueClient, queries *db.Queries, cfg *config.Config, hub *stream.Hub, bridgeOnline *atomic.Bool, version string, stop context.CancelFunc, pair PairFunc, notifierFactory NotifierFactory, notifierStore *NotifierStore) *Handler {
	return &Handler{
		hue:             hueClient,
		db:              queries,
		cfg:             cfg,
		hub:             hub,
		bridgeOnline:    bridgeOnline,
		version:         version,
		stop:            stop,
		pair:            pair,
		notifierFactory: notifierFactory,
		notifierStore:   notifierStore,
	}
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

// jsonProviderError reports a failure to reach a notification provider
// (Telegram/Discord) specifically.
func jsonProviderError(c *echo.Context, err error) error {
	slog.Error("failed to reach the notification provider", "error", err, "path", c.Request().URL.Path)
	return jsonError(c, http.StatusBadGateway, "failed to reach the notification provider")
}
