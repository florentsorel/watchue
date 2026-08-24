<template>
  <div class="h-wq-220" :class="{ 'opacity-60': dimmed }">
    <Line :data="data" :options="options" />
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { Line } from "vue-chartjs"
import {
  CategoryScale,
  Chart,
  Legend,
  LineController,
  LineElement,
  LinearScale,
  PointElement,
  Tooltip,
  type ChartOptions,
  type TooltipItem,
} from "chart.js"
import { useChartTheme } from "@/composables/useChartTheme"
import { formatAnchoredHour, formatSwitchTooltip, type NightPoint } from "@/utils/insights"

Chart.register(LineController, LineElement, PointElement, CategoryScale, LinearScale, Legend, Tooltip)

export interface SwitchSeries {
  label: string
  color: string
  points: Array<NightPoint | null>
}

const props = withDefaults(
  defineProps<{
    labels: string[]
    series: SwitchSeries[]
    /** Hour a column opens at — the Y axis counts from here. */
    anchorHour: number
    dimmed?: boolean
  }>(),
  { dimmed: false }
)

const theme = useChartTheme()

const data = computed(() => ({
  labels: props.labels,
  datasets: props.series.map((s) => ({
    label: s.label,
    // Plain values: the switch's real timestamp and count are looked up from
    // `series` by index in the tooltip rather than smuggled into the datum,
    // which chart.js types as a bare number or {x, y}.
    data: s.points.map((p) => p?.value ?? null),
    borderColor: s.color,
    backgroundColor: s.color,
    borderWidth: 2,
    pointRadius: 4,
    pointHoverRadius: 6,
    // A ring in the surface color keeps a dot legible where two curves cross.
    pointBorderColor: theme.value.surface,
    pointBorderWidth: 2,
    pointHitRadius: 14,
    tension: 0,
    // A day with no switch is a real gap, not a straight line drawn through it.
    spanGaps: false,
  })),
}))

function pointAt(ctx: TooltipItem<"line">): NightPoint | null {
  return props.series[ctx.datasetIndex]?.points[ctx.dataIndex] ?? null
}

const options = computed<ChartOptions<"line">>(() => ({
  responsive: true,
  maintainAspectRatio: false,
  animation: false,
  layout: { padding: { top: 4, right: 8 } },
  scales: {
    x: {
      border: { display: false },
      grid: { display: false },
      ticks: {
        color: theme.value.tick,
        font: { size: 11 },
        padding: 6,
        maxRotation: 0,
        autoSkipPadding: 16,
      },
    },
    y: {
      // A full 24 hours always, so the curve's height means the same thing
      // whatever range is selected — an axis fitted to the data would make a
      // quiet week look as spread out as a chaotic one.
      min: 0,
      max: 24,
      border: { display: false },
      grid: { color: theme.value.grid, drawTicks: false },
      afterBuildTicks: (axis) => {
        axis.ticks = [0, 3, 6, 9, 12, 15, 18, 21, 24].map((value) => ({ value }))
      },
      ticks: {
        color: theme.value.tick,
        font: { size: 11 },
        padding: 8,
        autoSkip: false,
        callback: (v: string | number) => formatAnchoredHour(Number(v), props.anchorHour),
      },
    },
  },
  plugins: {
    legend: {
      display: props.series.length > 1,
      position: "top",
      align: "end",
      labels: {
        color: theme.value.tick,
        boxWidth: 18,
        boxHeight: 2,
        font: { size: 11 },
        usePointStyle: true,
        pointStyle: "line",
      },
    },
    tooltip: {
      backgroundColor: theme.value.tooltipBg,
      titleColor: theme.value.tooltipText,
      bodyColor: theme.value.tooltipText,
      padding: 10,
      cornerRadius: 8,
      displayColors: props.series.length > 1,
      // Line keys rather than filled boxes: at tooltip density a solid swatch is
      // data-weight ink doing a label's job.
      usePointStyle: true,
      callbacks: {
        // The column, not one series' timestamp — every series hovered at this
        // x carries its own clock time on its own row below.
        title: (items: TooltipItem<"line">[]) => props.labels[items[0].dataIndex] ?? "",
        // A series with nothing that day would otherwise render a blank row.
        label: (ctx: TooltipItem<"line">) => {
          const point = pointAt(ctx)
          return point ? formatSwitchTooltip(point, ctx.dataset.label ?? "") : ""
        },
      },
      filter: (ctx: TooltipItem<"line">) => pointAt(ctx) !== null,
    },
  },
  interaction: { mode: "index", intersect: false },
}))
</script>
