<template>
  <div>
    <IconSprite />
    <RouterView v-slot="{ Component }">
      <KeepAlive>
        <component :is="Component" />
      </KeepAlive>
    </RouterView>
  </div>
</template>

<script setup lang="ts">
import IconSprite from "@/components/IconSprite.vue"
</script>
