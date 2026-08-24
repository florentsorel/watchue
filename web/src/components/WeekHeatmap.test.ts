import { describe, it, expect, beforeEach } from "vitest"
import { createPinia, setActivePinia } from "pinia"
import { render, screen } from "@testing-library/vue"
import WeekHeatmap from "./WeekHeatmap.vue"

function grid(fill: (row: number, hour: number) => number): number[][] {
  return Array.from({ length: 7 }, (_, r) => Array.from({ length: 24 }, (_, h) => fill(r, h)))
}

beforeEach(() => setActivePinia(createPinia()))

describe("WeekHeatmap", () => {
  it("renders a cell per weekday hour", () => {
    const { container } = render(WeekHeatmap, { props: { grid: grid(() => 0) } })
    expect(container.querySelectorAll(".aspect-square")).toHaveLength(7 * 24)
  })

  it("labels each cell with its weekday, hour and share", () => {
    render(WeekHeatmap, { props: { grid: grid((r, h) => (r === 2 && h === 19 ? 0.42 : 0)) } })
    expect(screen.getByTitle("Wed 19:00 — on 42% of the time")).toBeInTheDocument()
  })

  it("puts an hour label under every sixth column so they stay aligned", () => {
    render(WeekHeatmap, { props: { grid: grid(() => 0) } })
    for (const hour of ["0", "6", "12", "18"]) {
      expect(screen.getByText(hour)).toBeInTheDocument()
    }
  })
})
