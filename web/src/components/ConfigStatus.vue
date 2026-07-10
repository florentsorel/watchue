<template>
  <span
    class="inline-flex flex-none items-center gap-1.5 text-xs font-semibold"
    :class="ok ? 'text-wq-good' : 'text-wq-faint'"
  >
    <AppIcon v-if="ok" name="check" :size="14" />
    {{ label }}
  </span>
</template>

<script setup lang="ts">
import AppIcon from "@/components/AppIcon.vue"

defineProps<{ ok: boolean; label: string }>()
</script>
