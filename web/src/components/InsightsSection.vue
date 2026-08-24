<template>
  <section class="py-wq-34">
    <SectionHeading
      kicker="Insights"
      title="Usage over time"
      subtitle="Derived from the same recorded changes as the history below — how often your watched resources were switched, and how long they stayed on."
    >
      <!-- One filter row, above everything it scopes: every card below reads
           the same slice, so the numbers always agree. -->
      <SegmentedControl
        :model-value="String(statsStore.days)"
        :options="rangeOptions"
        @update:model-value="statsStore.load(Number($event))"
      />
    </SectionHeading>

    <div v-if="statsStore.loading && !statsStore.loaded" class="wq-card grid place-items-center py-wq-74">
      <AppIcon name="spinner" :size="20" class="animate-spin text-wq-faint" />
    </div>

    <div v-else-if="statsStore.sessions.length === 0" class="wq-card p-9 text-center">
      <p class="text-sm text-wq-muted">Nothing recorded in this range yet.</p>
      <p class="mt-1 text-wq-13 text-wq-faint">
        Watch a zone, room or light and its on/off changes will start showing up here.
      </p>
    </div>

    <template v-else>
      <div class="grid grid-cols-2 gap-3.5 lg:grid-cols-4">
        <InsightsStat
          icon="power"
          label="Switched on"
          :value="String(totals.turnedOn)"
          :hint="`${totals.turnedOff} switched off`"
        />
        <InsightsStat
          icon="bulb"
          label="Time on"
          :value="formatDuration(totals.onMs)"
          :hint="onTimeHint"
        />
        <InsightsStat
          icon="clock"
          label="Avg time on"
          :value="totals.turnedOn > 0 ? formatDuration(totals.onMs / totals.turnedOn) : '—'"
          hint="per switch-on"
        />
        <InsightsStat
          icon="signal"
          label="Busiest hour"
          :value="busiest.label"
          :hint="busiest.hint"
        />
      </div>

      <div class="mt-3.5 grid grid-cols-1 gap-3.5 lg:grid-cols-2">
        <InsightsCard
          title="Switched on and off"
          :hint="`Changes recorded per ${bucketNoun}`"
          :value="`${totals.turnedOn + totals.turnedOff} changes`"
        >
          <InsightsBarChart
            :labels="labels"
            :series="changeSeries"
            :format="formatCount"
            stacked
            :dimmed="statsStore.loading"
          />
        </InsightsCard>

        <InsightsCard
          title="Time on"
          :hint="`Hours lit per ${bucketNoun}`"
          :value="formatDuration(totals.onMs)"
        >
          <InsightsBarChart
            :labels="labels"
            :series="onTimeSeries"
            :format="formatHours"
            :suggested-max="bucketSize === 'day' ? 24 : undefined"
            :dimmed="statsStore.loading"
          />
        </InsightsCard>
      </div>

      <div class="mt-3.5">
        <InsightsCard
          title="When the lights go off"
          :hint="anchorHour === 12
            ? 'One column per night, from noon to noon. Each point is the last switch-off of that night — the higher it sits, the later it happened.'
            : 'One column per calendar day. Each point is the last switch-off of that day.'"
        >
          <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
            <ResourcePicker
              :resources="resources"
              :model-value="selectedIds"
              :max="theme.series.length"
              :color-of="colorOf"
              @update:model-value="selectedIds = $event"
            />
            <SegmentedControl
              :model-value="String(anchorHour)"
              :options="[
                { label: 'Night', value: '12' },
                { label: 'Calendar day', value: '0' },
              ]"
              @update:model-value="anchorHour = Number($event)"
            />
          </div>
          <NightlySwitchChart
            :labels="switchLabels"
            :series="switchSeries"
            :anchor-hour="anchorHour"
            :dimmed="statsStore.loading"
          />
        </InsightsCard>
      </div>

      <div class="mt-3.5">
        <InsightsCard
          title="Weekly rhythm"
          hint="Share of each weekday hour spent on, averaged across the range"
        >
          <WeekHeatmap :grid="heatmap" :dimmed="statsStore.loading" />
        </InsightsCard>
      </div>

      <div class="mt-3.5 grid grid-cols-1 gap-3.5 lg:grid-cols-2">
        <!-- Two small multiples rather than one chart with two scales: hours and
             counts don't share an axis, and a dual-axis plot would invent a
             correlation between them. -->
        <InsightsCard title="Time on by resource" :hint="resourceHint">
          <InsightsBarChart
            :labels="resourceLabels"
            :series="resourceOnTimeSeries"
            :format="formatHours"
            horizontal
            :dimmed="statsStore.loading"
          />
        </InsightsCard>

        <InsightsCard title="Changes by resource" :hint="resourceHint">
          <InsightsBarChart
            :labels="resourceLabels"
            :series="resourceChangeSeries"
            :format="formatCount"
            horizontal
            :dimmed="statsStore.loading"
          />
        </InsightsCard>
      </div>

      <div class="mt-3.5">
        <button
          type="button"
          class="text-wq-12-5 font-semibold text-wq-muted transition-colors hover:text-wq-text"
          @click="showTable = !showTable"
        >
          {{ showTable ? "Hide" : "Show" }} data table
        </button>
        <InsightsTable
          v-if="showTable"
          class="mt-3"
          :buckets="buckets"
          :bucket-noun="bucketNoun"
          :resources="resources"
          :heatmap="heatmap"
        />
      </div>
    </template>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue"
