<template>
  <div class="inline-flex gap-0.5 rounded-wq-10 border border-wq-border bg-wq-panel-2 p-wq-3">
    <button
      v-for="option in options"
      :key="option.value"
      type="button"
      class="rounded-wq-7 px-3 py-1.5 text-wq-12-5 font-semibold text-wq-muted transition-colors hover:text-wq-text"
      :class="option.value === modelValue ? 'wq-card !shadow-sm text-wq-text' : ''"
      @click="$emit('update:modelValue', option.value)"
    >
      {{ option.label }}
    </button>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  modelValue: string
  options: Array<{ label: string; value: string }>
}>()

defineEmits<{ "update:modelValue": [value: string] }>()
</script>
