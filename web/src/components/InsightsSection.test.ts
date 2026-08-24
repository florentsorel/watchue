import { describe, it, expect, vi, beforeEach, afterEach } from "vitest"
import { createPinia, setActivePinia } from "pinia"
import { render, screen } from "@testing-library/vue"
import userEvent from "@testing-library/user-event"
import InsightsSection from "./InsightsSection.vue"
import { useStatsStore } from "@/stores/useStatsStore"
import type { Session } from "@/utils/insights"

// Chart.js draws to a canvas, which happy-dom has no 2d context for; the marks
// are its own concern anyway — what matters here is the data the section feeds
// it and the numbers it prints alongside.
const stubs = { InsightsBarChart: true, NightlySwitchChart: true }

const NOW = new Date(2026, 7, 24, 12, 0, 0)

function session(start: string, end: string | null, name = "Chambre"): Session {
  return { resource_id: name, resource_type: "room", name, start, end }
}

function utc(local: Date): string {
  return local.toISOString().slice(0, 19).replace("T", " ")
}

// Scoped to the stat tiles: "Time on" also names one of the chart cards.
function tile(label: string): HTMLElement {
  const found = screen
    .getAllByText(label)
    .map((el) => el.closest(".wq-card"))
    .find((card): card is HTMLElement => card?.classList.contains("p-4") ?? false)
  if (!found) throw new Error(`no stat tile labelled "${label}"`)
  return found
}

function renderSection(sessions: Session[]) {
  setActivePinia(createPinia())
  const store = useStatsStore()
  store.sessions = sessions
  store.loaded = true
  const utils = render(InsightsSection, { global: { stubs } })
  return { ...utils, store }
}

// The section derives every bucket from Date.now(), so the clock is pinned;
// user-event has to be told to drive that same fake clock or its own waits hang.
function user() {
  return userEvent.setup({ advanceTimers: vi.advanceTimersByTime })
}

beforeEach(() => {
  vi.useFakeTimers()
  vi.setSystemTime(NOW)
})

afterEach(() => vi.useRealTimers())

describe("InsightsSection", () => {
  it("invites the user to watch something when nothing is recorded", () => {
    renderSection([])
    expect(screen.getByText("Nothing recorded in this range yet.")).toBeInTheDocument()
  })

  it("totals switch-ons and time on across the range", () => {
    renderSection([
      session(utc(new Date(2026, 7, 23, 18, 0)), utc(new Date(2026, 7, 23, 21, 0))),
      session(utc(new Date(2026, 7, 24, 9, 0)), utc(new Date(2026, 7, 24, 10, 0))),
    ])
    expect(tile("Switched on")).toHaveTextContent("2")
    expect(tile("Switched on")).toHaveTextContent("2 switched off")
    expect(tile("Time on")).toHaveTextContent("4h") // 3h + 1h
    expect(tile("Avg time on")).toHaveTextContent("2h")
  })

  it("names the busiest hour", () => {
    renderSection([
      session(utc(new Date(2026, 7, 22, 20, 0)), utc(new Date(2026, 7, 22, 21, 0))),
      session(utc(new Date(2026, 7, 23, 20, 0)), utc(new Date(2026, 7, 23, 21, 0))),
    ])
    expect(tile("Busiest hour")).toHaveTextContent("20h")
  })

  it("reloads the whole section when the range changes", async () => {
    const { store } = renderSection([session(utc(new Date(2026, 7, 23, 18, 0)), null)])
    const load = vi.spyOn(store, "load").mockResolvedValue()
    await user().click(screen.getByText("7d"))
    expect(load).toHaveBeenCalledWith(7)
  })

  it("reveals the table twin holding the same numbers as the charts", async () => {
    renderSection([session(utc(new Date(2026, 7, 23, 18, 0)), utc(new Date(2026, 7, 23, 21, 0)))])
    await user().click(screen.getByText("Show data table"))
    expect(screen.getByRole("table")).toHaveTextContent("3h")
  })
})

describe("InsightsSection — when the lights go off", () => {
  const sessions = [
    session(utc(new Date(2026, 7, 23, 20, 0)), utc(new Date(2026, 7, 24, 1, 17)), "Chambre"),
    session(utc(new Date(2026, 7, 23, 19, 0)), utc(new Date(2026, 7, 23, 19, 30)), "Bureau"),
  ]

  it("offers every resource and starts on the busiest one", () => {
    renderSection(sessions)
    expect(screen.getByRole("button", { name: "Chambre" })).toHaveAttribute("aria-pressed", "true")
    expect(screen.getByRole("button", { name: "Bureau" })).toHaveAttribute("aria-pressed", "false")
  })

  it("refuses to leave the chart with no resource selected", async () => {
    renderSection(sessions)
    await user().click(screen.getByRole("button", { name: "Chambre" }))
    expect(screen.getByRole("button", { name: "Chambre" })).toHaveAttribute("aria-pressed", "true")
  })

  it("explains the noon-to-noon columns by default", () => {
    renderSection(sessions)
    expect(screen.getByText(/from noon to noon/)).toBeInTheDocument()
  })

  it("switches to calendar days on request", async () => {
    renderSection(sessions)
    await user().click(screen.getByText("Calendar day"))
    expect(screen.getByText(/One column per calendar day/)).toBeInTheDocument()
  })
})
