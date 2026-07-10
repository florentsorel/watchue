import { describe, it, expect, vi, beforeEach } from "vitest"
import { createPinia, setActivePinia } from "pinia"
import { useNotifyStore } from "./useNotifyStore"

function jsonResponse(body: unknown, ok = true, status = 200): Response {
  return {
    ok,
    status,
    json: () => Promise.resolve(body),
  } as Response
}

beforeEach(() => {
  setActivePinia(createPinia())
})

describe("useNotifyStore", () => {
  it("test reports success", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse({ ok: true })))
    const store = useNotifyStore()

    await store.test({
      provider: "discord",
      discord_webhook_url: "https://discord.example/webhook",
    })

    expect(store.testStatus).toBe("success")
  })

  it("test reports failure with the server's error message", async () => {
    vi.stubGlobal(
      "fetch",
      vi
        .fn()
        .mockResolvedValue(
          jsonResponse({ error: "failed to reach the notification provider" }, false, 502)
        )
    )
    const store = useNotifyStore()

    await store.test({
      provider: "discord",
      discord_webhook_url: "https://discord.example/webhook",
    })

    expect(store.testStatus).toBe("error")
    expect(store.testError).toBe("failed to reach the notification provider")
  })

  it("resetTest clears status back to idle", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse({ ok: true })))
    const store = useNotifyStore()
    await store.test({ provider: "telegram", telegram_bot_token: "t", telegram_chat_id: "c" })
    expect(store.testStatus).toBe("success")

    store.resetTest()

    expect(store.testStatus).toBe("idle")
    expect(store.testError).toBe("")
  })

  it("save posts the config and throws on failure", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValue(jsonResponse({ error: "already configured via env" }, false, 409))
    vi.stubGlobal("fetch", fetchMock)
    const store = useNotifyStore()

    await expect(
      store.save({ provider: "telegram", telegram_bot_token: "t", telegram_chat_id: "c" })
    ).rejects.toThrow("already configured via env")
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/notify",
      expect.objectContaining({ method: "POST" })
    )
  })

  it("save resolves on success", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: true } as Response))
    const store = useNotifyStore()

    await expect(
      store.save({ provider: "discord", discord_webhook_url: "https://discord.example/webhook" })
    ).resolves.toBeUndefined()
  })

  it("fetchStatus reports both providers independently", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        jsonResponse({
          active_provider: "discord",
          env_locked: false,
          telegram: { configured: true },
          discord: { configured: true },
        })
      )
    )
    const store = useNotifyStore()

    await store.fetchStatus()

    expect(store.activeProvider).toBe("discord")
    expect(store.envLocked).toBe(false)
    expect(store.status.telegram.configured).toBe(true)
    expect(store.status.discord.configured).toBe(true)
  })

  it("fetchStatus reports env_locked", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        jsonResponse({
          active_provider: "discord",
          env_locked: true,
          telegram: { configured: false },
          discord: { configured: true },
        })
      )
    )
    const store = useNotifyStore()

    await store.fetchStatus()

    expect(store.envLocked).toBe(true)
  })

  it("activate posts the provider and throws on failure", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValue(jsonResponse({ error: "telegram is not configured yet" }, false, 400))
    vi.stubGlobal("fetch", fetchMock)
    const store = useNotifyStore()

    await expect(store.activate("telegram")).rejects.toThrow("telegram is not configured yet")
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/notify/activate",
      expect.objectContaining({ method: "POST", body: JSON.stringify({ provider: "telegram" }) })
    )
  })

  it("activate resolves on success", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: true } as Response))
    const store = useNotifyStore()

    await expect(store.activate("discord")).resolves.toBeUndefined()
  })
})
