import { defineStore } from "pinia"
import { ref } from "vue"

export type EventOutcome = "sent" | "muted" | "channel_off"

export interface WatchEvent {
  id: number
  resource_id: string
  resource_type: "zone" | "room" | "light"
  name: string
  on: boolean
  outcome: EventOutcome // fixed at the time of the change
  created_at: string
}

export const useEventsStore = defineStore("events", () => {
  const items = ref<WatchEvent[]>([])
  const loading = ref(false)

  async function load(limit = 50): Promise<void> {
    loading.value = true
    try {
      const res = await fetch(`/api/events?limit=${limit}`)
      if (res.ok) items.value = await res.json()
    } finally {
      loading.value = false
    }
  }

  function prepend(event: WatchEvent): void {
    items.value = [event, ...items.value]
  }

  return { items, loading, load, prepend }
})
