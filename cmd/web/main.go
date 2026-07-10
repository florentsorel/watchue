package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/florentsorel/watchue/internal/config"
	"github.com/florentsorel/watchue/internal/db"
	"github.com/florentsorel/watchue/internal/handler"
	"github.com/florentsorel/watchue/internal/hue"
	"github.com/florentsorel/watchue/internal/stream"
	"github.com/florentsorel/watchue/internal/telegram"
	"github.com/florentsorel/watchue/internal/watch"
	"github.com/florentsorel/watchue/internal/web"
	"github.com/labstack/echo/v5"
	"github.com/lmittmann/tint"
)

// version is set at build time via -ldflags "-X main.version=...";
// "dev" outside a released Docker image.
var version = "dev"

func main() {
	slog.SetDefault(slog.New(tint.NewHandler(os.Stdout, nil)))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	conn, err := db.Open(cfg.DBPath)
	if err != nil {
		slog.Error("failed to open database", "path", cfg.DBPath, "error", err)
		os.Exit(1)
	}
	defer conn.Close()
	queries := db.New(conn)

	appKey := cfg.HueAppKey
	if appKey == "" {
		if v, err := queries.GetSetting(ctx, db.HueAppKeyKey); err == nil {
			appKey = v
		} else if !errors.Is(err, sql.ErrNoRows) {
			slog.Error("failed to read stored hue app key", "error", err)
			os.Exit(1)
		}
	}
	configured := appKey != ""

	client := hue.NewClient(cfg.HueBridgeHost, appKey, nil)

	bridgeOnline := &atomic.Bool{}
	if !configured {
		slog.Warn("Hue app key not configured — visit /setup to pair with the bridge", "host", cfg.HueBridgeHost)
	} else if zones, err := client.Zones(ctx); err != nil {
		slog.Warn("failed to reach bridge at startup — will keep retrying", "host", cfg.HueBridgeHost, "error", err)
	} else {
		slog.Info("connected to bridge", "host", cfg.HueBridgeHost, "zones", len(zones))
		bridgeOnline.Store(true)
	}

	var notifier *telegram.Client
	if cfg.TelegramBotToken != "" {
		notifier = telegram.NewClient(cfg.TelegramBotToken, cfg.TelegramChatID, nil)
		slog.Info("telegram notifications configured")
	} else {
		slog.Warn("TELEGRAM_BOT_TOKEN not set — watched changes will be recorded but never sent")
	}

	hub := stream.NewHub()

	if configured {
		go runEventLoop(ctx, client, queries, notifier, hub, bridgeOnline)
	}

	e := echo.New()
	pair := func(ctx context.Context, bridgeAddr string) (string, error) { return hue.Pair(ctx, bridgeAddr, nil) }
	h := handler.New(client, queries, cfg, hub, bridgeOnline, version, stop, pair)
	e.GET("/api/setup/status", h.GetSetupStatus)
	e.POST("/api/setup/pair", h.PostSetupPair)
	e.GET("/api/zones", h.GetZones)
	e.GET("/api/rooms", h.GetRooms)
	e.GET("/api/watched", h.GetWatched)
	e.PUT("/api/watched/:id", h.PutWatched)
	e.PATCH("/api/watched/:id", h.PatchWatched)
	e.DELETE("/api/watched/:id", h.DeleteWatched)
	e.GET("/api/events", h.GetEvents)
	e.GET("/api/settings", h.GetSettings)
	e.PUT("/api/settings/telegram-enabled", h.PutTelegramEnabled)
	e.GET("/api/stream", h.GetStream)
	e.GET("/*", echo.WrapHandler(web.Handler()))

	// Echo v5 dropped Echo.Shutdown; use a plain http.Server instead.
	srv := &http.Server{Addr: ":8080", Handler: e}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("http server shutdown error", "error", err)
		}
	}()

	slog.Info("http server starting", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("http server stopped", "error", err)
		os.Exit(1)
	}
}

