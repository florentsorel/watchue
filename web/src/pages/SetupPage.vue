<template>
  <div class="flex min-h-screen items-center justify-center px-6">
    <div class="wq-card w-full max-w-wq-52ch p-8 text-center">
      <div
        class="mx-auto grid h-wq-74 w-wq-74 place-items-center rounded-wq-13 bg-wq-accent/10 text-wq-accent-2"
      >
        <AppIcon name="bridge" :size="40" />
      </div>

      <h1 class="mt-5 text-wq-19 font-bold">Connect your Hue Bridge</h1>
      <p v-if="setupStore.hueBridgeHost" class="font-mono-ui mt-1 text-xs text-wq-muted">
        {{ setupStore.hueBridgeHost }}
      </p>

      <template v-if="setupStore.status === 'idle'">
        <p class="mt-3 max-w-wq-44ch text-sm text-wq-muted mx-auto">
          Press the link button on your Hue Bridge, then click below. Watchue will detect it
          automatically.
        </p>
        <button
          type="button"
          class="mt-5 rounded-xl bg-wq-accent-2 px-6 py-2.5 text-sm font-semibold text-white"
          @click="onStartPairing"
        >
          Start pairing
        </button>
      </template>

      <template v-else-if="setupStore.status === 'waiting_for_button'">
        <p class="mt-3 text-sm text-wq-muted">
          Waiting for you to press the button on your bridge&hellip;
        </p>
      </template>

      <template v-else-if="setupStore.status === 'paired'">
        <p class="mt-3 flex items-center justify-center gap-1.5 text-sm font-semibold text-wq-good">
          <AppIcon name="check" :size="16" />
          Paired successfully!
        </p>
        <button
          type="button"
          class="mt-5 rounded-xl bg-wq-accent-2 px-6 py-2.5 text-sm font-semibold text-white"
          @click="onNext"
        >
          Next
        </button>
      </template>

      <template v-else-if="setupStore.status === 'restarting'">
        <p class="mt-3 text-sm text-wq-muted">Restarting Watchue&hellip;</p>
      </template>

      <template v-else-if="setupStore.status === 'error'">
        <p class="mt-3 text-sm text-wq-muted">{{ setupStore.errorMessage }}</p>
        <button
          type="button"
          class="mt-5 rounded-xl bg-wq-accent-2 px-6 py-2.5 text-sm font-semibold text-white"
          @click="onStartPairing"
        >
          Try again
        </button>
      </template>

      <p class="mt-6 max-w-wq-44ch text-wq-11 text-wq-faint mx-auto">
        Watchue restarts automatically once paired (under the Docker Compose
        <span class="font-mono-ui">restart: unless-stopped</span> policy from the Quick Start). If
        you're running it another way, restart it manually after this step.
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from "vue-router"
import AppIcon from "@/components/AppIcon.vue"
import { useSetupStore } from "@/stores/useSetupStore"

// hueBridgeHost/configured are already populated by the router guard's
// checkStatus() call before this page ever mounts (see main.ts) — the guard
// is also what keeps an already-configured session from reaching /setup.
const router = useRouter()
const setupStore = useSetupStore()

async function onStartPairing(): Promise<void> {
  await setupStore.startPairing()
}

async function onNext(): Promise<void> {
  await setupStore.waitForRestart()
  router.replace("/")
}
</script>