import SectionHeading from "@/components/SectionHeading.vue"
import SegmentedControl from "@/components/SegmentedControl.vue"
import AppIcon from "@/components/AppIcon.vue"
import InsightsCard from "@/components/InsightsCard.vue"
import InsightsStat from "@/components/InsightsStat.vue"
import InsightsBarChart from "@/components/InsightsBarChart.vue"
import NightlySwitchChart from "@/components/NightlySwitchChart.vue"
import ResourcePicker from "@/components/ResourcePicker.vue"
import InsightsTable from "@/components/InsightsTable.vue"
import WeekHeatmap from "@/components/WeekHeatmap.vue"
import { useStatsStore, STATS_RANGES } from "@/stores/useStatsStore"
import { useChartTheme } from "@/composables/useChartTheme"
import {
  bucketSizeFor,
  bucketize,
  formatDuration,
  nightlySwitches,
  perResource,
  toHours,
  weeklyHeatmap,
  WEEKDAY_LABELS,
} from "@/utils/insights"

const statsStore = useStatsStore()
const theme = useChartTheme()
const showTable = ref(false)

const rangeOptions = STATS_RANGES.map((r) => ({ label: r.label, value: String(r.value) }))

const buckets = computed(() => bucketize(statsStore.sessions, statsStore.days))
const bucketSize = computed(() => bucketSizeFor(statsStore.days))
const bucketNoun = computed(() => (bucketSize.value === "week" ? "week" : "day"))
const labels = computed(() => buckets.value.map((b) => b.label))

const changeSeries = computed(() => [
  { label: "Switched on", color: theme.value.on, values: buckets.value.map((b) => b.turnedOn) },
  { label: "Switched off", color: theme.value.off, values: buckets.value.map((b) => b.turnedOff) },
])

const onTimeSeries = computed(() => [
  { label: "Time on", color: theme.value.on, values: buckets.value.map((b) => toHours(b.onMs)) },
])

const heatmap = computed(() => weeklyHeatmap(statsStore.sessions, statsStore.days))

const resources = computed(() => perResource(statsStore.sessions, statsStore.days))
const resourceLabels = computed(() => resources.value.map((r) => r.name))
const resourceHint = computed(
  () => `${resources.value.length} ${resources.value.length === 1 ? "resource" : "resources"} with recorded changes`
)
// Nominal categories, so one hue for every bar: coloring each by its own value
// would re-encode what the bar length already shows.
const resourceOnTimeSeries = computed(() => [
  { label: "Time on", color: theme.value.on, values: resources.value.map((r) => toHours(r.onMs)) },
])
const resourceChangeSeries = computed(() => [
  { label: "Changes", color: theme.value.on, values: resources.value.map((r) => r.changes) },
])

const totals = computed(() => ({
  turnedOn: buckets.value.reduce((sum, b) => sum + b.turnedOn, 0),
  turnedOff: buckets.value.reduce((sum, b) => sum + b.turnedOff, 0),
  onMs: buckets.value.reduce((sum, b) => sum + b.onMs, 0),
}))

