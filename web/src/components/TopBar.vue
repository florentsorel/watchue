<template>
  <div class="sticky top-0 z-30 border-b border-wq-border-2 bg-wq-bg/80 backdrop-blur-md">
    <div class="mx-auto flex h-16 max-w-wq-container items-center gap-4 px-wq-22">
      <div class="flex items-center gap-wq-11">
        <div
          class="grid h-wq-38 w-wq-38 place-items-center rounded-wq-11 bg-gradient-to-br from-wq-accent to-wq-accent-2 text-white shadow-wq-logo"
        >
          <AppIcon name="bulb" :size="18" />
        </div>
        <div>
          <div class="text-wq-17 font-bold tracking-tight">Watchue</div>
          <div class="font-mono-ui mt-0.5 text-wq-11 tracking-wide text-wq-faint">hue · watch</div>
        </div>
      </div>

      <div
        class="hidden h-wq-34 items-center gap-wq-7 rounded-wq-9 border border-wq-border bg-wq-panel px-3 text-wq-12-5 font-semibold text-wq-muted sm:inline-flex"
      >
        <span
          class="h-wq-7 w-wq-7 rounded-full"
          :class="settings.bridgeOnline ? 'bg-wq-good shadow-wq-status-ring' : 'bg-wq-faint'"
        />
        {{ settings.bridgeOnline ? "Bridge online" : "Bridge unreachable" }}
      </div>

      <div class="flex-1" />

      <div
        v-if="settings.notifyProvider"
        class="hidden h-wq-34 items-center gap-wq-7 rounded-wq-9 border border-wq-border bg-wq-panel px-3 text-wq-12-5 font-semibold text-wq-muted sm:inline-flex"
      >
        <ProviderIcon :provider="settings.notifyProvider" :size="16" />
        {{ providerLabel }} {{ settings.notifyEnabled ? "on" : "off" }}
      </div>

      <button
        v-if="settings.notifyProvider"
        type="button"
        class="relative h-6 w-wq-42 flex-none rounded-full transition-colors"
        :class="settings.notifyEnabled ? providerToggleClass : 'bg-wq-faint/50'"
        title="Toggle notifications"
        @click="settings.setNotifyEnabled(!settings.notifyEnabled)"
      >
        <span
          class="absolute top-0.5 h-5 w-5 rounded-full bg-white shadow transition-all"
          :class="settings.notifyEnabled ? 'left-wq-20' : 'left-0.5'"
        />
      </button>

      <button
        type="button"
        class="grid h-wq-34 w-wq-34 place-items-center rounded-wq-9 border border-wq-border bg-wq-panel text-wq-muted transition-colors hover:text-wq-text"
        title="Toggle theme"
        @click="ui.toggleTheme"
      >
        <AppIcon :name="ui.theme === 'dark' ? 'sun' : 'moon'" :size="18" />
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue"
import AppIcon from "@/components/AppIcon.vue"
import ProviderIcon from "@/components/ProviderIcon.vue"
import { useSettingsStore } from "@/stores/useSettingsStore"
import { useUiStore } from "@/stores/useUiStore"

const settings = useSettingsStore()
const ui = useUiStore()

const providerLabel = computed(() =>
  settings.notifyProvider === "discord" ? "Discord" : "Telegram"
)
const providerToggleClass = computed(() =>
  settings.notifyProvider === "discord" ? "bg-wq-discord" : "bg-wq-tg"
)
</script>
