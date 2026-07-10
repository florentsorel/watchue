<template>
  <span
    class="inline-flex flex-none items-center gap-1.5 rounded-full border border-wq-border bg-wq-panel-2 px-2.5 py-1 text-wq-11-5 font-semibold text-wq-muted"
    :class="on ? '!border-wq-accent/30 !bg-wq-accent/10 !text-wq-accent-2' : ''"
  >
    <span class="h-wq-7 w-wq-7 rounded-full bg-current" />
    {{ on ? "On" : "Off" }}
  </span>
</template>

<script setup lang="ts">
defineProps<{ on: boolean }>()
</script>
