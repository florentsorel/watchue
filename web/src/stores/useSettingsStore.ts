import { defineStore } from "pinia"
import { ref } from "vue"

export const useSettingsStore = defineStore("settings", () => {
  const telegramEnabled = ref(true)
  const telegramConfigured = ref(false)
  const hueBridgeHost = ref("")
  const bridgeOnline = ref(true)
  const version = ref("")
  const loading = ref(false)

  async function load(): Promise<void> {
    loading.value = true
    try {
      const res = await fetch("/api/settings")
      if (res.ok) {
        const data = await res.json()
        telegramEnabled.value = data.telegram_enabled
        telegramConfigured.value = data.telegram_configured
        hueBridgeHost.value = data.hue_bridge_host
        bridgeOnline.value = data.bridge_online
        version.value = data.version
      }
    } finally {
      loading.value = false
    }
  }

  async function setTelegramEnabled(enabled: boolean): Promise<void> {
    const res = await fetch("/api/settings/telegram-enabled", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ enabled }),
    })
    if (!res.ok) {
      const data = await res.json().catch(() => ({}))
      throw new Error(data.error ?? "Failed to update setting")
    }
    telegramEnabled.value = enabled
  }

  return {
    telegramEnabled,
    telegramConfigured,
    hueBridgeHost,
    bridgeOnline,
    version,
    loading,
    load,
    setTelegramEnabled,
  }
})
