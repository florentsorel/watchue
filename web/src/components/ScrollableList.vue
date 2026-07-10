<template>
  <div class="overflow-y-auto" :style="{ maxHeight: `${itemsDisplayed * itemHeight}px` }">
    <slot />
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{ itemsDisplayed: number; itemHeight?: number }>(), {
  itemHeight: 67,
})
</script>
