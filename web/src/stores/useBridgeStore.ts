import { defineStore } from "pinia"
import { ref } from "vue"

export interface BridgeLight {
  id: string
  name: string
  on: boolean
  archetype: string
}

export interface BridgeGroup {
  id: string
  name: string
  on: boolean
  archetype: string
  lights: BridgeLight[]
}

// Bridge's current zones/rooms/lights — distinct from useWatchedStore (the saved selection).
export const useBridgeStore = defineStore("bridge", () => {
  const zones = ref<BridgeGroup[]>([])
  const rooms = ref<BridgeGroup[]>([])
  const loading = ref(false)

  async function load(): Promise<void> {
    loading.value = true
    try {
      const [zonesRes, roomsRes] = await Promise.all([fetch("/api/zones"), fetch("/api/rooms")])
      if (zonesRes.ok) zones.value = await zonesRes.json()
      if (roomsRes.ok) rooms.value = await roomsRes.json()
    } finally {
      loading.value = false
    }
  }

  // Applies a live on/off update (see useEventStream) to whichever zone,
  // room, or nested light matches id.
  function applyOnUpdate(id: string, on: boolean): void {
    for (const z of zones.value) {
      if (z.id === id) {
        z.on = on
        continue
      }
      const light = z.lights.find((l) => l.id === id)
      if (light) light.on = on
    }
    for (const r of rooms.value) {
      if (r.id === id) {
        r.on = on
        continue
      }
      const light = r.lights.find((l) => l.id === id)
      if (light) light.on = on
    }
  }

  return { zones, rooms, loading, load, applyOnUpdate }
})
