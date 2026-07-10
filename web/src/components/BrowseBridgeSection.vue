<template>
  <section class="py-wq-34">
    <SectionHeading
      kicker="Choose what to watch"
      title="Browse your bridge"
      subtitle="Zones group lights across rooms; rooms contain the physical lights."
    />
    <BridgeOfflineBanner v-if="!settings.bridgeOnline" />
    <div class="grid grid-cols-1 items-start gap-4 lg:grid-cols-2">
      <ZonesPanel />
      <RoomsPanel />
    </div>
  </section>
</template>

<script setup lang="ts">
import SectionHeading from "@/components/SectionHeading.vue"
import ZonesPanel from "@/components/ZonesPanel.vue"
import RoomsPanel from "@/components/RoomsPanel.vue"
import BridgeOfflineBanner from "@/components/BridgeOfflineBanner.vue"
import { useSettingsStore } from "@/stores/useSettingsStore"

const settings = useSettingsStore()
</script>
