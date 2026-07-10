<template>
  <AppIcon :name="iconName" :size="size" />
</template>

<script setup lang="ts">
import { computed } from "vue"
import AppIcon from "@/components/AppIcon.vue"

// Maps a notification provider id to its icon symbol — one place to extend
// when a new provider is added, instead of updating every call site's own
// fallback/ternary individually.
const props = withDefaults(defineProps<{ provider: string; size?: number }>(), { size: 16 })

const iconName = computed(() => {
  switch (props.provider) {
    case "telegram":
      return "telegram"
    case "discord":
      return "discord"
    default:
      return "bell"
  }
})
</script>
