import { defineStore } from "pinia"
import { ref, watch } from "vue"

export type ResourceLayout = "glow" | "compact" | "wall"
export type HistoryFilter = "all" | "on" | "off" | "muted"

function readLocalStorage(key: string): string | null {
  try {
    return localStorage.getItem(key)
  } catch {
    return null
  }
}

function writeLocalStorage(key: string, value: string): void {
  try {
    localStorage.setItem(key, value)
  } catch {
    // ignore (e.g. private-browsing storage restrictions)
  }
}

function prefersDark(): boolean {
  try {
    return window.matchMedia("(prefers-color-scheme: dark)").matches
  } catch {
    return false
  }
}

function initialTheme(): "light" | "dark" {
  const stored = readLocalStorage("watchue-theme")
  if (stored === "dark" || stored === "light") return stored
  return prefersDark() ? "dark" : "light"
}

// Pure UI state — theme and display preferences — not backed by the API.
export const useUiStore = defineStore("ui", () => {
  const theme = ref<"light" | "dark">(initialTheme())
  const layout = ref<ResourceLayout>(
    (readLocalStorage("watchue-layout") as ResourceLayout) || "glow"
  )
  const historyFilter = ref<HistoryFilter>("all")

  function applyTheme(): void {
    document.documentElement.classList.toggle("dark", theme.value === "dark")
  }
  applyTheme()

  function toggleTheme(): void {
    theme.value = theme.value === "dark" ? "light" : "dark"
  }

  watch(theme, (value) => {
    writeLocalStorage("watchue-theme", value)
    applyTheme()
  })

  function setLayout(value: ResourceLayout): void {
    layout.value = value
  }

  watch(layout, (value) => writeLocalStorage("watchue-layout", value))

  function setHistoryFilter(value: HistoryFilter): void {
    historyFilter.value = value
  }

  return { theme, toggleTheme, layout, setLayout, historyFilter, setHistoryFilter }
})
