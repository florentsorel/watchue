# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Watchue detects when specific Philips Hue zones/rooms/lights are switched on or off, records the
history, and sends a Telegram notification. A web app (Vue) lets the user browse zones/rooms,
choose what to watch, mute notifications per resource, and review history. It is a personal/hobby
project, not a public product.

## Architecture

```
[ Philips Hue Bridge ]
        │  SSE: GET /eventstream/clip/v2 (HTTPS, app-key header)
        ▼
[ Go backend, Docker container on an LXC host ]
        │  filters events against configured zones/rooms/lights, records history
        ▼
[ Telegram Bot API ]                    [ Web app (Vue), served by the same backend ]
        ▼
[ Telegram client — phone/desktop/web, no app to build ]
```

Key decisions made for this architecture (don't relitigate without reason):

- **Hue events**: Hue Bridge v2's CLIP API exposes events as Server-Sent Events. No SSE library is
  needed — plain `net/http` client streaming the response body is enough.
- **Notification channel is Telegram**: needs no Firebase project, no OAuth2, no app to build at
  all — Telegram already has clients everywhere.
- **Channels are meant to be pluggable**: the user mentioned possibly adding Discord later. Don't
  over-build a generic multi-channel abstraction preemptively for a single real channel (YAGNI) —
  but keep `internal/telegram`'s shape (a small `Send(ctx, text) error` client) easy to replicate for
  a second channel when/if that actually happens, rather than entangling Telegram specifics into
  the event loop itself.
- **Two independent on/off switches for notifications**: `watched_resources.notify` (per resource —
  "mute this specific zone/room/light, but keep recording its history") and the `settings` table's
  `telegram_enabled` (global — "turn the whole channel off/on"). A muted resource's changes are
  still recorded to `events`; only the send is skipped. See `internal/watch`'s `Change.Notify` and
  `cmd/web`'s event loop.
- **Terminology matches the Hue CLIP v2 API**: `zone` and `room` are Hue's own resource type names
  (confirmed against Hue's API reference/docs) — don't rename these to "area": `area` is a distinct,
  newer Hue concept (Bridge Pro / MotionAware™ motion-sensing zones), unrelated to what we watch
  here.
- **Config/state storage**: SQLite (watched zones/lights, settings), via
  `modernc.org/sqlite` (pure Go, CGO-free, cross-compile-friendly for the LXC target). Migrations
  run with `goose` (embedded, applied automatically on `db.Open`), query code generated with `sqlc`
  — mirrors the setup in the `postr` project. Never hand-edit generated files under `internal/db`
  (`db.go`, `models.go`, `*.sql.go`); edit the `.sql` under `internal/db/queries/` and re-run
  `sqlc generate` (or `make generate`).
- **Deployment**: Docker image (multi-stage `Dockerfile`: Vite frontend build → Go binary →
  `gcr.io/distroless/static-debian13` runtime — mirrors `postr`'s Dockerfile exactly), published
  to `ghcr.io/florentsorel/watchue` by `.github/workflows/docker.yml` on push to `main` (`:latest`)
  and on `v*` tags (semver tags derived straight from the git tag via `docker/metadata-action` —
  no version bump in a tracked file, just tag-and-push, matching `charset.school`'s release flow).
  `compose.yml` itself is intentionally not committed — documented inline in the README's Quick
  Start instead, same as `postr`. Supersedes an earlier plain-binary-via-systemd-in-LXC plan
  (no reason beyond "hadn't set up CI yet" — don't relitigate without one).
- **Version display (footer)**: no package.json-bump PR, deliberately — that's exactly the release
  pattern the tag-push flow above was chosen to avoid. Instead `docker.yml` passes the tag resolved
  by `docker/metadata-action` (`steps.meta.outputs.version`) as a Docker build arg, which the
  `Dockerfile` bakes into the Go binary via `-ldflags "-X main.version=..."` (`cmd/web`'s `version`
  var, `"dev"` outside a released image). Flows through `handler.New`'s `version` param → `GET
  /api/settings`'s `version` field → `useSettingsStore.version` → the dashboard footer. Zero files
  to touch for a release beyond pushing the git tag.
- **Config loading**: `internal/config`, using `caarlos0/env` (env-var struct tags, validated in
  `Load()`) — same pattern as `postr`. Bridge host is a plain required env var (`HUE_BRIDGE_HOST`),
  not auto-discovered: this is a fixed home LXC service talking to one known bridge, not a client
  needing onboarding UX. The app-key (`HUE_APP_KEY`) requires a one-time physical bridge button
  press to obtain regardless of who performs the pairing call — not worth automating away yet.
  `TELEGRAM_BOT_TOKEN`/`TELEGRAM_CHAT_ID` are optional (Telegram simply stays unconfigured — events
  still get recorded — if unset) but must be set together if used at all.
- **Logging**: `log/slog` with `lmittmann/tint` for readable console output — same as `postr`.
- **HTTP layer**: `echo/v5`, same as `postr`. Echo v5 dropped `Echo.Shutdown` — for graceful
  shutdown, wrap `e` in a plain `*http.Server{Handler: e}` and call the standard library's
  `srv.Shutdown(ctx)`/`srv.ListenAndServe()` instead of `e.Start()`.
- **Resource grouping for the client**: the bridge's own data model doesn't nest lights under
  zones/rooms directly — `internal/catalog` assembles it. Zone `children` reference lights
  directly (`rtype: light`); Room `children` reference devices (`rtype: device`), so a room's
  lights are found by matching `Light.Owner` against those device ids, not via a separate
  `device` resource fetch. Each zone/room's own on/off state comes from the `grouped_light`
  resource referenced in its `services`, not from the zone/room resource itself.
- **Event matching (`internal/watch`)**: an eventstream event for a `light` is matched against
  `watched_resources` by its own id. An event for a `grouped_light` (a zone/room's aggregate
  on/off state) is matched by `EventData.Owner.RID` instead — `watched_resources` stores the
  zone/room's own id, not the grouped_light's id, and the eventstream payload already carries
  that owner reference (no extra bridge call needed to resolve it).
- **Real-time updates are SSE, not WebSocket**: `postr` already has this exact pattern (its Plex
  import progress stream), and it's a good fit here too — one-way server→browser push, no need for
  the client to send anything back, works through the Vite dev proxy transparently. `internal/stream`
  is a tiny in-process pub/sub `Hub`; `cmd/web`'s event loop publishes a `resource` message for
  *every* bridge on/off change (watched or not — the browse view needs to stay live for unwatched
  lights too) and an `event` message only when a watched change gets recorded. A slow/stalled
  subscriber has messages dropped rather than blocking the publisher (bounded channel + non-blocking
  send) — broadcasting must never stall the event loop that also handles Telegram/history.

## Repository layout

Flat at the repo root, same shape as `postr` — Go module at top level, `web/` a sibling directory
for the frontend:

- `go.mod` — module `github.com/florentsorel/watchue`.
- `cmd/web` — entrypoint: loads config, connects to the bridge, runs the HTTP API and the
  event-subscription loop concurrently (both tied to the same `signal.NotifyContext`; the event
  loop reconnects with exponential backoff on stream drop).
- `internal/hue` — CLIP v2 API wrapper: `Lights()`/`Zones()`/`Rooms()`/`GroupedLights()` REST
  calls plus `Subscribe()` for the SSE event stream.
- `internal/catalog` — assembles zones/rooms with their nested lights and on/off state for the
  client (see grouping note above). Depends on an unexported `HueClient` interface, not the
  concrete `*hue.Client`, so it's testable with a mock instead of a fake HTTP server.
  `Catalog.Resolve(id)` looks up a given id against the bridge's actual current resources and
  returns its real type/name — used by `PutWatched` so the client only ever needs to send the
  `id` (from `GET /api/zones`/`GET /api/rooms`); type and display name are never trusted from
  the request body, both to reject (`422`) a stale/typo'd/deleted id instead of silently storing
  it, and to keep the cached name in sync with the bridge on every re-watch. It checks light ids
  via `Rooms`' nested lights, not `Zones`': every light belongs to exactly one room (a bridge
  requirement) whereas zone membership is a partial/optional subset.
- `internal/handler` — Echo v5 HTTP handlers: `GET /api/zones`, `GET /api/rooms`,
  `GET|PUT|PATCH|DELETE /api/watched(/:id)` (`PATCH` toggles `notify` only — see the two-switches
  decision above), `GET /api/events`, `GET /api/settings` + `PUT /api/settings/telegram-enabled`
  (`GetSettings` also returns `telegram_configured`/`hue_bridge_host` — non-secret config status
  the web app's Settings page shows; never the bot token/chat id themselves), `GET /api/stream`
  (SSE — see real-time decision above).
  `GetZones`/`GetRooms` use the `HueClient`-interface + mock pattern (mirroring `postr`'s
  `PlexClient`); the DB-backed handlers instead hold a concrete `*db.Queries` (plus `*config.Config`
  for the non-secret status fields, same as `postr`'s `Handler.config`, and `*stream.Hub` for
  `GetStream`) and are tested against a real `db.Open(":memory:")`, matching `postr`'s own
  handler-test convention (mock the network dependency, use the real DB for SQL correctness).
- `internal/watch` — matches eventstream data against `watched_resources` (see event matching
  note above). `ResolveResourceID` (the id-resolution half of `Match`, without the DB lookup) is
  exported and reused by `cmd/web` to broadcast raw resource updates for *unwatched* resources too.
  Depends on a narrow `Queries` interface, not `*db.Queries`, for mock testing.
- `internal/stream` — the real-time pub/sub `Hub` (see decision above).
  `Change.Notify` carries the resource's own mute state through to the caller.
- `internal/telegram` — `Client.Send(ctx, text)` posts to the Bot API's `sendMessage`, checking
  the `{"ok":bool,"description":...}` envelope (the authoritative success signal, not just HTTP
  status). White-box test file (`package telegram`, not `telegram_test`) overrides the unexported
  `baseURL` to point at an `httptest` server — there's no public constructor param for it since
  real callers never need anything but the real API.
- `internal/config` — env-based config loading/validation.
- `internal/db` — SQLite access: `migrations/` (goose), `queries/` (sqlc input), most of the rest
  is sqlc-generated except `helpers.go` (hand-written companion functions on `*Queries`, e.g.
  `GetBoolSetting`/`SetBoolSetting` — safe to hand-write here since sqlc only regenerates files
  matching its own `.sql` sources, never touches files it didn't generate). Tables:
  `watched_resources` (light/zone/room ids the user watches, plus `notify` — an
  `INTEGER` 0/1 per-resource mute flag, intentionally excluded from `UpsertWatchedResource`'s
  `ON CONFLICT` update so re-watching never un-mutes), `events` (history of on/off changes for
  watched resources — `on_state` is `INTEGER`
  0/1, same convention as `postr`'s `library_settings.enabled`; `outcome` is `"sent"`/`"muted"`/
  `"channel_off"`, computed once in `cmd/web`'s event loop and stored on the row rather than
  recomputed later — recomputing from *current* `notify`/`telegram_enabled` would silently
  rewrite history whenever those settings change), `settings` (generic key/value, currently just
  `telegram_enabled`).
- `internal/web` — `//go:embed all:dist` + a `Handler()` serving the built Vue app with SPA
  fallback (unknown paths serve `index.html`, same as `postr/internal/web`). `dist/` is gitignored
  except a tracked `.gitkeep` placeholder, so a fresh checkout still compiles (`go:embed` needs the
  dir to exist) even before `cd web && npm run build` has been run once.
- `web/` — Vue web app, mirrors `postr/web` exactly:
  - Vite + Vue + TypeScript, plain **Tailwind CSS v4** via the `@tailwindcss/vite` plugin.
    Started out on Nuxt UI v4 (matching `postr`), but nothing beyond its root `UApp` wrapper was
    ever actually used — no `UButton`/`UCard`/etc. anywhere, since the whole UI is a pixel-perfect
    mockup translation with its own hand-built components — so it was dropped entirely (removed
    ~193 transitive packages; CSS bundle went from ~196KB to ~26KB). `src/assets/main.css` just
    does `@import "tailwindcss";` plus a `@theme static { ... }` block mapping our own `--wq-*` CSS
    custom properties (light/dark values swapped via a `.dark` class on `<html>`, toggled by
    `useUiStore`, which defaults to the browser's `prefers-color-scheme` when no `watchue-theme` is
    stored yet) to Tailwind color utilities (`bg-wq-panel`, `text-wq-muted`, etc.) — this is a
    bespoke palette from a design mockup (`_archive/Watchue.html`), not Nuxt UI's own semantic
    colors, which is exactly why dropping Nuxt UI's component library cost nothing visually.
  - **No Tailwind arbitrary-value brackets** (`text-[11px]`, `h-[30px]`, etc.) anywhere in
    `src/**/*.vue` — every pixel-perfect mockup value is a named `@theme` token instead, all
    `wq`-prefixed (`text-wq-11`, `h-wq-30`, `px-wq-17`, ...) in `src/assets/main.css`. The prefix
    isn't just a style preference: several needed values (`3`, `7`, `9`, `11`, `20`, ...) collide
    with Tailwind's own default numeric spacing-scale keys, so an unprefixed `--spacing-7: 7px`
    would silently override the *real* default spacing-7 (28px) wherever it's used unrelatedly.
    `font-size`/`radius`/`tracking`/`shadow` namespaces don't have this collision risk (Tailwind's
    own defaults there are named, not numeric) but got the same prefix for consistency. One
    non-scale keyword got a plain custom class instead of a token: `.rounded-inherit` (`inherit`
    isn't a length). Adding a new one-off pixel value: add a `--<namespace>-wq-<value>` entry to
    the `@theme static` block in `main.css`, don't reach for a bracket.
  - No API client layer — raw `fetch()` in Pinia stores (`src/stores/`, setup-store syntax),
    matching `postr`'s convention exactly (no `api/`/`client.ts`, no shared response DTOs — each
    store declares the small interface it needs).
  - `src/components/` flat (not nested by feature), `src/pages/` for route-level components
    (currently just `DashboardPage.vue` — everything is one continuous page, no router-worthy
    second route yet), routes declared inline in `main.ts` (no `router/` folder) — all matching
    `postr`.
  - Custom icon set (`IconSprite.vue` + `AppIcon.vue`, `<use href="#i-name">`) instead of an
    Iconify/Nuxt UI `Icon` set: the mockup's icons (room types, zone, bulb/lightstrip/spot, etc.)
    are hand-drawn and not in a standard set.
  - `vite.config.ts`: `build.outDir` → `../internal/web/dist`; dev server proxies `/api` to
    `http://localhost:8080`.
  - `npm run build` (or `vue-tsc -b --noEmit` for typecheck only), `npm run lint`, `npm run test`
    (vitest + `@testing-library/vue`, tests co-located as `Component.test.ts`).

## Backend (Go) — commands

Run from the repo root. Scope to `./cmd/... ./internal/...`, not bare `./...` — the latter also
walks `web/node_modules`.

```
go build ./cmd/... ./internal/...
go run ./cmd/web
go test ./cmd/... ./internal/...
go test ./internal/hue/... -run TestName -v   # single test
go vet ./cmd/... ./internal/...
gofmt -l cmd internal
make generate   # regenerate internal/db from queries/ + migrations/ (wraps `sqlc generate`)
make build      # build/watchue
make test
```

## Status

Backend is functionally complete end-to-end for its core loop: Hue CLIP v2 client (REST + SSE) in
`internal/hue`, resource grouping (`internal/catalog`), event matching against `watched_resources`
(`internal/watch`, `Change.Notify`-aware), a Telegram sender (`internal/telegram`), env config
including Telegram credentials (`internal/config`), SQLite/goose/sqlc (`internal/db`, tables:
`watched_resources`, `events`, `settings`), and a full HTTP API (`internal/handler`).
`cmd/web` wires all of it and also serves the built frontend (`internal/web`) — the event loop
records every watched change to `events` (with a fixed `outcome`) and sends a Telegram message
unless muted, channel-disabled, or unconfigured.

The web app (`web/`) has a working dashboard: live stats, watched-resources grid (3 layouts, mute/
unwatch), browse zones/rooms with watch toggles (long lists cap at a fixed height and scroll —
`ScrollableList.vue`, reused across panels via an `itemsDisplayed` prop), history with filters
(absolute date/time shown past 24h instead of an ever-growing "Nd ago"), and a settings panel
(Telegram on/off + non-secret config status) — all updating in real time over `/api/stream` (SSE),
not just on page load. Light AND room/zone icons are archetype-driven: `catalog.Group`/`Light` both
carry the bridge's `metadata.archetype`, mapped in `resourceIcon.ts` to SVG sprite symbols
(`i-hue-*` in `IconSprite.vue`, sourced from `arallsopp/hass-hue-icons`, CC BY-NC-SA — fine for this
non-commercial hobby project) that tint via `currentColor` like the rest of the icon set; a handful
of archetypes with no matching icon (`garden`, `top_floor`, `man_cave`, `music`, `tv`, `reading`)
fall back to a name-keyword guess. `ResourceIconBox.vue` no longer special-cases images — everything
is a tintable SVG now, no more always-colored tile. Loading states across the dashboard show a
spinner (`i-spinner`, `animate-spin`), not a skeleton placeholder. Favicon + `docs/assets/logo.svg`
mirror the header logo, now used in the root `README.md`. Verified end-to-end against the real
compiled binary (fake bridge, real SQLite) — build/typecheck/lint/tests all green on both sides.
Not yet built/verified: real interaction against a live Hue bridge + a real Telegram bot in a
browser, the bridge pairing helper (a manual curl-based workaround is documented in the README),
and any UI polish pass beyond the initial mockup translation.

Also since the last update: `cmd/web` no longer exits at startup if the bridge is unreachable
(that fully defeated the point of live bridge-status tracking — an unreachable bridge should show
"offline" in the UI, not crash the whole app); the initial `Zones()` check just logs a warning and
`bridgeOnline` starts `false`, with `runEventLoop`'s own retry/backoff taking it from there. Live
bridge connectivity is now pushed to the client via a `bridge_status` SSE message (not just a
load-time snapshot), read from `useSettingsStore.bridgeOnline`; a `BridgeOfflineBanner.vue` shows
in "Browse your bridge" and "Watched resources" when offline instead of silently rendering
zeros/false. Telegram messages are HTML-formatted (`parse_mode: "HTML"`, resource name
`html.EscapeString`'d) instead of plain text. Nuxt UI is gone entirely (see the `web/` bullet
above) and so is every Tailwind arbitrary-value bracket. Root `LICENSE.txt` (MIT), `README.md`,
`Dockerfile`, `.dockerignore`, and `.github/workflows/{ci,docker,yamllint}.yml` +
`.yamllint.yml` now exist too — deployment story mirrors `postr`/`charset.school` (tag-triggered
ghcr.io image, no `compose.yml` committed, documented inline in the README instead).
`.github/renovate.json` also ported from `postr` — needed no adaptation, since watchue's stack
(Go + npm + Dockerfile + github-actions) and its README's `go-1.26.3-00ADD8` badge format already
match postr's custom regex managers exactly.

`ResourceLayout`'s internal values
now match their UI labels (`"compact"`/`"wall"`, previously `"minimal"`/`"board"` — only `"glow"`
ever matched) to stop that exact mismatch from causing confusion again. Also fixed a real layout
bug in `ResourceCard.vue`'s Compact mode: the icon+content+pill+buttons row is wrapped in its own
inner flex container, which had no `w-full`/`flex-1` of its own — invisible in Glow/Wall (that
wrapper sits in a plain block `<div>` there, so it naturally stretches full-width) but very visible
in Compact, where the *outer* row is also `flex`, making the inner wrapper just shrink to its
content's width instead of stretching to the row. Reproduced and confirmed only at real card width
(~1140px+) — narrow-viewport testing hid it enough to look fine at first.

Deployment is now Docker-based (see the Deployment decision above): `Dockerfile`,
`.github/workflows/docker.yml` (builds/pushes to `ghcr.io` on `main` and `v*` tags), root
`LICENSE.txt` (MIT) and `README.md` — mirroring `postr`'s conventions, including not committing a
`compose.yml` (documented inline in the README instead). Not yet done: an actual `git tag` push to
verify the release workflow end-to-end (no commits/remote on this repo yet).
