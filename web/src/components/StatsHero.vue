<template>
  <div class="grid grid-cols-1 gap-3.5 md:grid-cols-wq-stats">
    <div class="wq-card relative overflow-hidden bg-gradient-to-br from-wq-panel to-wq-panel-2 p-5">
      <div
        class="pointer-events-none absolute -top-8 -right-8 h-wq-120 w-wq-120 rounded-full opacity-70"
        style="
          background: radial-gradient(
            circle,
            color-mix(in srgb, var(--wq-accent) 42%, transparent),
            transparent 68%
          );
        "
      />
      <div class="flex items-center gap-2 text-xs font-semibold text-wq-muted">
        <AppIcon name="signal" :size="16" class="text-wq-accent-2" />
        Watched &amp; on right now
      </div>
      <div class="mt-3.5 text-wq-40 leading-none font-extrabold tracking-tight">
        {{ onCount
        }}<span class="ml-1 text-xl font-semibold text-wq-faint">
          / {{ watchedStore.items.length }}</span
        >
      </div>
      <div class="mt-2 text-xs text-wq-faint">
        of your watched resources are currently
        <span class="font-semibold text-wq-accent-2">lit</span>. Every change is recorded.
      </div>
    </div>

    <div class="wq-card p-5">
      <div class="flex items-center gap-2 text-xs font-semibold text-wq-muted">
        <AppIcon name="clock" :size="16" />
        Changes today
      </div>
      <div class="mt-3.5 text-wq-40 leading-none font-extrabold tracking-tight">
        {{ changesToday.length }}
      </div>
      <div class="mt-2 text-xs text-wq-faint">
        <span class="font-semibold text-wq-accent-2">{{ sentCount }} sent</span> ·
        {{ suppressedCount }} muted
      </div>
    </div>

    <div class="wq-card p-5">
      <div class="flex items-center gap-2 text-xs font-semibold text-wq-muted">
        <AppIcon name="bell" :size="16" />
        Notifications
      </div>
      <div
        class="mt-5 text-wq-26 font-extrabold"
        :class="settingsStore.notifyEnabled ? providerTextClass : 'text-wq-faint'"
      >
        {{ settingsStore.notifyEnabled ? "Enabled" : "Paused" }}
      </div>
      <div class="mt-2 text-xs text-wq-faint">
        {{
          settingsStore.notifyEnabled
            ? "Watched changes ping your chat instantly."
            : "Recording continues — no messages are sent."
        }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue"
import AppIcon from "@/components/AppIcon.vue"
import { useWatchedStore } from "@/stores/useWatchedStore"
import { useEventsStore } from "@/stores/useEventsStore"
import { useSettingsStore } from "@/stores/useSettingsStore"
import { useBridgeStore } from "@/stores/useBridgeStore"

const watchedStore = useWatchedStore()
const eventsStore = useEventsStore()
const settingsStore = useSettingsStore()
const bridgeStore = useBridgeStore()

const onCount = computed(() => {
  const onById = new Map<string, boolean>()
  for (const z of bridgeStore.zones) onById.set(z.id, z.on)
  for (const r of bridgeStore.rooms) {
    onById.set(r.id, r.on)
    for (const l of r.lights) onById.set(l.id, l.on)
  }
  return watchedStore.items.filter((w) => onById.get(w.id)).length
})

const todayUtc = new Date().toISOString().slice(0, 10)
const changesToday = computed(() =>
  eventsStore.items.filter((e) => e.created_at.startsWith(todayUtc))
)
const sentCount = computed(() => changesToday.value.filter((e) => e.outcome === "sent").length)
const suppressedCount = computed(() => changesToday.value.length - sentCount.value)
const providerTextClass = computed(() =>
  settingsStore.notifyProvider === "discord" ? "text-wq-discord" : "text-wq-tg"
)
</script>
