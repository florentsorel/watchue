import { describe, it, expect, vi, beforeEach } from "vitest"
import { createPinia, setActivePinia } from "pinia"
import { defineComponent } from "vue"
import { render } from "@testing-library/vue"
import { useEventStream } from "./useEventStream"
import { useBridgeStore } from "@/stores/useBridgeStore"
import { useEventsStore } from "@/stores/useEventsStore"
import { useSettingsStore } from "@/stores/useSettingsStore"

class FakeEventSource {
  onmessage: ((e: { data: string }) => void) | null = null
  close = vi.fn()
}

let lastSource: FakeEventSource

beforeEach(() => {
  setActivePinia(createPinia())
  vi.stubGlobal(
    "EventSource",
    vi.fn().mockImplementation(function () {
      lastSource = new FakeEventSource()
      return lastSource
    })
  )
})

const Host = defineComponent({
  setup() {
    useEventStream()
    return () => null
  },
})

describe("useEventStream", () => {
  it("applies a resource message to the bridge store", () => {
    const bridgeStore = useBridgeStore()
    bridgeStore.zones = [{ id: "zone-1", name: "Salon", on: false, archetype: "", lights: [] }]
    render(Host)

    lastSource.onmessage?.({ data: JSON.stringify({ kind: "resource", id: "zone-1", on: true }) })

    expect(bridgeStore.zones[0].on).toBe(true)
  })

  it("prepends an event message to the events store", () => {
    const eventsStore = useEventsStore()
    render(Host)

    lastSource.onmessage?.({
      data: JSON.stringify({
        kind: "event",
        id: 1,
        resource_id: "zone-1",
        resource_type: "zone",
        name: "Salon",
        on: true,
        outcome: "sent",
        created_at: "2026-01-01 00:00:00",
      }),
    })

    expect(eventsStore.items).toHaveLength(1)
    expect(eventsStore.items[0].name).toBe("Salon")
  })

  it("ignores malformed messages instead of throwing", () => {
    render(Host)
    expect(() => lastSource.onmessage?.({ data: "not json" })).not.toThrow()
  })

  it("updates bridge online status from a bridge_status message", () => {
    const settingsStore = useSettingsStore()
    settingsStore.bridgeOnline = true
    render(Host)

    lastSource.onmessage?.({ data: JSON.stringify({ kind: "bridge_status", online: false }) })

    expect(settingsStore.bridgeOnline).toBe(false)
  })
})
