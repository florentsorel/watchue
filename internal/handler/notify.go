package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/florentsorel/watchue/internal/db"
	"github.com/labstack/echo/v5"
)

// envConfiguredNotifyProvider reports which provider, if any, is configured via env.
func (h *Handler) envConfiguredNotifyProvider() (provider string, ok bool) {
	if h.cfg.TelegramBotToken != "" && h.cfg.TelegramChatID != "" {
		return "telegram", true
	}
	if h.cfg.DiscordWebhookURL != "" {
		return "discord", true
	}
	return "", false
}

// notifyProvider resolves the active provider, env taking precedence over DB.
func (h *Handler) notifyProvider(ctx context.Context) (string, error) {
	if provider, ok := h.envConfiguredNotifyProvider(); ok {
		return provider, nil
	}
	v, err := h.db.GetSetting(ctx, db.NotifyProviderKey)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return v, nil
}

// providerConfigured reports whether usable credentials exist for provider,
// via env or a previous /provider save — independent of which is active.
func (h *Handler) providerConfigured(ctx context.Context, provider string) (bool, error) {
	if envProvider, ok := h.envConfiguredNotifyProvider(); ok && envProvider == provider {
		return true, nil
	}
	switch provider {
	case "telegram":
		botToken, err := h.db.GetSetting(ctx, db.NotifyTelegramBotTokenKey)
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		chatID, err := h.db.GetSetting(ctx, db.NotifyTelegramChatIDKey)
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		return botToken != "" && chatID != "", nil
	case "discord":
		webhookURL, err := h.db.GetSetting(ctx, db.NotifyDiscordWebhookURLKey)
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		return webhookURL != "", nil
	default:
		return false, nil
	}
}

// storedNotifyConfig reads provider's previously-saved credentials, erroring if none are stored.
func (h *Handler) storedNotifyConfig(ctx context.Context, provider string) (NotifyConfig, error) {
	switch provider {
	case "telegram":
		botToken, err := h.db.GetSetting(ctx, db.NotifyTelegramBotTokenKey)
		if err != nil {
			return NotifyConfig{}, errors.New("telegram is not configured yet")
		}
		chatID, err := h.db.GetSetting(ctx, db.NotifyTelegramChatIDKey)
		if err != nil {
			return NotifyConfig{}, errors.New("telegram is not configured yet")
		}
		return NotifyConfig{Provider: "telegram", TelegramBotToken: botToken, TelegramChatID: chatID}, nil
	case "discord":
		webhookURL, err := h.db.GetSetting(ctx, db.NotifyDiscordWebhookURLKey)
		if err != nil {
			return NotifyConfig{}, errors.New("discord is not configured yet")
		}
		return NotifyConfig{Provider: "discord", DiscordWebhookURL: webhookURL}, nil
	default:
		return NotifyConfig{}, fmt.Errorf("unknown provider %q", provider)
	}
}

type notifyProviderStatus struct {
	Configured bool `json:"configured"`
}

type notifyStatusResponse struct {
	ActiveProvider string `json:"active_provider"`
	// EnvLocked is true when a provider is configured via env — PostNotify/PostNotifyActivate 409 then.
	EnvLocked bool                 `json:"env_locked"`
	Telegram  notifyProviderStatus `json:"telegram"`
	Discord   notifyProviderStatus `json:"discord"`
}

// GetNotify reports each provider's configured status independently, plus which is active.
func (h *Handler) GetNotify(c *echo.Context) error {
	ctx := c.Request().Context()

	active, err := h.notifyProvider(ctx)
	if err != nil {
		return jsonInternalError(c, err)
	}
	_, envLocked := h.envConfiguredNotifyProvider()
	telegramConfigured, err := h.providerConfigured(ctx, "telegram")
	if err != nil {
		return jsonInternalError(c, err)
	}
	discordConfigured, err := h.providerConfigured(ctx, "discord")
	if err != nil {
		return jsonInternalError(c, err)
	}

	return c.JSON(http.StatusOK, notifyStatusResponse{
		ActiveProvider: active,
		EnvLocked:      envLocked,
		Telegram:       notifyProviderStatus{Configured: telegramConfigured},
		Discord:        notifyProviderStatus{Configured: discordConfigured},
	})
}

type notifyTestResponse struct {
	OK bool `json:"ok"`
}

