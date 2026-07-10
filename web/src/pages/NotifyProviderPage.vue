<template>
  <div class="flex min-h-screen items-center justify-center px-6 py-10">
    <div class="w-full max-w-wq-52ch">
      <div class="text-center">
        <div
          class="mx-auto grid h-wq-74 w-wq-74 place-items-center rounded-wq-13 bg-wq-accent/10 text-wq-accent-2"
        >
          <AppIcon name="bell" :size="40" />
        </div>
        <h1 class="mt-5 text-wq-19 font-bold">Get notified</h1>
        <p class="mt-3 max-w-wq-44ch text-sm text-wq-muted mx-auto">
          Choose where Watchue should send a message when a watched light/zone/room turns on or off.
          Optional — click a provider to configure it, or click an already-configured one to make it
          active.
        </p>
      </div>

      <p
        v-if="notifyStore.envLocked"
        class="mt-5 rounded-lg border border-wq-border bg-wq-panel-2 px-4 py-3 text-left text-xs text-wq-muted"
      >
        Configured via <span class="font-mono-ui">{{ envLockVarName }}</span> — remove that
        environment variable and restart Watchue to manage this from here instead.
      </p>

      <p v-if="activateError" class="mt-5 text-center text-sm text-wq-muted">
        {{ activateError }}
      </p>

      <div class="mt-5 grid grid-cols-1 gap-4 sm:grid-cols-2">
        <ProviderCard
          provider="telegram"
          :configured="notifyStore.status.telegram.configured"
          :active="notifyStore.activeProvider === 'telegram'"
          :env-locked="notifyStore.envLocked"
          @click="onCardClick('telegram')"
        />
        <ProviderCard
          provider="discord"
          :configured="notifyStore.status.discord.configured"
          :active="notifyStore.activeProvider === 'discord'"
          :env-locked="notifyStore.envLocked"
          @click="onCardClick('discord')"
        />
      </div>

      <div class="mt-6 text-center">
        <button
          type="button"
          class="cursor-pointer text-xs font-semibold text-wq-muted underline"
          @click="router.push('/')"
        >
          Back to dashboard
        </button>
      </div>
    </div>

    <Modal v-model="modalOpen">
      <h2 class="text-wq-15 font-bold">{{ editingLabel }} credentials</h2>

      <div class="mt-4 space-y-2.5 text-left">
        <template v-if="editingProvider === 'telegram'">
          <div>
            <label for="telegram-bot-token" class="mb-1 block text-xs font-semibold text-wq-muted"
              >Bot token</label
            >
            <input
              id="telegram-bot-token"
              v-model="telegramBotToken"
              type="text"
              :placeholder="isEditingConfigured ? '•••••••• (already configured)' : ''"
              class="w-full rounded-lg border border-wq-border bg-wq-panel-2 px-3 py-2 text-sm"
            />
          </div>
          <div>
            <label for="telegram-chat-id" class="mb-1 block text-xs font-semibold text-wq-muted"
              >Chat id</label
            >
            <input
              id="telegram-chat-id"
              v-model="telegramChatId"
              type="text"
              :placeholder="isEditingConfigured ? '•••••••• (already configured)' : ''"
              class="w-full rounded-lg border border-wq-border bg-wq-panel-2 px-3 py-2 text-sm"
            />
          </div>
        </template>
        <template v-else-if="editingProvider === 'discord'">
          <div>
            <label for="discord-webhook-url" class="mb-1 block text-xs font-semibold text-wq-muted"
              >Discord webhook URL</label
            >
            <input
              id="discord-webhook-url"
              v-model="discordWebhookUrl"
              type="text"
              :placeholder="isEditingConfigured ? '•••••••• (already configured)' : ''"
              class="w-full rounded-lg border border-wq-border bg-wq-panel-2 px-3 py-2 text-sm"
            />
          </div>
        </template>
      </div>
      <p v-if="isEditingConfigured" class="mt-2 text-xs text-wq-faint">
        Credentials are never sent back to this page — re-enter a value to test or change it.
      </p>

      <p
        v-if="notifyStore.testStatus === 'success'"
        class="mt-3 flex items-center justify-center gap-1.5 text-sm font-semibold text-wq-good"
      >
        <AppIcon name="check" :size="16" />
        Test notification sent!
      </p>
      <p v-else-if="notifyStore.testStatus === 'error'" class="mt-3 text-sm text-wq-muted">
        {{ notifyStore.testError }}
      </p>
      <p v-if="modalError" class="mt-3 text-sm text-wq-muted">{{ modalError }}</p>

      <div class="mt-5 flex justify-center gap-3">
        <button
          type="button"
          class="rounded-xl bg-wq-accent px-6 py-2.5 text-sm font-semibold text-white disabled:opacity-40"
          :disabled="!canTest || notifyStore.testStatus === 'testing'"
          @click="onTest"
        >
          {{ notifyStore.testStatus === "testing" ? "Testing…" : "Send test" }}
        </button>
        <button
          v-if="notifyStore.testStatus === 'success' && hasTypedValue"
          type="button"
          class="rounded-xl bg-wq-accent-2 px-6 py-2.5 text-sm font-semibold text-white"
          @click="onSave"
        >
          Save
        </button>
      </div>
      <button
        type="button"
        class="mt-4 block w-full cursor-pointer text-center text-xs font-semibold text-wq-muted underline"
        @click="modalOpen = false"
      >
        Cancel
      </button>
    </Modal>
  </div>
