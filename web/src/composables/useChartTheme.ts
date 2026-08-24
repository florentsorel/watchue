import { computed } from "vue"
import { useUiStore } from "@/stores/useUiStore"

export interface ChartTheme {
  on: string
  off: string
  series: string[]
  heat: string[]
  surface: string
  grid: string
  tick: string
  tooltipBg: string
  tooltipText: string
}

function cssVar(name: string): string {
  return getComputedStyle(document.documentElement).getPropertyValue(name).trim()
}

/**
 * The chart palette, resolved from the CSS custom properties in main.css so the
 * charts and the rest of the UI cannot drift apart.
 *
 * Canvas needs literal color strings — unlike the DOM it can't follow a `var()`
 * through a theme switch — so this re-reads them whenever the theme changes.
 * Reading `ui.theme` is what registers that dependency; the `.dark` class is
 * already on `<html>` by the time this recomputes, since useUiStore toggles it
 * from a pre-flush watcher, which runs before any component re-renders.
 */
export function useChartTheme() {
  const ui = useUiStore()
  return computed<ChartTheme>(() => {
    const dark = ui.theme === "dark"
    return {
      on: cssVar("--wq-chart-on"),
      off: cssVar("--wq-chart-off"),
      series: [1, 2, 3].map((i) => cssVar(`--wq-series-${i}`)),
      heat: [1, 2, 3, 4, 5].map((i) => cssVar(`--wq-heat-${i}`)),
      surface: cssVar("--wq-panel"),
      grid: cssVar("--wq-border-2"),
      tick: cssVar("--wq-faint"),
      tooltipBg: dark ? cssVar("--wq-panel-2") : cssVar("--wq-text"),
      tooltipText: dark ? cssVar("--wq-text") : cssVar("--wq-panel"),
    }
  })
}
