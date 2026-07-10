import { defineStore } from "pinia"
import { ref } from "vue"

export type SetupStatus = "idle" | "waiting_for_button" | "paired" | "restarting" | "error"

const PAIR_POLL_INTERVAL_MS = 2000
const PAIR_MAX_ATTEMPTS = 20 // ~40s, matching the bridge's own ~30s button window plus buffer
const RESTART_POLL_INTERVAL_MS = 1000

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

export const useSetupStore = defineStore("setup", () => {
  const configured = ref(false)
  const hueBridgeHost = ref("")
  const status = ref<SetupStatus>("idle")
  const errorMessage = ref("")

  async function checkStatus(): Promise<boolean> {
    const res = await fetch("/api/setup/status")
    if (!res.ok) {
      throw new Error("Failed to check setup status")
    }
    const data = await res.json()
    configured.value = data.configured
    hueBridgeHost.value = data.hue_bridge_host
    return configured.value
  }

  async function startPairing(): Promise<void> {
    status.value = "waiting_for_button"
    errorMessage.value = ""

    for (let attempt = 0; attempt < PAIR_MAX_ATTEMPTS; attempt++) {
      const res = await fetch("/api/setup/pair", { method: "POST" })
      const data = await res.json().catch(() => ({}))

      if (res.ok && data.paired) {
        status.value = "paired"
        return
      }
      if (!res.ok) {
        status.value = "error"
        errorMessage.value = data.error ?? "Failed to reach the bridge"
        return
      }

      // res.ok but not paired yet ("waiting_for_button") — keep polling.
      await sleep(PAIR_POLL_INTERVAL_MS)
    }

    status.value = "error"
    errorMessage.value = "Didn't detect a button press — press the button again and retry."
  }

  async function waitForRestart(): Promise<void> {
    status.value = "restarting"
    // The server is restarting itself after pairing (see internal/handler
    // setup.go) — poll until it responds again rather than a fixed delay.
    for (;;) {
      try {
        const ok = await checkStatus()
        if (ok) return
      } catch {
        // still restarting; keep polling
      }
      await sleep(RESTART_POLL_INTERVAL_MS)
    }
  }

  return {
    configured,
    hueBridgeHost,
    status,
    errorMessage,
    checkStatus,
    startPairing,
    waitForRestart,
  }
})
