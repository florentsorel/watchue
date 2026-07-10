import { describe, it, expect } from "vitest"
import { render, screen } from "@testing-library/vue"
import userEvent from "@testing-library/user-event"
import SegmentedControl from "./SegmentedControl.vue"

const options = [
  { label: "All", value: "all" },
  { label: "On", value: "on" },
  { label: "Off", value: "off" },
]

describe("SegmentedControl", () => {
  it("renders every option", () => {
    render(SegmentedControl, { props: { modelValue: "all", options } })
    expect(screen.getByText("All")).toBeInTheDocument()
    expect(screen.getByText("On")).toBeInTheDocument()
    expect(screen.getByText("Off")).toBeInTheDocument()
  })

  it("emits update:modelValue with the clicked option's value", async () => {
    const { emitted } = render(SegmentedControl, { props: { modelValue: "all", options } })
    await userEvent.click(screen.getByText("On"))
    expect(emitted()["update:modelValue"]).toEqual([["on"]])
  })
})
