import { defineStore } from "pinia"
import { ref } from "vue"

export const useSettingsStore = defineStore("settings", () => {
  const notifyEnabled = ref(true)
  const notifyConfigured = ref(false)
  const notifyProvider = ref("")
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
        notifyEnabled.value = data.notify_enabled
        notifyConfigured.value = data.notify_configured
        notifyProvider.value = data.notify_provider
        hueBridgeHost.value = data.hue_bridge_host
        bridgeOnline.value = data.bridge_online
        version.value = data.version
      }
    } finally {
      loading.value = false
    }
  }

  async function setNotifyEnabled(enabled: boolean): Promise<void> {
    const res = await fetch("/api/settings/notify-enabled", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ enabled }),
    })
    if (!res.ok) {
      const data = await res.json().catch(() => ({}))
      throw new Error(data.error ?? "Failed to update setting")
    }
    notifyEnabled.value = enabled
  }

  return {
    notifyEnabled,
    notifyConfigured,
    notifyProvider,
    hueBridgeHost,
    bridgeOnline,
    version,
    loading,
    load,
    setNotifyEnabled,
  }
})
