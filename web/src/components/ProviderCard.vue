<template>
  <button
    type="button"
    class="wq-card flex w-full flex-col items-center gap-3 p-6 text-center transition-colors"
    :class="[
      active ? 'border-wq-good/40' : '',
      envLocked ? 'cursor-default' : 'cursor-pointer hover:border-wq-accent/30',
    ]"
    @click="$emit('click')"
  >
    <div class="grid h-14 w-14 place-items-center rounded-xl border border-wq-border bg-wq-panel-2">
      <AppIcon :name="`${provider}-logo`" :size="32" />
    </div>
    <div class="text-wq-19 font-semibold tracking-tight">{{ label }}</div>
    <div v-if="envLocked" class="text-xs text-wq-faint">Configured via env</div>
    <div v-else-if="active" class="text-xs font-semibold text-wq-good">Active</div>
    <div
      v-else-if="configured"
      class="inline-flex items-center gap-1.5 text-xs font-semibold text-wq-info"
    >
      <AppIcon name="check" :size="14" />
      Configured
    </div>
    <div v-else class="text-xs font-semibold text-wq-faint">Not set</div>
  </button>
</template>

<script setup lang="ts">
import { computed } from "vue"
import AppIcon from "@/components/AppIcon.vue"
import type { NotifyProvider } from "@/stores/useNotifyStore"

const props = defineProps<{
  provider: NotifyProvider
  configured: boolean
  active: boolean
  envLocked: boolean
}>()

defineEmits<{ click: [] }>()

const label = computed(() => (props.provider === "discord" ? "Discord" : "Telegram"))
</script>
