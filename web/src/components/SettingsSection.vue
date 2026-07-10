<template>
  <section class="py-wq-34">
    <SectionHeading
      kicker="Settings"
      title="Notification channel"
      subtitle="Watchue records history regardless. This controls whether a message is sent."
    />
    <div class="grid grid-cols-1 items-start gap-4 lg:grid-cols-wq-settings">
      <div class="wq-card">
        <div class="flex items-center gap-3.5 border-b border-wq-border-2 px-wq-18 py-4">
          <div class="flex-1">
            <div class="text-sm font-semibold">Notifications</div>
            <div class="mt-0.5 max-w-wq-44ch text-xs text-wq-muted">
              Global switch. When off, changes are still recorded but no message is sent.
            </div>
          </div>
          <button
            type="button"
            class="relative h-6 w-wq-42 flex-none rounded-full transition-colors"
            :class="settings.notifyEnabled ? providerToggleClass : 'bg-wq-faint/50'"
            @click="settings.setNotifyEnabled(!settings.notifyEnabled)"
          >
            <span
              class="absolute top-0.5 h-5 w-5 rounded-full bg-white shadow transition-all"
              :class="settings.notifyEnabled ? 'left-wq-20' : 'left-0.5'"
            />
          </button>
        </div>
        <div class="flex items-center gap-3.5 border-b border-wq-border-2 px-wq-18 py-4">
          <div class="flex-1">
            <div class="text-sm font-semibold">
              {{ providerLabel || "Notification" }} credentials
            </div>
            <div class="mt-0.5 max-w-wq-44ch text-xs text-wq-muted">
              Set via env vars or the {{ settings.notifyConfigured ? "Change" : "Configure" }}
              button — never sent to this page.
            </div>
          </div>
          <ConfigStatus
            :ok="settings.notifyConfigured"
            :label="settings.notifyConfigured ? 'Configured' : 'Not set'"
          />
          <RouterLink
            to="/provider"
            class="rounded-lg border border-wq-border bg-wq-panel-2 px-3 py-2 text-xs font-semibold text-wq-muted transition-colors hover:text-wq-text"
          >
            {{ settings.notifyConfigured ? "Change" : "Configure" }}
          </RouterLink>
        </div>
        <div class="flex items-center gap-3.5 px-wq-18 py-4">
          <div class="flex-1">
            <div class="text-sm font-semibold">Hue bridge</div>
            <div class="mt-0.5 text-xs text-wq-muted">CLIP v2 event stream · app-key paired.</div>
          </div>
          <div
            class="font-mono-ui flex items-center gap-2.5 rounded-lg border border-wq-border bg-wq-panel-2 px-3 py-2 text-xs font-medium text-wq-muted"
          >
            <span>{{ settings.hueBridgeHost || "unknown" }}</span>
            <ConfigStatus
              :ok="settings.bridgeOnline"
              :label="settings.bridgeOnline ? 'Live' : 'Offline'"
            />
          </div>
        </div>
      </div>

      <div
        class="wq-card overflow-hidden p-5"
        :class="
          settings.notifyProvider === 'discord'
            ? 'bg-gradient-to-br from-wq-discord/10 to-transparent'
            : 'bg-gradient-to-br from-wq-tg/10 to-transparent'
        "
      >
        <div class="flex items-center gap-3">
          <div
            class="grid h-wq-46 w-wq-46 flex-none place-items-center rounded-wq-13 text-white"
            :class="providerToggleClass"
            :style="{ boxShadow: providerShadow }"
          >
            <ProviderIcon :provider="settings.notifyProvider" :size="28" />
          </div>
          <div>
            <div class="text-wq-15 font-bold">
              {{
                settings.notifyConfigured
                  ? `Watchue ${providerLabel === "Discord" ? "Webhook" : "Bot"}`
                  : "No provider configured"
              }}
            </div>
            <div class="font-mono-ui text-wq-11-5 text-wq-muted">
              {{ settings.notifyConfigured ? "your configured channel" : "set one up via /setup" }}
            </div>
          </div>
        </div>
        <div class="wq-card mt-4 rounded-tl-wq-3 p-3.5 text-wq-13 leading-relaxed">
          <div class="mb-1 text-xs font-semibold text-wq-accent-2">🔆 Watchue</div>
          <div v-if="settings.notifyProvider === 'discord'">**Office** was turned **on**.</div>
          <div v-else><strong>Office</strong> was turned <strong>on</strong>.</div>
          <div class="font-mono-ui mt-2 text-wq-11 text-wq-faint">Desk Lamp · today at 20:14</div>
        </div>
        <div class="mt-3.5 text-xs text-wq-muted">
          This is how each notification looks. Muted resources and a disabled channel skip the send.
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from "vue"
import SectionHeading from "@/components/SectionHeading.vue"
import ConfigStatus from "@/components/ConfigStatus.vue"
import ProviderIcon from "@/components/ProviderIcon.vue"
import { useSettingsStore } from "@/stores/useSettingsStore"

const settings = useSettingsStore()

const providerLabel = computed(() => {
  if (settings.notifyProvider === "discord") return "Discord"
  if (settings.notifyProvider === "telegram") return "Telegram"
  return ""
})
const providerToggleClass = computed(() =>
  settings.notifyProvider === "discord" ? "bg-wq-discord" : "bg-wq-tg"
)
const providerShadow = computed(() =>
  settings.notifyProvider === "discord" ? "var(--shadow-wq-discord)" : "var(--shadow-wq-telegram)"
)
</script>
