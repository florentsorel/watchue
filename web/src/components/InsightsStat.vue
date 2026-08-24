<template>
  <div class="wq-card p-4">
    <div class="flex items-center gap-1.5 text-wq-11 font-semibold text-wq-muted">
      <AppIcon :name="icon" :size="14" />
      {{ label }}
    </div>
    <!-- Proportional figures, not tabular: these stand alone rather than in a
         column, and equal-width digits read loose at this size. -->
    <div class="mt-2 text-wq-26 leading-none font-extrabold tracking-tight">{{ value }}</div>
    <div v-if="hint" class="mt-1.5 text-wq-11 text-wq-faint">{{ hint }}</div>
  </div>
</template>

<script setup lang="ts">
import AppIcon from "@/components/AppIcon.vue"

defineProps<{ icon: string; label: string; value: string; hint?: string }>()
</script>
