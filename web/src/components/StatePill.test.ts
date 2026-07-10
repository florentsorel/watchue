import { describe, it, expect } from "vitest"
import { render, screen } from "@testing-library/vue"
import StatePill from "./StatePill.vue"

describe("StatePill", () => {
  it("shows On when on is true", () => {
    render(StatePill, { props: { on: true } })
    expect(screen.getByText("On")).toBeInTheDocument()
  })

  it("shows Off when on is false", () => {
    render(StatePill, { props: { on: false } })
    expect(screen.getByText("Off")).toBeInTheDocument()
  })
})
