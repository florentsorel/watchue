<template>
  <section class="py-wq-34">
    <SectionHeading
      kicker="History"
      title="Recent changes"
      subtitle="Every on/off change for a watched resource is recorded here — even when notifications are muted."
    />
    <div class="wq-card overflow-hidden">
      <div class="flex flex-wrap items-center gap-3 border-b border-wq-border-2 px-4 py-3">
        <SegmentedControl
          :model-value="ui.historyFilter"
          :options="[
            { label: 'All', value: 'all' },
            { label: 'Turned on', value: 'on' },
            { label: 'Turned off', value: 'off' },
            { label: 'Muted', value: 'muted' },
          ]"
          @update:model-value="ui.setHistoryFilter($event as HistoryFilter)"
        />
        <div class="flex-1" />
        <div class="font-mono-ui hidden text-wq-11 font-semibold text-wq-faint sm:block">
          {{ filtered.length }} events
        </div>
      </div>
      <div v-if="eventsStore.loading" class="grid place-items-center py-10">
        <AppIcon name="spinner" :size="20" class="animate-spin text-wq-faint" />
      </div>
      <template v-else>
        <ScrollableList v-if="filtered.length > 0" :items-displayed="20">
          <HistoryEventRow
            v-for="e in filtered"
            :key="e.id"
            :name="e.name"
            :type-label="typeLabel(e.resource_type)"
            :on="e.on"
            :outcome="e.outcome"
            :time="formatRelativeTime(e.created_at)"
          />
        </ScrollableList>
        <div v-else class="p-9 text-center text-sm text-wq-faint">
          No changes match this filter yet.
        </div>
      </template>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from "vue"
import SectionHeading from "@/components/SectionHeading.vue"
import SegmentedControl from "@/components/SegmentedControl.vue"
import HistoryEventRow from "@/components/HistoryEventRow.vue"
import AppIcon from "@/components/AppIcon.vue"
import ScrollableList from "@/components/ScrollableList.vue"
import { useEventsStore } from "@/stores/useEventsStore"
import { useUiStore, type HistoryFilter } from "@/stores/useUiStore"
import { formatRelativeTime } from "@/utils/time"

const eventsStore = useEventsStore()
const ui = useUiStore()

function typeLabel(t: string): string {
  return t === "zone" ? "Zone" : t === "room" ? "Room" : "Light"
}

const filtered = computed(() => {
  switch (ui.historyFilter) {
    case "on":
      return eventsStore.items.filter((e) => e.on)
    case "off":
      return eventsStore.items.filter((e) => !e.on)
    case "muted":
      return eventsStore.items.filter((e) => e.outcome !== "sent")
    default:
      return eventsStore.items
  }
})
</script>
