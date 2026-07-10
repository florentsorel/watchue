<template>
  <div class="relative transition-colors" :class="cardClass">
    <div
      v-if="on && layout !== 'compact'"
      class="pointer-events-none absolute inset-0 rounded-inherit"
      style="
        background: radial-gradient(
          120% 90% at 80% -10%,
          color-mix(in srgb, var(--wq-accent) 13%, transparent),
          transparent 60%
        );
      "
    />

    <div class="relative flex w-full items-center gap-3">
      <ResourceIconBox :icon="icon" :on="on" :box-size="boxSize" :icon-size="iconSize" />

      <div class="min-w-0 flex-1">
        <div
          class="font-semibold tracking-tight"
          :class="layout === 'wall' ? 'text-wq-19' : 'text-wq-15'"
        >
          {{ name }}
        </div>
        <div
          v-if="layout !== 'compact'"
          class="font-mono-ui mt-1 text-wq-10-5 font-semibold tracking-wide text-wq-faint uppercase"
        >
          {{ type }}
        </div>
        <div class="mt-0.5 truncate text-wq-12-5 text-wq-muted">{{ meta }}</div>
      </div>

      <StatePill :on="on" />

      <div v-if="layout === 'compact'" class="ml-auto flex flex-none items-center gap-3">
        <button
          type="button"
          class="grid h-wq-30 w-wq-30 flex-none place-items-center rounded-lg border border-wq-border text-wq-muted transition-colors hover:border-wq-accent/40 hover:text-wq-text"
          :title="notify ? 'Mute notifications' : 'Unmute notifications'"
          @click="$emit('toggle-mute')"
        >
          <AppIcon :name="notify ? 'bell' : 'bell-off'" :size="16" />
        </button>
        <button
          type="button"
          class="grid h-wq-30 w-wq-30 flex-none place-items-center rounded-lg border border-wq-border text-wq-faint transition-colors hover:border-red-400/40 hover:text-red-500"
          title="Stop watching"
          @click="$emit('unwatch')"
        >
          <AppIcon name="trash" :size="16" />
        </button>
      </div>
    </div>

    <div
      v-if="layout !== 'compact'"
      class="relative mt-4 flex items-center justify-between gap-2.5 border-t border-wq-border-2 pt-3.5"
    >
      <button
        type="button"
        class="inline-flex h-wq-30 items-center gap-1.5 rounded-lg border border-wq-border px-2.5 text-xs font-semibold text-wq-muted transition-colors hover:border-wq-accent/40 hover:text-wq-text"
        :class="!notify ? 'text-wq-faint' : ''"
        @click="$emit('toggle-mute')"
      >
        <AppIcon :name="notify ? 'bell' : 'bell-off'" :size="16" />
        {{ notify ? "Notifying" : "Muted" }}
      </button>
      <button
        type="button"
        class="grid h-wq-30 w-wq-30 flex-none place-items-center rounded-lg border border-wq-border text-wq-faint transition-colors hover:border-red-400/40 hover:text-red-500"
        title="Stop watching"
        @click="$emit('unwatch')"
      >
        <AppIcon name="trash" :size="16" />
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue"
import AppIcon from "@/components/AppIcon.vue"
import StatePill from "@/components/StatePill.vue"
import ResourceIconBox from "@/components/ResourceIconBox.vue"
import type { ResourceLayout } from "@/stores/useUiStore"

const props = defineProps<{
  name: string
  type: string
  icon: string
  on: boolean
  notify: boolean
  meta: string
  layout: ResourceLayout
}>()

defineEmits<{ "toggle-mute": []; unwatch: [] }>()

const cardClass = computed(() =>
  props.layout === "compact"
    ? "flex items-center border-b border-wq-border-2 px-wq-18 py-wq-13 last:border-b-0"
    : [
        "wq-card",
        props.layout === "wall" ? "p-6" : "p-wq-17",
        props.on ? "border-wq-accent/40" : "",
      ]
)

const boxSize = computed(() =>
  props.layout === "wall" ? 56 : props.layout === "compact" ? 34 : 44
)
const iconSize = computed(() =>
  props.layout === "wall" ? 28 : props.layout === "compact" ? 18 : 20
)
</script>
