<template>
  <section class="py-wq-34">
    <SectionHeading
      kicker="Now watching"
      title="Watched resources"
      subtitle="Zones, rooms and lights you're tracking. Mute a resource to keep recording its history without a ping."
    >
      <SegmentedControl
        :model-value="ui.layout"
        :options="[
          { label: 'Glow', value: 'glow' },
          { label: 'Compact', value: 'compact' },
          { label: 'Wall', value: 'wall' },
        ]"
        @update:model-value="ui.setLayout($event as ResourceLayout)"
      />
    </SectionHeading>

    <BridgeOfflineBanner
      v-if="!settings.bridgeOnline"
      message="Can't reach the Hue bridge right now — on/off states below may be stale."
    />
    <div v-if="loading" class="wq-card grid place-items-center p-12">
      <AppIcon name="spinner" :size="22" class="animate-spin text-wq-faint" />
    </div>
    <div v-else-if="items.length === 0" class="wq-card p-9 text-center text-sm text-wq-faint">
      Nothing watched yet — pick a zone, room, or light below.
    </div>
    <div v-else :class="gridClass">
      <ResourceCard
        v-for="item in items"
        :key="item.id"
        v-bind="item"
        :layout="ui.layout"
        @toggle-mute="watchedStore.setNotify(item.id, !item.notify)"
        @unwatch="watchedStore.unwatch(item.id)"
      />
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from "vue"
import SectionHeading from "@/components/SectionHeading.vue"
import SegmentedControl from "@/components/SegmentedControl.vue"
import ResourceCard from "@/components/ResourceCard.vue"
import AppIcon from "@/components/AppIcon.vue"
import BridgeOfflineBanner from "@/components/BridgeOfflineBanner.vue"
import { useBridgeStore } from "@/stores/useBridgeStore"
import { useWatchedStore } from "@/stores/useWatchedStore"
import { useSettingsStore } from "@/stores/useSettingsStore"
import { useUiStore, type ResourceLayout } from "@/stores/useUiStore"
import { roomIcon, zoneIcon, lightIcon } from "@/utils/resourceIcon"

const bridgeStore = useBridgeStore()
const watchedStore = useWatchedStore()
const settings = useSettingsStore()
const ui = useUiStore()

const loading = computed(() => bridgeStore.loading || watchedStore.loading)

const items = computed(() =>
  watchedStore.items.map((w) => {
    if (w.type === "zone") {
      const z = bridgeStore.zones.find((z) => z.id === w.id)
      return {
        id: w.id,
        name: w.name,
        type: "Zone",
        icon: zoneIcon(z?.archetype),
        on: z?.on ?? false,
        notify: w.notify,
        meta: "Grouped light",
      }
    }
    if (w.type === "room") {
      const r = bridgeStore.rooms.find((r) => r.id === w.id)
      const total = r?.lights.length ?? 0
      const on = r?.lights.filter((l) => l.on).length ?? 0
      return {
        id: w.id,
        name: w.name,
        type: "Room",
        icon: roomIcon(r?.archetype, w.name),
        on: r?.on ?? false,
        notify: w.notify,
        meta: `${total} light${total === 1 ? "" : "s"} · ${on} on`,
      }
    }
    const room = bridgeStore.rooms.find((r) => r.lights.some((l) => l.id === w.id))
    const light = room?.lights.find((l) => l.id === w.id)
    return {
      id: w.id,
      name: w.name,
      type: "Light",
      icon: lightIcon(light?.archetype, w.name),
      on: light?.on ?? false,
      notify: w.notify,
      meta: room?.name ?? "Light",
    }
  })
)

const gridClass = computed(() => {
  if (ui.layout === "compact") return "wq-card overflow-hidden"
  if (ui.layout === "wall") return "grid grid-cols-1 gap-4 sm:grid-cols-2"
  return "grid grid-cols-1 gap-3.5 sm:grid-cols-2 lg:grid-cols-3"
})
</script>
