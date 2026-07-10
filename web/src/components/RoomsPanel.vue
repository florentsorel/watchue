<template>
  <div class="wq-card overflow-hidden">
    <div class="flex items-center gap-2.5 border-b border-wq-border-2 px-wq-17 py-wq-15">
      <AppIcon name="room-living" :size="18" class="text-wq-accent-2" />
      <div class="text-sm font-semibold">Rooms</div>
      <div class="flex-1" />
      <div
        class="font-mono-ui rounded-md bg-wq-panel-2 px-2 py-0.5 text-wq-11 font-semibold text-wq-faint"
      >
        {{ bridgeStore.rooms.length }}
      </div>
    </div>
    <div v-if="bridgeStore.loading" class="grid place-items-center py-10">
      <AppIcon name="spinner" :size="20" class="animate-spin text-wq-faint" />
    </div>
    <ScrollableList v-else :items-displayed="10">
      <div
        v-for="r in bridgeStore.rooms"
        :key="r.id"
        class="border-b border-wq-border-2 last:border-b-0"
      >
        <BridgeResourceRow
          :name="r.name"
          :icon="roomIcon(r.archetype, r.name)"
          :on="r.on"
          :watched="watchedStore.isWatched(r.id)"
          :meta="roomMeta(r)"
          expandable
          :expanded="expanded.has(r.id)"
          @toggle-watch="toggleWatch(r.id)"
          @toggle-expand="toggleExpand(r.id)"
        />
        <template v-if="expanded.has(r.id)">
          <BridgeResourceRow
            v-for="l in r.lights"
            :key="l.id"
            :name="l.name"
            :icon="lightIcon(l.archetype, l.name)"
            :on="l.on"
            :watched="watchedStore.isWatched(l.id)"
            indent
            @toggle-watch="toggleWatch(l.id)"
          />
        </template>
      </div>
    </ScrollableList>
  </div>
</template>

<script setup lang="ts">
import { reactive } from "vue"
import AppIcon from "@/components/AppIcon.vue"
import BridgeResourceRow from "@/components/BridgeResourceRow.vue"
import ScrollableList from "@/components/ScrollableList.vue"
import { useBridgeStore, type BridgeGroup } from "@/stores/useBridgeStore"
import { useWatchedStore } from "@/stores/useWatchedStore"
import { useToggleWatch } from "@/composables/useToggleWatch"
import { roomIcon, lightIcon } from "@/utils/resourceIcon"

const bridgeStore = useBridgeStore()
const watchedStore = useWatchedStore()
const toggleWatch = useToggleWatch()

const expanded = reactive(new Set<string>())

function toggleExpand(id: string) {
  if (expanded.has(id)) expanded.delete(id)
  else expanded.add(id)
}

function roomMeta(r: BridgeGroup): string {
  const on = r.lights.filter((l) => l.on).length
  return `${r.lights.length} light${r.lights.length === 1 ? "" : "s"} · ${on} on`
}
</script>
