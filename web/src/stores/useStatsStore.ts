import { defineStore } from "pinia"
import { ref } from "vue"
import type { Session } from "@/utils/insights"

// "All" is a ceiling rather than a real range — the backend caps `days` at
// 3650 and there is no earlier data than the first recorded event anyway.
export const STATS_RANGES = [
  { label: "7d", value: 7 },
  { label: "30d", value: 30 },
  { label: "90d", value: 90 },
  { label: "1y", value: 365 },
] as const

export const useStatsStore = defineStore("stats", () => {
  const sessions = ref<Session[]>([])
  const days = ref<number>(30)
  const loading = ref(false)
  const loaded = ref(false)

  async function load(range = days.value): Promise<void> {
    days.value = range
    loading.value = true
    try {
      const res = await fetch(`/api/stats?days=${range}`)
      if (res.ok) {
        const body: { sessions: Session[] } = await res.json()
        sessions.value = body.sessions
        loaded.value = true
      }
    } finally {
      loading.value = false
    }
  }

  return { sessions, days, loading, loaded, load }
})
