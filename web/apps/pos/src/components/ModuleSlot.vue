<script setup>
import { computed, defineAsyncComponent } from 'vue'
import { getSlotContributions } from '../moduleLoader.js'

const props = defineProps({
  name: { type: String, required: true },
  activeModule: { type: String, default: null },
  slotProps: { type: Object, default: () => ({}) },
})

const contributions = computed(() => {
  return getSlotContributions(props.name, props.activeModule).map((c) => {
    // If component is a function returning dynamic import, wrap in defineAsyncComponent
    const comp = typeof c.component === 'function' && !c.component.render
      ? defineAsyncComponent(c.component)
      : c.component
    return {
      moduleId: c.moduleId,
      component: comp,
    }
  })
})
</script>

<template>
  <div v-if="contributions.length" class="module-slot-container" :data-slot="name">
    <div
      v-for="item in contributions"
      :key="item.moduleId"
      class="module-slot-item"
      :data-module="item.moduleId"
    >
      <component :is="item.component" v-bind="slotProps" />
    </div>
  </div>
</template>

<style scoped>
.module-slot-container {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.module-slot-item {
  width: 100%;
}
</style>
