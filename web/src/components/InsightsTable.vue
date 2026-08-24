<template>
  <div class="wq-card overflow-hidden">
    <div class="flex flex-wrap items-center gap-3 border-b border-wq-border-2 px-4 py-3">
      <SegmentedControl
        :model-value="view"
        :options="[
          { label: `By ${bucketNoun}`, value: 'bucket' },
          { label: 'By resource', value: 'resource' },
          { label: 'By hour', value: 'hour' },
        ]"
        @update:model-value="view = $event as View"
      />
    </div>

    <!-- The chart's readable twin: every plotted value is here as text, so no
         number is reachable only by hovering a mark. -->
    <div class="max-h-wq-220 overflow-y-auto">
      <table class="font-mono-ui w-full text-wq-11-5 tabular-nums">
        <thead class="sticky top-0 bg-wq-panel text-left text-wq-faint">
          <tr>
            <th v-for="h in headers" :key="h" class="px-4 py-2 font-semibold" :class="alignOf(h)">
              {{ h }}
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in rows" :key="row[0]" class="border-t border-wq-border-2">
            <td
              v-for="(cell, i) in row"
              :key="i"
              class="px-4 py-1.5"
              :class="i === 0 ? 'text-wq-text' : 'text-right text-wq-muted'"
            >
              {{ cell }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue"
import SegmentedControl from "@/components/SegmentedControl.vue"
import { formatDuration, WEEKDAY_LABELS, type Bucket, type ResourceTotal } from "@/utils/insights"

type View = "bucket" | "resource" | "hour"

const props = defineProps<{
  buckets: Bucket[]
  bucketNoun: string
  resources: ResourceTotal[]
  heatmap: number[][]
}>()

const view = ref<View>("bucket")

const headers = computed(() => {
  if (view.value === "resource") return ["Resource", "Type", "Time on", "Changes"]
  if (view.value === "hour") return ["Hour", "Share on", "Busiest day"]
  return [props.bucketNoun === "week" ? "Week of" : "Day", "On", "Off", "Time on"]
})

const rows = computed<string[][]>(() => {
  if (view.value === "resource") {
    return props.resources.map((r) => [r.name, r.type, formatDuration(r.onMs), String(r.changes)])
  }
  if (view.value === "hour") {
    return Array.from({ length: 24 }, (_, h) => {
      const shares = props.heatmap.map((row) => row[h])
      const mean = shares.reduce((a, b) => a + b, 0) / shares.length
      const peakDay = shares.indexOf(Math.max(...shares))
      return [
        `${String(h).padStart(2, "0")}:00`,
        `${Math.round(mean * 100)}%`,
        mean > 0 ? WEEKDAY_LABELS[peakDay] : "—",
      ]
    })
  }
  return props.buckets.map((b) => [
    b.label,
    String(b.turnedOn),
    String(b.turnedOff),
    formatDuration(b.onMs),
  ])
})

function alignOf(header: string): string {
  return headers.value.indexOf(header) === 0 ? "" : "text-right"
}
</script>