// runEventLoop matches bridge events against watched resources, records
// history, notifies via Telegram, and broadcasts real-time updates to any
// connected web clients. notifier may be nil if unconfigured.
func runEventLoop(ctx context.Context, client *hue.Client, queries *db.Queries, notifier *telegram.Client, hub *stream.Hub, bridgeOnline *atomic.Bool) {
	const maxBackoff = 30 * time.Second
	backoff := time.Second

	for ctx.Err() == nil {
		events, errs := client.Subscribe(ctx)

		for ev := range events {
			for _, d := range ev.Data {
				if d.On == nil {
					continue
				}
				slog.Debug("light state changed", "id", d.ID, "resource", d.Type, "on", d.On.On)

				if resourceID, ok := watch.ResolveResourceID(d); ok {
					broadcastResource(hub, resourceID, d.On.On)
				}

				change, ok, err := watch.Match(ctx, queries, d)
				if err != nil {
					slog.Error("failed to match event against watched resources", "id", d.ID, "error", err)
					continue
				}
				if !ok {
					continue
				}

				onState := int64(0)
				if change.On {
					onState = 1
				}

				outcome := "sent"
				switch {
				case notifier == nil:
					outcome = "channel_off"
				case !change.Notify:
					outcome = "muted"
				default:
					enabled, err := queries.GetBoolSetting(ctx, db.TelegramEnabledKey, true)
					if err != nil {
						slog.Error("failed to read telegram_enabled setting", "error", err)
						outcome = "channel_off" // fail closed: don't send if we can't confirm it's enabled
					} else if !enabled {
						outcome = "channel_off"
					}
				}

				inserted, err := queries.InsertEvent(ctx, db.InsertEventParams{
					ResourceID:   change.ResourceID,
					ResourceType: change.ResourceType,
					Name:         change.Name,
					OnState:      onState,
					Outcome:      outcome,
				})
				if err != nil {
					slog.Error("failed to record event", "id", change.ResourceID, "error", err)
				} else {
					broadcastEvent(hub, inserted)
				}

				slog.Info("watched resource changed", "id", change.ResourceID, "type", change.ResourceType, "name", change.Name, "on", change.On, "outcome", outcome)

				if outcome != "sent" {
					continue
				}

				state := "off"
				if change.On {
					state = "on"
				}
				message := fmt.Sprintf("🔆 <b>%s</b> was turned <b>%s</b>.", html.EscapeString(change.Name), state)
				if err := notifier.Send(ctx, message); err != nil {
					slog.Error("failed to send telegram notification", "id", change.ResourceID, "error", err)
				}
			}
			backoff = time.Second
			setBridgeOnline(hub, bridgeOnline, true)
		}

		if ctx.Err() != nil {
			return
		}

		setBridgeOnline(hub, bridgeOnline, false)
		if err := <-errs; err != nil {
			slog.Warn("eventstream disconnected", "error", err, "retry_in", backoff)
		} else {
			slog.Warn("eventstream disconnected", "retry_in", backoff)
		}

		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return
		}
		if backoff < maxBackoff {
			backoff *= 2
		}
	}
}

func broadcastResource(hub *stream.Hub, id string, on bool) {
	data, err := json.Marshal(stream.ResourceMessage{Kind: "resource", ID: id, On: on})
	if err != nil {
		slog.Error("failed to encode resource stream message", "error", err)
		return
	}
	hub.Publish(data)
}

func broadcastEvent(hub *stream.Hub, e db.Event) {
	data, err := json.Marshal(stream.EventMessage{
		Kind:         "event",
		ID:           e.ID,
		ResourceID:   e.ResourceID,
		ResourceType: e.ResourceType,
		Name:         e.Name,
		On:           e.OnState != 0,
		Outcome:      e.Outcome,
		CreatedAt:    e.CreatedAt,
	})
	if err != nil {
		slog.Error("failed to encode event stream message", "error", err)
		return
	}
	hub.Publish(data)
}

// setBridgeOnline updates the shared bridge connection state and broadcasts
// it, but only on an actual transition — avoids spamming clients on every
// event once already online.
func setBridgeOnline(hub *stream.Hub, bridgeOnline *atomic.Bool, online bool) {
	if bridgeOnline.Swap(online) == online {
		return
	}
	data, err := json.Marshal(stream.BridgeStatusMessage{Kind: "bridge_status", Online: online})
	if err != nil {
		slog.Error("failed to encode bridge status stream message", "error", err)
		return
	}
	hub.Publish(data)
}
