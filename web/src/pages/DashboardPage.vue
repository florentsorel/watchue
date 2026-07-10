<template>
  <div class="min-h-screen">
    <TopBar />
    <div class="mx-auto max-w-wq-container px-wq-22">
      <div class="pt-wq-30">
        <StatsHero />
      </div>
      <WatchedResourcesSection />
      <BrowseBridgeSection />
      <HistorySection />
      <SettingsSection />
    </div>
    <div class="flex items-center justify-center gap-2 py-10 text-xs text-wq-faint">
      <AppIcon name="bulb" :size="16" />
      Watchue · self-hosted Hue watcher
      <span v-if="settingsStore.version" class="font-mono-ui text-wq-faint">
        · {{ settingsStore.version }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from "vue"
import TopBar from "@/components/TopBar.vue"
import AppIcon from "@/components/AppIcon.vue"
import StatsHero from "@/components/StatsHero.vue"
import WatchedResourcesSection from "@/components/WatchedResourcesSection.vue"
import BrowseBridgeSection from "@/components/BrowseBridgeSection.vue"
import HistorySection from "@/components/HistorySection.vue"
import SettingsSection from "@/components/SettingsSection.vue"
import { useBridgeStore } from "@/stores/useBridgeStore"
import { useWatchedStore } from "@/stores/useWatchedStore"
import { useEventsStore } from "@/stores/useEventsStore"
import { useSettingsStore } from "@/stores/useSettingsStore"
import { useEventStream } from "@/composables/useEventStream"

const bridgeStore = useBridgeStore()
const watchedStore = useWatchedStore()
const eventsStore = useEventsStore()
const settingsStore = useSettingsStore()

useEventStream()

onMounted(() => {
  bridgeStore.load()
  watchedStore.load()
  eventsStore.load()
  settingsStore.load()
})
</script>
