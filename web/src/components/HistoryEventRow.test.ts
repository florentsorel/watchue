import { describe, it, expect } from "vitest"
import { render, screen } from "@testing-library/vue"
import HistoryEventRow from "./HistoryEventRow.vue"

function renderRow(outcome: "sent" | "muted" | "channel_off") {
  return render(HistoryEventRow, {
    props: {
      name: "Salon",
      typeLabel: "Zone",
      on: true,
      outcome,
      time: "5m ago",
    },
  })
}

describe("HistoryEventRow", () => {
  it("labels a sent event", () => {
    renderRow("sent")
    expect(screen.getByText("sent")).toBeInTheDocument()
  })

  it("labels a muted event", () => {
    renderRow("muted")
    expect(screen.getByText("muted")).toBeInTheDocument()
  })

  it("labels a channel_off event", () => {
    renderRow("channel_off")
    expect(screen.getByText("channel off")).toBeInTheDocument()
  })

  it("shows the resource name and turned-on/off state", () => {
    renderRow("sent")
    expect(screen.getByText("Salon")).toBeInTheDocument()
    expect(screen.getByText("Turned on")).toBeInTheDocument()
  })
})
