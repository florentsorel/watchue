import "@testing-library/jest-dom"
import { vi, afterEach } from "vitest"
import { config } from "@vue/test-utils"
import { createRouter, createMemoryHistory } from "vue-router"

vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: false, json: async () => ({}) }))

vi.stubGlobal("localStorage", {
  getItem: vi.fn(() => null),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn(),
})

vi.stubGlobal("matchMedia", vi.fn().mockReturnValue({ matches: false }))

vi.stubGlobal(
  "EventSource",
  class {
    onmessage: ((e: { data: string }) => void) | null = null
    close = vi.fn()
  }
)

const router = createRouter({
  history: createMemoryHistory(),
  routes: [{ path: "/", component: { template: "<div />" } }],
})

config.global.plugins = [router]

afterEach(() => {
  document.body.innerHTML = ""
})
