import { describe, it, expect } from "vitest"
import { render, screen } from "@testing-library/vue"
import userEvent from "@testing-library/user-event"
import InsightsTable from "./InsightsTable.vue"
import type { Bucket, ResourceTotal } from "@/utils/insights"

const buckets: Bucket[] = [
  { start: new Date(2026, 7, 23), label: "Aug 23", onMs: 5_400_000, turnedOn: 2, turnedOff: 1 },
  { start: new Date(2026, 7, 24), label: "Aug 24", onMs: 0, turnedOn: 0, turnedOff: 1 },
]

const resources: ResourceTotal[] = [
  { id: "r1", name: "Chambre", type: "room", onMs: 5_400_000, changes: 4 },
]

const heatmap = Array.from({ length: 7 }, (_, r) =>
  Array.from({ length: 24 }, (_, h) => (r === 0 && h === 20 ? 1 : 0))
)

function renderTable() {
  return render(InsightsTable, {
    props: { buckets, bucketNoun: "day", resources, heatmap },
  })
}

describe("InsightsTable", () => {
  it("lists every plotted bucket value as text", () => {
    renderTable()
    expect(screen.getByText("Aug 23")).toBeInTheDocument()
    expect(screen.getByText("1h 30m")).toBeInTheDocument()
  })

  it("switches to per-resource rows", async () => {
    renderTable()
    await userEvent.click(screen.getByText("By resource"))
    expect(screen.getByText("Chambre")).toBeInTheDocument()
    expect(screen.getByText("room")).toBeInTheDocument()
    expect(screen.getByText("4")).toBeInTheDocument()
  })

  it("switches to one row per hour of the day, naming the busiest weekday", async () => {
    renderTable()
    await userEvent.click(screen.getByText("By hour"))
    expect(screen.getAllByText(/^\d\d:00$/)).toHaveLength(24)
    // A single Monday slot at 100% averages to 1/7 of the week.
    expect(screen.getByText("14%")).toBeInTheDocument()
    expect(screen.getByText("Mon")).toBeInTheDocument()
  })
})
