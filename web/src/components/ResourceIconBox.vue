<template>
  <div
    class="grid flex-none place-items-center rounded-xl border"
    :class="boxClass"
    :style="{ width: `${boxSize}px`, height: `${boxSize}px` }"
  >
    <AppIcon :name="icon" :size="iconSize" />
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue"
import AppIcon from "@/components/AppIcon.vue"

const props = defineProps<{ icon: string; on: boolean; boxSize: number; iconSize: number }>()

const boxClass = computed(() =>
  props.on
    ? "border-wq-accent/30 bg-wq-accent/10 text-wq-accent-2"
    : "border-wq-border bg-wq-panel-2 text-wq-muted"
)
</script>
