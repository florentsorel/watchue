<template>
  <div class="mb-wq-18 flex flex-wrap items-end justify-between gap-4">
    <div>
      <div
        class="font-mono-ui mb-wq-9 flex items-center gap-2 text-wq-11 font-semibold tracking-wq-wide text-wq-accent-2 uppercase"
      >
        <span class="h-wq-1-5 w-4 rounded bg-wq-accent" />
        {{ kicker }}
      </div>
      <h2 class="text-2xl font-bold tracking-tight">{{ title }}</h2>
      <p v-if="subtitle" class="mt-1.5 max-w-wq-52ch text-wq-13-5 text-wq-muted">{{ subtitle }}</p>
    </div>
    <slot />
  </div>
</template>

<script setup lang="ts">
defineProps<{ kicker: string; title: string; subtitle?: string }>()
</script>
