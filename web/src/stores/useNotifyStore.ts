import { defineStore } from "pinia"
import { reactive, ref } from "vue"

export type NotifyProvider = "telegram" | "discord"
export type NotifyTestStatus = "idle" | "testing" | "success" | "error"

export interface NotifyConfig {
  provider: NotifyProvider
  telegram_bot_token?: string
  telegram_chat_id?: string
  discord_webhook_url?: string
}

interface ProviderStatus {
  configured: boolean
}

// Backs the standalone /provider page — reachable any time, since the active
// notifier is hot-swappable server-side (see internal/handler.NotifierStore).
export const useNotifyStore = defineStore("notify", () => {
  const testStatus = ref<NotifyTestStatus>("idle")
  const testError = ref("")

  const activeProvider = ref("")
  const status = reactive<Record<NotifyProvider, ProviderStatus>>({
    telegram: { configured: false },
    discord: { configured: false },
  })
  // True when a provider is configured via env — PostNotify/PostNotifyActivate 409 then.
  const envLocked = ref(false)

  function resetTest(): void {
    testStatus.value = "idle"
    testError.value = ""
  }

  async function fetchStatus(): Promise<void> {
    const res = await fetch("/api/notify")
    if (!res.ok) {
      throw new Error("Failed to load notification provider status")
    }
    const data = await res.json()
    activeProvider.value = data.active_provider
    envLocked.value = data.env_locked
    status.telegram.configured = data.telegram.configured
    status.discord.configured = data.discord.configured
  }

  async function test(config: NotifyConfig): Promise<void> {
    testStatus.value = "testing"
    testError.value = ""
    const res = await fetch("/api/notify/test", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(config),
    })
    const data = await res.json().catch(() => ({}))
    if (res.ok && data.ok) {
      testStatus.value = "success"
      return
    }
    testStatus.value = "error"
    testError.value = data.error ?? "Test notification failed"
  }

  async function save(config: NotifyConfig): Promise<void> {
    const res = await fetch("/api/notify", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(config),
    })
    if (!res.ok) {
      const data = await res.json().catch(() => ({}))
      throw new Error(data.error ?? "Failed to save notification settings")
    }
  }

  // Switches to an already-configured provider, reusing its stored credentials.
  async function activate(provider: NotifyProvider): Promise<void> {
    const res = await fetch("/api/notify/activate", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ provider }),
    })
    if (!res.ok) {
      const data = await res.json().catch(() => ({}))
      throw new Error(data.error ?? "Failed to activate provider")
    }
  }

  return {
    testStatus,
    testError,
    activeProvider,
    status,
    envLocked,
    resetTest,
    fetchStatus,
    test,
    save,
    activate,
  }
})