// notifyConfigBlank reports whether cfg carries no typed credentials for its own provider.
func notifyConfigBlank(cfg NotifyConfig) bool {
	switch cfg.Provider {
	case "telegram":
		return cfg.TelegramBotToken == "" && cfg.TelegramChatID == ""
	case "discord":
		return cfg.DiscordWebhookURL == ""
	default:
		return true
	}
}

// PostNotifyTest sends a real test notification without touching the DB. If the
// request carries no typed credentials, it falls back to the stored config (see
// notifyConfigBlank), so re-verifying an already-configured provider needs no retyping.
func (h *Handler) PostNotifyTest(c *echo.Context) error {
	ctx := c.Request().Context()

	var cfg NotifyConfig
	if err := c.Bind(&cfg); err != nil {
		return jsonError(c, http.StatusBadRequest, "invalid request body")
	}

	if notifyConfigBlank(cfg) {
		stored, err := h.storedNotifyConfig(ctx, cfg.Provider)
		if err != nil {
			return jsonError(c, http.StatusBadRequest, err.Error())
		}
		cfg = stored
	}

	notifier, err := h.notifierFactory(cfg)
	if err != nil {
		return jsonError(c, http.StatusBadRequest, err.Error())
	}
	if err := notifier.SendTest(ctx); err != nil {
		return jsonProviderError(c, err)
	}
	return c.JSON(http.StatusOK, notifyTestResponse{OK: true})
}

// PostNotify stores the given provider+credentials and hot-swaps the active
// Notifier (see NotifierStore). 409s if a provider is configured via env.
func (h *Handler) PostNotify(c *echo.Context) error {
	ctx := c.Request().Context()

	if _, ok := h.envConfiguredNotifyProvider(); ok {
		return jsonError(c, http.StatusConflict, "already configured via env")
	}

	var cfg NotifyConfig
	if err := c.Bind(&cfg); err != nil {
		return jsonError(c, http.StatusBadRequest, "invalid request body")
	}

	notifier, err := h.notifierFactory(cfg)
	if err != nil {
		return jsonError(c, http.StatusBadRequest, err.Error())
	}

	if err := h.db.SetSetting(ctx, db.SetSettingParams{Key: db.NotifyProviderKey, Value: cfg.Provider}); err != nil {
		return jsonInternalError(c, err)
	}
	switch cfg.Provider {
	case "telegram":
		if err := h.db.SetSetting(ctx, db.SetSettingParams{Key: db.NotifyTelegramBotTokenKey, Value: cfg.TelegramBotToken}); err != nil {
			return jsonInternalError(c, err)
		}
		if err := h.db.SetSetting(ctx, db.SetSettingParams{Key: db.NotifyTelegramChatIDKey, Value: cfg.TelegramChatID}); err != nil {
			return jsonInternalError(c, err)
		}
	case "discord":
		if err := h.db.SetSetting(ctx, db.SetSettingParams{Key: db.NotifyDiscordWebhookURLKey, Value: cfg.DiscordWebhookURL}); err != nil {
			return jsonInternalError(c, err)
		}
	}

	h.notifierStore.Set(notifier)
	return c.NoContent(http.StatusNoContent)
}

type notifyActivateRequest struct {
	Provider string `json:"provider"`
}

// PostNotifyActivate switches the active provider to one with stored
// credentials, without resending or re-testing them. 409s if a provider is
// configured via env.
func (h *Handler) PostNotifyActivate(c *echo.Context) error {
	ctx := c.Request().Context()

	if _, ok := h.envConfiguredNotifyProvider(); ok {
		return jsonError(c, http.StatusConflict, "already configured via env")
	}

	var req notifyActivateRequest
	if err := c.Bind(&req); err != nil {
		return jsonError(c, http.StatusBadRequest, "invalid request body")
	}

	cfg, err := h.storedNotifyConfig(ctx, req.Provider)
	if err != nil {
		return jsonError(c, http.StatusBadRequest, err.Error())
	}

	notifier, err := h.notifierFactory(cfg)
	if err != nil {
		return jsonInternalError(c, err)
	}

	if err := h.db.SetSetting(ctx, db.SetSettingParams{Key: db.NotifyProviderKey, Value: req.Provider}); err != nil {
		return jsonInternalError(c, err)
	}

	h.notifierStore.Set(notifier)
	return c.NoContent(http.StatusNoContent)
}
