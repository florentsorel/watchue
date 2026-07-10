<template>
  <div
    class="mb-4 flex items-center gap-2.5 rounded-xl border border-red-400/30 bg-red-500/10 px-4 py-3 text-sm text-red-500"
  >
    <AppIcon name="warning" :size="18" class="flex-none" />
    <span>{{ message }}</span>
  </div>
</template>

<script setup lang="ts">
import AppIcon from "@/components/AppIcon.vue"

withDefaults(defineProps<{ message?: string }>(), {
  message: "Can't reach the Hue bridge right now — showing the last known state.",
})
</script>