// Measured against the time the range actually covers, not its nominal length:
// the first bucket starts at a midnight that may predate the oldest event.
const onShare = computed(() => {
  const first = buckets.value[0]
  if (!first || resources.value.length !== 1) return null
  const elapsed = Math.max(1, Date.now() - first.start.getTime())
  return Math.min(1, totals.value.onMs / elapsed)
})

// Summing several resources' on-time can exceed the range's own length, so the
// share only means anything when a single resource is in play.
const onTimeHint = computed(() =>
  onShare.value === null
    ? `across ${resources.value.length} resources`
    : `${Math.round(onShare.value * 100)}% of the range`
)

const busiest = computed(() => {
  const byHour = Array.from({ length: 24 }, (_, h) => {
    const shares = heatmap.value.map((row) => row[h])
    return shares.reduce((a, b) => a + b, 0) / shares.length
  })
  const peak = byHour.indexOf(Math.max(...byHour))
  if (byHour[peak] <= 0) return { label: "—", hint: "nothing recorded" }

  const byDay = heatmap.value.map((row) => row.reduce((a, b) => a + b, 0))
  const peakDay = byDay.indexOf(Math.max(...byDay))
  return {
    label: `${String(peak).padStart(2, "0")}h`,
    hint: `${WEEKDAY_LABELS[peakDay]} is the busiest day`,
  }
})

// --- When the lights go off ------------------------------------------------

// Noon by default, so a night is one column: a switch-off at 01:17 belongs to
// the evening before it, and plots above that evening's 23:05 rather than
// dropping to the bottom of the next column.
const anchorHour = ref(12)
const selectedIds = ref<string[]>([])

// Color follows the resource, not its position in the selection: dropping one
// resource must not repaint the others. A slot is held until its resource is
// deselected, then freed for the next one picked.
const colorSlots = ref<Record<string, number>>({})

watch(
  [selectedIds, resources],
  () => {
    const available = new Set(resources.value.map((r) => r.id))
    const kept = selectedIds.value.filter((id) => available.has(id))
    if (kept.length === 0 && resources.value.length > 0) {
      selectedIds.value = [resources.value[0].id]
      return
    }
    if (kept.length !== selectedIds.value.length) {
      selectedIds.value = kept
      return
    }
    const slots: Record<string, number> = {}
    for (const id of kept) {
      if (colorSlots.value[id] !== undefined) slots[id] = colorSlots.value[id]
    }
    const taken = new Set(Object.values(slots))
    for (const id of kept) {
      if (slots[id] !== undefined) continue
      let slot = 0
      while (taken.has(slot)) slot++
      slots[id] = slot
      taken.add(slot)
    }
    colorSlots.value = slots
  },
  { immediate: true, deep: true }
)

function colorOf(id: string): string {
  return theme.value.series[colorSlots.value[id] ?? 0]
}

const nightly = computed(() =>
  selectedIds.value
    .map((id) => resources.value.find((r) => r.id === id))
    .filter((r) => r !== undefined)
    .map((r) => ({
      resource: r,
      switches: nightlySwitches(
        statsStore.sessions.filter((s) => s.resource_id === r.id),
        statsStore.days,
        anchorHour.value
      ),
    }))
)

const switchLabels = computed(() => nightly.value[0]?.switches.labels ?? [])

// One resource gets both curves — the switch-on says how long the light was up
// before it went off. Several resources would make six lines out of that, so
// there the switch-off alone carries the comparison.
const switchSeries = computed(() => {
  if (nightly.value.length === 1) {
    const { resource, switches } = nightly.value[0]
    const off = colorOf(resource.id)
    // Whichever slot the resource holds, the switch-on curve takes a different
    // one — any two of the three clear the color-blind separation, but a
    // resource sitting in slot 2 would otherwise collide with a fixed choice.
    const on = theme.value.series.find((c) => c !== off) ?? theme.value.series[1]
    return [
      { label: "Switched on", color: on, points: switches.on },
      { label: "Switched off", color: off, points: switches.off },
    ]
  }
  return nightly.value.map(({ resource, switches }) => ({
    label: resource.name,
    color: colorOf(resource.id),
    points: switches.off,
  }))
})

function formatHours(value: number): string {
  return `${Math.round(value)}h`
}

function formatCount(value: number): string {
  return String(Math.round(value))
}
</script>
