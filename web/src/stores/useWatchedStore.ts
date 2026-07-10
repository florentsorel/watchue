import { defineStore } from "pinia"
import { ref } from "vue"

export interface WatchedResource {
  id: string
  type: "zone" | "room" | "light"
  name: string
  notify: boolean
}

export const useWatchedStore = defineStore("watched", () => {
  const items = ref<WatchedResource[]>([])
  const loading = ref(false)

  async function load(): Promise<void> {
    loading.value = true
    try {
      const res = await fetch("/api/watched")
      if (res.ok) items.value = await res.json()
    } finally {
      loading.value = false
    }
  }

  function isWatched(id: string): boolean {
    return items.value.some((w) => w.id === id)
  }

  // type/name are resolved server-side from the bridge — no body needed.
  async function watch(id: string): Promise<void> {
    const res = await fetch(`/api/watched/${id}`, { method: "PUT" })
    if (!res.ok) {
      const data = await res.json().catch(() => ({}))
      throw new Error(data.error ?? "Failed to watch resource")
    }
    await load()
  }

  async function unwatch(id: string): Promise<void> {
    const res = await fetch(`/api/watched/${id}`, { method: "DELETE" })
    if (!res.ok) {
      const data = await res.json().catch(() => ({}))
      throw new Error(data.error ?? "Failed to unwatch resource")
    }
    items.value = items.value.filter((w) => w.id !== id)
  }

  async function setNotify(id: string, notify: boolean): Promise<void> {
    const res = await fetch(`/api/watched/${id}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ notify }),
    })
    if (!res.ok) {
      const data = await res.json().catch(() => ({}))
      throw new Error(data.error ?? "Failed to update notification setting")
    }
    const item = items.value.find((w) => w.id === id)
    if (item) item.notify = notify
  }

  return { items, loading, load, isWatched, watch, unwatch, setNotify }
})
