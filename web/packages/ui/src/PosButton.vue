<script setup>
defineProps({ variant: { type: String, default: 'primary' }, type: { type: String, default: 'button' } })
</script>
<template>
  <button class="pos-button" :class="`pos-button--${variant}`" :type="type"><slot /></button>
</template>
<style scoped>
.pos-button { min-height: 44px; border: 0; border-radius: 10px; padding: 0 18px; font: inherit; font-weight: 650; cursor: pointer; }
.pos-button--primary { background: var(--color-brand); color: white; }
.pos-button--secondary { background: var(--color-surface-muted); color: var(--color-ink); }
.pos-button:focus-visible { outline: 3px solid var(--color-focus); outline-offset: 2px; }
</style>
