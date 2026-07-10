<template>
  <div
    class="flex items-center gap-3.5 border-b border-wq-border-2 px-wq-18 py-3.5 last:border-b-0"
  >
    <div class="font-mono-ui w-wq-74 flex-none text-right text-xs font-medium text-wq-faint">
      {{ time }}
    </div>
    <div
      class="grid h-wq-34 w-wq-34 flex-none place-items-center rounded-lg border border-wq-border"
      :class="
        on
          ? '!border-wq-accent/30 !bg-wq-accent/10 !text-wq-accent-2'
          : 'bg-wq-panel-2 text-wq-faint'
      "
    >
      <AppIcon name="power" :size="18" />
    </div>
    <div class="min-w-0 flex-1">
      <div class="text-sm font-semibold">{{ name }}</div>
      <div class="mt-0.5 flex flex-wrap items-center gap-1.5 text-xs text-wq-muted">
        <span class="font-mono-ui text-wq-10 font-semibold tracking-wide text-wq-faint uppercase">{{
          typeLabel
        }}</span>
        <span>·</span>
        <span>Turned {{ on ? "on" : "off" }}</span>
      </div>
    </div>
    <span
      class="font-mono-ui inline-flex flex-none items-center gap-1.5 rounded-md px-1.5 py-1 text-wq-10-5 font-semibold tracking-wide"
      :class="tagClass"
    >
      <AppIcon :name="tagIcon" :size="13" />
      {{ tagLabel }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue"
import AppIcon from "@/components/AppIcon.vue"
import type { EventOutcome } from "@/stores/useEventsStore"

const props = defineProps<{
  name: string
  typeLabel: string
  on: boolean
  outcome: EventOutcome
  time: string
}>()

const tagClass = computed(() =>
  props.outcome === "sent"
    ? "bg-wq-tg/15 text-wq-tg"
    : "border border-wq-border bg-wq-panel-2 text-wq-faint"
)
const tagIcon = computed(() => (props.outcome === "sent" ? "telegram" : "bell-off"))
const tagLabel = computed(() =>
  props.outcome === "sent" ? "sent" : props.outcome === "muted" ? "muted" : "channel off"
)
</script>
