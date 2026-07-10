<template>
  <svg
    class="inline-block flex-none stroke-current"
    :style="{ width: `${size}px`, height: `${size}px` }"
    fill="none"
    :stroke-width="strokeWidth"
    stroke-linecap="round"
    stroke-linejoin="round"
    aria-hidden="true"
  >
    <use :href="`#i-${name}`" />
  </svg>
</template>

<script setup lang="ts">
import { computed } from "vue"

const props = withDefaults(defineProps<{ name: string; size?: number }>(), {
  size: 20,
})

const strokeWidth = computed(() => (props.size >= 32 ? 1.5 : 1.7))
</script>
