<p align="center">
  <img src="docs/assets/logo.svg" alt="Watchue" width="160" />
</p>

<p align="center">
  A self-hosted Philips Hue on/off watcher with Telegram or Discord notifications.
</p>

<p align="center">
  <img src="https://img.shields.io/badge/docker-ready-blue?logo=docker" alt="Docker ready" />
  <img src="https://img.shields.io/badge/go-1.27.0-00ADD8?logo=go" alt="Go version" />
  <img src="https://img.shields.io/badge/vue-3-42b883?logo=vuedotjs" alt="Vue 3" />
</p>

---

## Overview

Watchue detects when specific Philips Hue zones, rooms or lights are switched on or off,
records the history, and sends a notification via Telegram or Discord (pick one). A web app
lets you browse your bridge, choose what to watch, mute notifications per resource, and review
the history — all updating live as things happen, no page refresh needed.

<p align="center">
  <img src="docs/assets/watchue-light.png" alt="Watchue — light theme" width="49%" />
  <img src="docs/assets/watchue-dark.png" alt="Watchue — dark theme" width="49%" />
</p>

---

## Features

- **Watch zones, rooms and lights** — pick exactly which Hue resources to track from your
  whole bridge.
- **Telegram or Discord notifications** — get pinged the moment a watched resource turns on or
  off, via whichever one you configure (only one is active at a time).
- **Per-resource mute** — keep recording a resource's history without being pinged for it.
- **Global notification toggle** — turn the whole channel off/on without losing any history.
- **History** — every on/off change for a watched resource is recorded, even while muted.
- **Live updates** — the dashboard reflects bridge changes in real time over SSE, including
  lights toggled from outside the app (the Hue app, a physical switch, etc.).

---

## Quick Start

### Docker Compose

Save the following as `compose.yml`:

```yaml
services:
  watchue:
    image: ghcr.io/florentsorel/watchue:latest
    container_name: watchue
    ports:
      - "8080:8080"
    volumes:
      - ./data:/data
    environment:
      DB_PATH: /data/watchue.db
      HUE_BRIDGE_HOST: 192.168.1.x
      # Optional — pick at most one provider (see below), or skip both and
      # configure one later via the in-app /provider page.
      # TELEGRAM_BOT_TOKEN: your-telegram-bot-token
      # TELEGRAM_CHAT_ID: your-telegram-chat-id
      # DISCORD_WEBHOOK_URL: your-discord-webhook-url
    restart: unless-stopped
```

Then:

```sh
docker compose up -d
```

Open [http://localhost:8080](http://localhost:8080) in your browser — with no `HUE_APP_KEY` set, Watchue
redirects you to a `/setup` wizard that pairs with your bridge, then sends you to `/provider` to
pick and test a notification provider (see below for both).

> **Pinning a version** — replace `:latest` with a specific release tag (e.g.
> `ghcr.io/florentsorel/watchue:1.2.3`) to avoid unexpected changes on container restart.
> Available tags are listed on the
> [container registry](https://github.com/florentsorel/watchue/pkgs/container/watchue).

---

## Environment Variables

| Variable | Required | Description |
|---|---|---|
| `DB_PATH` | No | Path to the SQLite database file (default: `data/watchue.db`) |
| `HUE_BRIDGE_HOST` | Yes | IP or hostname of your Philips Hue Bridge (e.g. `192.168.1.10`) |
| `HUE_APP_KEY` | No | Bridge application key — leave unset and pair via `/setup` instead (see below) |
| `TELEGRAM_BOT_TOKEN` | No | Telegram bot token — required if `TELEGRAM_CHAT_ID` is set |
| `TELEGRAM_CHAT_ID` | No | Telegram chat id to notify — required if `TELEGRAM_BOT_TOKEN` is set |
| `DISCORD_WEBHOOK_URL` | No | Discord incoming webhook URL — cannot be set together with the Telegram variables |

Notifications are entirely optional: leave everything unset (or configure a provider later via
`/provider`) and Watchue will keep recording history, it just won't send anything. Only one provider
may be active at a time — setting both a complete Telegram pair and `DISCORD_WEBHOOK_URL` fails
at startup rather than silently picking one.

### Getting a Hue Bridge app key

The Hue Bridge only hands out an application key after a physical button press. The easiest way
to get one is the in-app setup flow:

1. Leave `HUE_APP_KEY` unset and start Watchue — it redirects to `/setup`.
2. Press the physical link button on your Hue Bridge, then click **Start pairing**. Watchue polls
   the bridge and detects the press automatically.
3. Click **Next**. Watchue stores the key and restarts itself to pick it up — this relies on your
   process supervisor restarting a cleanly-exited process, true for the `restart: unless-stopped`
   policy in the Compose file above. If you're running the binary directly (e.g. under `go run`
   or a bare `systemd` unit without `Restart=always`), restart it manually after this step.

If you'd rather obtain the key yourself (e.g. to set `HUE_APP_KEY` directly, scripting a headless
setup, or without ever visiting `/setup`), call the bridge's pairing endpoint by hand instead:

1. Press the physical link button on your Hue Bridge.
2. Within 30 seconds, call the bridge's pairing endpoint:

   ```sh
   curl -k -X POST https://<HUE_BRIDGE_HOST>/api \
     -H "Content-Type: application/json" \
     -d '{"devicetype":"watchue#homelab"}'
   ```

3. The response contains your app key:

   ```json
   [{ "success": { "username": "your-hue-app-key" } }]
   ```

   Use that value as `HUE_APP_KEY`.

### Setting up notifications (optional)

Skip this section if you don't want notifications — Watchue works fine with neither provider
configured, it just won't send anything. The easiest way to set one up is the in-app `/provider`
page (linked right after pairing with your bridge, and any time after from Settings): pick
Telegram or Discord, enter its credentials, send a test message, then save. The credentials are
stored in Watchue's own database and take effect immediately — no env vars or restart needed, and
you can come back and switch providers later.

To obtain the credentials that page asks for:

**Telegram**

1. Create a bot via [@BotFather](https://t.me/BotFather) and grab its token.
2. Message your new bot (or add it to a group), then call
   `https://api.telegram.org/bot<token>/getUpdates` to find your chat id.

**Discord**

1. In the target channel, open Settings → Integrations → Webhooks → New Webhook.
2. Copy its URL.

If you'd rather configure this via env vars instead of `/provider` (e.g. a headless/scripted
setup), set `TELEGRAM_BOT_TOKEN`/`TELEGRAM_CHAT_ID` (both required together) or
`DISCORD_WEBHOOK_URL` — never both.

---

## Data

Watchue writes its SQLite database under the mounted volume:

```
data/
└── watchue.db
```

Application logs (startup, bridge connection, event handling) are written to **stdout**.

### Log rotation

Since logs go to stdout rather than a file in the volume, rotate them via Docker's own logging
driver instead of `logrotate` — add this to your `compose.yml`:

```yaml
    logging:
      driver: json-file
      options:
        max-size: "10m"
        max-file: "5"
```
