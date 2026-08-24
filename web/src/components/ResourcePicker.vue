<template>
  <div class="flex flex-wrap items-center gap-1.5">
    <button
      v-for="r in resources"
      :key="r.id"
      type="button"
      :aria-pressed="isSelected(r.id)"
      :disabled="!isSelected(r.id) && modelValue.length >= max"
      class="flex items-center gap-1.5 rounded-wq-9 border px-2.5 py-1 text-wq-12-5 font-semibold transition-colors disabled:cursor-not-allowed disabled:opacity-40"
      :class="
        isSelected(r.id)
          ? 'border-wq-border bg-wq-panel-2 text-wq-text'
          : 'border-transparent text-wq-muted hover:text-wq-text'
      "
      @click="toggle(r.id)"
    >
      <span
        class="h-2 w-2 rounded-full"
        :style="{ background: isSelected(r.id) ? colorOf(r.id) : 'var(--wq-faint)' }"
      />
      {{ r.name }}
    </button>
    <span v-if="modelValue.length >= max" class="text-wq-11 text-wq-faint">
      {{ max }} at a time
    </span>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  resources: Array<{ id: string; name: string }>
  modelValue: string[]
  max: number
  /** The color the chart gave this resource, so the chip matches its line. */
  colorOf: (id: string) => string
}>()

const emit = defineEmits<{ "update:modelValue": [value: string[]] }>()

function isSelected(id: string): boolean {
  return props.modelValue.includes(id)
}

function toggle(id: string): void {
  if (isSelected(id)) {
    // Never empty: an empty picker would leave the chart with nothing to draw
    // and no obvious way back.
    if (props.modelValue.length > 1)
      emit(
        "update:modelValue",
        props.modelValue.filter((v) => v !== id)
      )
    return
  }
  if (props.modelValue.length < props.max) emit("update:modelValue", [...props.modelValue, id])
}
</script>
