import { onMounted, onUnmounted } from "vue"
import { useBridgeStore } from "@/stores/useBridgeStore"
import { useEventsStore, type WatchEvent } from "@/stores/useEventsStore"
import { useSettingsStore } from "@/stores/useSettingsStore"

interface ResourceMessage {
  kind: "resource"
  id: string
  on: boolean
}

type EventMessage = { kind: "event" } & WatchEvent

interface BridgeStatusMessage {
  kind: "bridge_status"
  online: boolean
}

type StreamMessage = ResourceMessage | EventMessage | BridgeStatusMessage

// Live updates over /api/stream (SSE). Plain EventSource, not a fetch+
// ReadableStream reader like postr's useSSEStream: that composable exists to
// stream a POST request's response (EventSource only does GET), which
// doesn't apply here — a plain GET stream is exactly what EventSource is
// for, and it gets automatic reconnection for free.
export function useEventStream() {
  const bridgeStore = useBridgeStore()
  const eventsStore = useEventsStore()
  const settingsStore = useSettingsStore()

  let source: EventSource | null = null

  function handleMessage(raw: string): void {
    let msg: StreamMessage
    try {
      msg = JSON.parse(raw)
    } catch {
      return
    }
    if (msg.kind === "resource") {
      bridgeStore.applyOnUpdate(msg.id, msg.on)
    } else if (msg.kind === "event") {
      eventsStore.prepend(msg)
    } else if (msg.kind === "bridge_status") {
      settingsStore.bridgeOnline = msg.online
    }
  }

  onMounted(() => {
    source = new EventSource("/api/stream")
    source.onmessage = (e) => handleMessage(e.data)
  })

  onUnmounted(() => {
    source?.close()
  })
}
