import { useWatchedStore } from "@/stores/useWatchedStore"

export function useToggleWatch() {
  const watchedStore = useWatchedStore()
  return async function toggleWatch(id: string) {
    if (watchedStore.isWatched(id)) await watchedStore.unwatch(id)
    else await watchedStore.watch(id)
  }
}
