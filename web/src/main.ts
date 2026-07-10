import "./assets/main.css"
import { createApp } from "vue"
import { createPinia } from "pinia"
import { createRouter, createWebHistory } from "vue-router"
import App from "./App.vue"
import DashboardPage from "./pages/DashboardPage.vue"
import SetupPage from "./pages/SetupPage.vue"
import NotifyProviderPage from "./pages/NotifyProviderPage.vue"
import { useSetupStore } from "./stores/useSetupStore"
import { useUiStore } from "./stores/useUiStore"

const app = createApp(App)

const pinia = createPinia()
app.use(pinia)

useUiStore(pinia)

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: "/", component: DashboardPage },
    { path: "/setup", component: SetupPage },
    { path: "/provider", component: NotifyProviderPage },
  ],
})

let statusChecked = false
router.beforeEach(async (to) => {
  const setupStore = useSetupStore(pinia)
  if (!statusChecked) {
    await setupStore.checkStatus().catch(() => {})
    statusChecked = true
  }
  if (!setupStore.configured && to.path !== "/setup") return "/setup"
  if (setupStore.configured && to.path === "/setup") return "/"
  return true
})

app.use(router)

app.mount("#app")
