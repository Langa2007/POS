<template>
  <HardwareCheckoutSidebar :cart="cart" />
</template>

<script setup>
import HardwareCheckoutSidebar from './HardwareCheckoutSidebar.vue'

defineProps({
  cart: {
    type: Array,
    default: () => [],
  },
})
</script>
