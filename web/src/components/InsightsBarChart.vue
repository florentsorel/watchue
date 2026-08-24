<template>
  <div class="h-wq-220" :class="{ 'opacity-60': dimmed }">
    <Bar :data="data" :options="options" />
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { Bar } from "vue-chartjs"
import {
  BarController,
  BarElement,
  CategoryScale,
  Chart,
  LinearScale,
  Legend,
  Tooltip,
  type ChartOptions,
  type TooltipItem,
} from "chart.js"
import { useChartTheme } from "@/composables/useChartTheme"

// Registered once at module scope: chart.js is tree-shaken, so anything not
// registered is silently missing at render time.
Chart.register(BarController, BarElement, CategoryScale, LinearScale, Legend, Tooltip)

export interface Series {
  label: string
  color: string
  values: number[]
}

const props = withDefaults(
  defineProps<{
    labels: string[]
    series: Series[]
    /** Formats a value for the tooltip and the axis ticks. */
    format: (value: number) => string
    horizontal?: boolean
    stacked?: boolean
    suggestedMax?: number
    /** Held at reduced opacity while a refetch is in flight, rather than
     * unmounting into a skeleton — no layout jump. */
    dimmed?: boolean
  }>(),
  { horizontal: false, stacked: false, dimmed: false, suggestedMax: undefined }
)

const theme = useChartTheme()

const data = computed(() => ({
  labels: props.labels,
  datasets: props.series.map((s) => ({
    label: s.label,
    data: s.values,
    backgroundColor: s.color,
    // Rounded at the data end, square at the baseline. A stack's segments get
    // the same treatment so the top of the column reads as one shape.
    borderRadius: 4,
    borderSkipped: false as const,
    // A gap in the surface color, not a stroke, is what separates touching
    // marks — 2px on the stacking axis only.
    borderWidth: props.stacked ? { top: 2 } : 0,
    borderColor: theme.value.surface,
    maxBarThickness: 24,
  })),
}))

const options = computed<ChartOptions<"bar">>(() => {
  const valueAxis = {
    stacked: props.stacked,
    beginAtZero: true,
    suggestedMax: props.suggestedMax,
    border: { display: false },
    grid: { color: theme.value.grid, drawTicks: false },
    ticks: {
      color: theme.value.tick,
      font: { size: 11 },
      padding: 8,
      callback: (v: string | number) => props.format(Number(v)),
    },
  }
  const categoryAxis = {
    stacked: props.stacked,
    border: { display: false },
    grid: { display: false },
    ticks: {
      color: theme.value.tick,
      font: { size: 11 },
      padding: 6,
      autoSkipPadding: 16,
      maxRotation: 0,
    },
  }

  return {
    responsive: true,
    maintainAspectRatio: false,
    indexAxis: props.horizontal ? "y" : "x",
    layout: { padding: { top: 4 } },
    scales: props.horizontal
      ? { x: valueAxis, y: categoryAxis }
      : { x: categoryAxis, y: valueAxis },
    plugins: {
      // A single series needs no legend box — the card title already names it.
      legend: {
        display: props.series.length > 1,
        position: "top",
        align: "end",
        labels: {
          color: theme.value.tick,
          boxWidth: 10,
          boxHeight: 10,
          borderRadius: 3,
          useBorderRadius: true,
          font: { size: 11 },
        },
      },
      tooltip: {
        backgroundColor: theme.value.tooltipBg,
        titleColor: theme.value.tooltipText,
        bodyColor: theme.value.tooltipText,
        padding: 10,
        cornerRadius: 8,
        displayColors: props.series.length > 1,
        callbacks: {
          label: (ctx: TooltipItem<"bar">) => {
            const value = props.format(Number(ctx.parsed[props.horizontal ? "x" : "y"]))
            return props.series.length > 1 ? `${value} ${ctx.dataset.label}` : value
          },
        },
      },
    },
    // The mark is the hit target, and the pointer only has to share its
    // category — no aiming at a 2px sliver.
    interaction: { mode: "index", intersect: false },
  }
})
</script>
