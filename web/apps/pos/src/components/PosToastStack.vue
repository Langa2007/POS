<script setup>
import { useToastStore } from '../stores/toast.js'
import { PosToast } from '../../../../packages/ui/src/index.js'

const toast = useToastStore()
</script>

<template>
  <div class="toast-stack" aria-live="polite">
    <PosToast :toasts="toast.toasts" @dismiss="toast.dismiss" />
  </div>
</template>

<style scoped>
.toast-stack { position:fixed; right:24px; bottom:24px; z-index:20; display:grid; gap:8px; width:min(360px, calc(100vw - 32px)); }
</style>
