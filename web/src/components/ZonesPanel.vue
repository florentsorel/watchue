<template>
  <div class="wq-card overflow-hidden">
    <div class="flex items-center gap-2.5 border-b border-wq-border-2 px-wq-17 py-wq-15">
      <AppIcon name="zone" :size="18" class="text-wq-accent-2" />
      <div class="text-sm font-semibold">Zones</div>
      <div class="flex-1" />
      <div
        class="font-mono-ui rounded-md bg-wq-panel-2 px-2 py-0.5 text-wq-11 font-semibold text-wq-faint"
      >
        {{ bridgeStore.zones.length }}
      </div>
    </div>
    <div v-if="bridgeStore.loading" class="grid place-items-center py-10">
      <AppIcon name="spinner" :size="20" class="animate-spin text-wq-faint" />
    </div>
    <ScrollableList v-else :items-displayed="10">
      <BridgeResourceRow
        v-for="z in bridgeStore.zones"
        :key="z.id"
        :name="z.name"
        :icon="zoneIcon(z.archetype)"
        :on="z.on"
        :watched="watchedStore.isWatched(z.id)"
        meta="Grouped light"
        @toggle-watch="toggleWatch(z.id)"
      />
    </ScrollableList>
  </div>
</template>

<script setup lang="ts">
import AppIcon from "@/components/AppIcon.vue"
import BridgeResourceRow from "@/components/BridgeResourceRow.vue"
import ScrollableList from "@/components/ScrollableList.vue"
import { useBridgeStore } from "@/stores/useBridgeStore"
import { useWatchedStore } from "@/stores/useWatchedStore"
import { useToggleWatch } from "@/composables/useToggleWatch"
import { zoneIcon } from "@/utils/resourceIcon"

const bridgeStore = useBridgeStore()
const watchedStore = useWatchedStore()
const toggleWatch = useToggleWatch()
</script>
