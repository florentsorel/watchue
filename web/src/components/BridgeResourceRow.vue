<template>
  <div
    class="flex items-center gap-3.5 border-b border-wq-border-2 px-wq-17 py-3.5 last:border-b-0 hover:bg-wq-panel-2"
    :class="indent ? 'border-t border-dashed border-t-wq-border-2 py-2.5 pl-10' : ''"
  >
    <ResourceIconBox
      :icon="icon"
      :on="on"
      :box-size="indent ? 30 : 34"
      :icon-size="indent ? 16 : 18"
    />

    <div class="min-w-0 flex-1">
      <div class="text-sm font-semibold">{{ name }}</div>
      <div v-if="meta" class="mt-0.5 text-xs text-wq-muted">{{ meta }}</div>
    </div>

    <StatePill :on="on" />

    <button
      type="button"
      class="inline-flex h-wq-30 flex-none items-center gap-1.5 rounded-lg border border-wq-border px-2.5 text-xs font-semibold text-wq-muted transition-colors hover:border-wq-accent/40 hover:text-wq-text"
      :class="watched ? '!border-wq-accent/40 !bg-wq-accent/10 !text-wq-accent-2' : ''"
      @click="$emit('toggle-watch')"
    >
      <AppIcon :name="watched ? 'check' : 'plus'" :size="16" />
      <span v-if="!indent">{{ watched ? "Watching" : "Watch" }}</span>
    </button>

    <button
      v-if="expandable"
      type="button"
      class="grid h-wq-26 w-wq-26 flex-none place-items-center rounded-md text-wq-faint transition-colors hover:bg-wq-panel-2 hover:text-wq-text"
      @click="$emit('toggle-expand')"
    >
      <AppIcon
        name="chevron"
        :size="18"
        :class="expanded ? 'rotate-180' : ''"
        class="transition-transform"
      />
    </button>
  </div>
</template>

<script setup lang="ts">
import AppIcon from "@/components/AppIcon.vue"
import StatePill from "@/components/StatePill.vue"
import ResourceIconBox from "@/components/ResourceIconBox.vue"

withDefaults(
  defineProps<{
    name: string
    icon: string
    on: boolean
    watched: boolean
    meta?: string
    indent?: boolean
    expandable?: boolean
    expanded?: boolean
  }>(),
  { meta: "", indent: false, expandable: false, expanded: false }
)

defineEmits<{ "toggle-watch": []; "toggle-expand": [] }>()
</script>
