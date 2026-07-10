import "./assets/main.css"
import { createApp } from "vue"
import { createPinia } from "pinia"
import { createRouter, createWebHistory } from "vue-router"
import App from "./App.vue"
import DashboardPage from "./pages/DashboardPage.vue"

const app = createApp(App)

app.use(createPinia())

const router = createRouter({
  history: createWebHistory(),
  routes: [{ path: "/", component: DashboardPage }],
})
app.use(router)

app.mount("#app")
