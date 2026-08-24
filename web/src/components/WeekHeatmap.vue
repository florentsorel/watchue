<template>
  <div :class="{ 'opacity-60': dimmed }">
    <div class="flex gap-2">
      <div class="font-mono-ui flex flex-col justify-between pt-px text-wq-10 text-wq-faint">
        <span v-for="day in WEEKDAY_LABELS" :key="day" class="leading-none">{{ day }}</span>
      </div>
      <div class="min-w-0 flex-1">
        <div v-for="(row, r) in grid" :key="r" class="grid grid-cols-wq-hours gap-px pb-px">
          <div
            v-for="(share, h) in row"
            :key="h"
            class="aspect-square rounded-wq-3"
            :style="{ background: cellColor(share) }"
            :title="cellTitle(r, h, share)"
          />
        </div>
        <!-- Laid out on the same 24-column grid as the cells rather than spaced
             apart, so each label sits under the hour it names. -->
        <div class="font-mono-ui mt-1.5 grid grid-cols-wq-hours gap-px text-wq-10 text-wq-faint">
          <span v-for="h in 24" :key="h" class="text-center">
            {{ (h - 1) % 6 === 0 ? h - 1 : "" }}
          </span>
        </div>
      </div>
    </div>

    <div class="mt-3 flex items-center justify-end gap-1.5">
      <span class="text-wq-10 text-wq-faint">Never on</span>
      <span
        v-for="(color, i) in legendSteps"
        :key="i"
        class="h-2.5 w-2.5 rounded-wq-3"
        :style="{ background: color }"
      />
      <span class="text-wq-10 text-wq-faint">Always on</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { WEEKDAY_LABELS } from "@/utils/insights"
import { useChartTheme } from "@/composables/useChartTheme"
import { useUiStore } from "@/stores/useUiStore"

withDefaults(
  defineProps<{
    /** 7 rows (Monday first) × 24 columns of 0..1. */
    grid: number[][]
    dimmed?: boolean
  }>(),
  { dimmed: false }
)

const theme = useChartTheme()
const ui = useUiStore()

// Re-read on theme change, same reason as useChartTheme: an inline style needs a
// literal color, and --wq-panel-2 is not one of the chart tokens.
const emptyColor = computed(() => {
  void ui.theme
  return getComputedStyle(document.documentElement).getPropertyValue("--wq-panel-2").trim()
})

const legendSteps = computed(() => [emptyColor.value, ...theme.value.heat])

// Bucketed into the ramp's five steps rather than interpolated: a continuous
// gradient gives the reader no legend swatch to match a cell against.
function cellColor(share: number): string {
  if (share <= 0.001) return emptyColor.value
  const steps = theme.value.heat
  return steps[Math.min(steps.length - 1, Math.floor(share * steps.length))]
}

function cellTitle(row: number, hour: number, share: number): string {
  return `${WEEKDAY_LABELS[row]} ${String(hour).padStart(2, "0")}:00 — on ${Math.round(share * 100)}% of the time`
}
</script>