</template>

<script setup lang="ts">
import { computed, onActivated, ref, watch } from "vue"
import { useRouter } from "vue-router"
import AppIcon from "@/components/AppIcon.vue"
import Modal from "@/components/Modal.vue"
import ProviderCard from "@/components/ProviderCard.vue"
import { useNotifyStore, type NotifyConfig, type NotifyProvider } from "@/stores/useNotifyStore"

const router = useRouter()
const notifyStore = useNotifyStore()

const editingProvider = ref<NotifyProvider | null>(null)
const telegramBotToken = ref("")
const telegramChatId = ref("")
const discordWebhookUrl = ref("")
const activateError = ref("")
const modalError = ref("")

const modalOpen = computed({
  get: () => editingProvider.value !== null,
  set: (value: boolean) => {
    if (!value) editingProvider.value = null
  },
})

const editingLabel = computed(() => (editingProvider.value === "discord" ? "Discord" : "Telegram"))

const envLockVarName = computed(() =>
  notifyStore.activeProvider === "discord"
    ? "DISCORD_WEBHOOK_URL"
    : "TELEGRAM_BOT_TOKEN / TELEGRAM_CHAT_ID"
)

// The modal only ever opens for an unconfigured provider or an already-active
// one — a configured-but-inactive card activates directly, see onCardClick.
const isEditingConfigured = computed(() =>
  editingProvider.value ? notifyStore.status[editingProvider.value].configured : false
)

// A blank value falls back to the stored credentials (see notify.go's notifyConfigBlank).
const hasTypedValue = computed(() =>
  editingProvider.value === "telegram"
    ? telegramBotToken.value.trim() !== "" && telegramChatId.value.trim() !== ""
    : discordWebhookUrl.value.trim() !== ""
)

const canTest = computed(() => hasTypedValue.value || isEditingConfigured.value)

const notifyConfig = computed<NotifyConfig>(() =>
  editingProvider.value === "telegram"
    ? {
        provider: "telegram",
        telegram_bot_token: telegramBotToken.value,
        telegram_chat_id: telegramChatId.value,
      }
    : { provider: "discord", discord_webhook_url: discordWebhookUrl.value }
)

// A stale test result/error must not survive an edited credential.
watch([telegramBotToken, telegramChatId, discordWebhookUrl], () => {
  notifyStore.resetTest()
  modalError.value = ""
})

// onActivated, not onMounted: App.vue wraps every route in <KeepAlive>, so
// onMounted would only fire once and miss later status changes on revisit.
onActivated(async () => {
  editingProvider.value = null
  activateError.value = ""
  await notifyStore.fetchStatus()
})

function openModal(provider: NotifyProvider): void {
  editingProvider.value = provider
  telegramBotToken.value = ""
  telegramChatId.value = ""
  discordWebhookUrl.value = ""
  notifyStore.resetTest()
  modalError.value = ""
}

async function onCardClick(provider: NotifyProvider): Promise<void> {
  if (notifyStore.envLocked) return

  if (!notifyStore.status[provider].configured) {
    openModal(provider)
    return
  }

  if (notifyStore.activeProvider !== provider) {
    activateError.value = ""
    try {
      await notifyStore.activate(provider)
      await notifyStore.fetchStatus()
    } catch (err) {
      activateError.value = err instanceof Error ? err.message : "Failed to activate provider"
    }
    return
  }

  // Already configured and already active — nothing to activate, open the
  // modal instead so it can be retested or replaced.
  openModal(provider)
}

async function onTest(): Promise<void> {
  await notifyStore.test(notifyConfig.value)
}

async function onSave(): Promise<void> {
  modalError.value = ""
  try {
    await notifyStore.save(notifyConfig.value)
    await notifyStore.fetchStatus()
    editingProvider.value = null
  } catch (err) {
    modalError.value = err instanceof Error ? err.message : "Failed to save notification settings"
  }
}
</script>
