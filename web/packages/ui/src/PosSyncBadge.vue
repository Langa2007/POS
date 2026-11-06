<!-- Connectivity / sync status badge for the shell topbar -->
<script setup>
defineProps({
  online:      { type: Boolean, default: true },
  pending:     { type: Number, default: 0 },
  rejected:    { type: Number, default: 0 },
})
</script>

<template>
  <div class="sync-badge" :class="{ 'sync-badge--offline': !online, 'sync-badge--error': rejected > 0 }"
    :aria-label="online ? 'Connected' : 'Offline'">
    <span class="sync-badge__dot" aria-hidden="true" />
    <span v-if="!online">Offline<template v-if="pending"> · {{ pending }} queued</template></span>
    <span v-else-if="rejected > 0">{{ rejected }} sync error{{ rejected > 1 ? 's' : '' }}</span>
    <span v-else-if="pending > 0">{{ pending }} syncing…</span>
    <span v-else>Online</span>
  </div>
</template>

<style scoped>
.sync-badge {
  display: inline-flex; align-items: center; gap: 7px;
  font-size: 11px; font-weight: 650;
  color: #16765d; background: #e5f5ee;
  padding: 8px 12px; border-radius: 9px;
}
.sync-badge--offline { color: #7a5c08; background: #fffbea; }
.sync-badge--error   { color: #8b1a1a; background: #fff1f0; }
.sync-badge__dot {
  width: 7px; height: 7px; border-radius: 50%;
  background: currentColor; flex-shrink: 0;
}
</style>

