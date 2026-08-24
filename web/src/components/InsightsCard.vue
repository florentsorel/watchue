<template>
  <div class="wq-card p-5">
    <div class="mb-4 flex flex-wrap items-baseline justify-between gap-2">
      <div>
        <h3 class="text-wq-15 font-semibold">{{ title }}</h3>
        <p v-if="hint" class="mt-0.5 text-wq-11-5 text-wq-faint">{{ hint }}</p>
      </div>
      <div v-if="value" class="font-mono-ui text-wq-13 font-semibold text-wq-accent-2">
        {{ value }}
      </div>
    </div>
    <slot />
  </div>
</template>

<script setup lang="ts">
defineProps<{ title: string; hint?: string; value?: string }>()
</script>
