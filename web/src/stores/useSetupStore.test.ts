import { describe, it, expect, vi, beforeEach } from "vitest"
import { createPinia, setActivePinia } from "pinia"
import { useSetupStore } from "./useSetupStore"

function jsonResponse(body: unknown, ok = true, status = 200): Response {
  return {
    ok,
    status,
    json: () => Promise.resolve(body),
  } as Response
}

beforeEach(() => {
  setActivePinia(createPinia())
  vi.useFakeTimers()
})

describe("useSetupStore", () => {
  it("checkStatus reflects configured state and bridge host", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(jsonResponse({ configured: true, hue_bridge_host: "192.168.1.10" }))
    )
    const store = useSetupStore()

    const configured = await store.checkStatus()

    expect(configured).toBe(true)
    expect(store.configured).toBe(true)
    expect(store.hueBridgeHost).toBe("192.168.1.10")
  })

  it("startPairing transitions to paired once the bridge reports success", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(jsonResponse({ paired: false, reason: "waiting_for_button" }))
      .mockResolvedValueOnce(jsonResponse({ paired: true }))
    vi.stubGlobal("fetch", fetchMock)
    const store = useSetupStore()

    const promise = store.startPairing()
    expect(store.status).toBe("waiting_for_button")

    await vi.runOnlyPendingTimersAsync()
    await promise

    expect(store.status).toBe("paired")
    expect(fetchMock).toHaveBeenCalledTimes(2)
  })

  it("startPairing surfaces a hard error without retrying", async () => {
    vi.stubGlobal(
      "fetch",
      vi
        .fn()
        .mockResolvedValue(jsonResponse({ error: "failed to reach the Hue bridge" }, false, 502))
    )
    const store = useSetupStore()

    await store.startPairing()

    expect(store.status).toBe("error")
    expect(store.errorMessage).toBe("failed to reach the Hue bridge")
  })

  it("startPairing gives up after the max number of attempts", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(jsonResponse({ paired: false, reason: "waiting_for_button" }))
    )
    const store = useSetupStore()

    const promise = store.startPairing()
    await vi.runAllTimersAsync()
    await promise

    expect(store.status).toBe("error")
    expect(store.errorMessage).toMatch(/press the button again/i)
  })

  it("waitForRestart keeps polling through failures until the server responds", async () => {
    const fetchMock = vi
      .fn()
      .mockRejectedValueOnce(new Error("connection refused"))
      .mockResolvedValueOnce(jsonResponse({ configured: true, hue_bridge_host: "192.168.1.10" }))
    vi.stubGlobal("fetch", fetchMock)
    const store = useSetupStore()

    const promise = store.waitForRestart()
    expect(store.status).toBe("restarting")

    await vi.runOnlyPendingTimersAsync()
    await promise

    expect(store.configured).toBe(true)
  })
})
