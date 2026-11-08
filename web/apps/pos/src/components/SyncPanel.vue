<script setup>
import { computed } from 'vue'
import { PosButton } from '../../../../packages/ui/src/index.js'
import { useConnectivityStore } from '../stores/connectivity.js'
import { useToastStore } from '../stores/toast.js'

const connectivity = useConnectivityStore()
const toast = useToastStore()
const stateLabel = computed(() => connectivity.online ? 'Online' : 'Offline')

async function syncNow() {
  const result = await connectivity.sync()
  if (result?.rejected) toast.warning(`${result.rejected} offline sale(s) need attention.`)
  else if (result?.accepted) toast.success(`${result.accepted} offline sale(s) synced.`)
  else toast.info('Nothing waiting to sync.')
}
</script>

<template>
  <section class="sync-panel" aria-label="Connection and sync status">
    <div>
      <strong><span class="sync-dot" :data-online="connectivity.online" /> {{ stateLabel }}</strong>
      <small v-if="connectivity.online">Sales are syncing with the server.</small>
      <small v-else>Sales can be saved offline and synced when you reconnect.</small>
    </div>
    <div class="sync-actions">
      <span v-if="connectivity.pending" class="sync-count">{{ connectivity.pending }} pending</span>
      <PosButton variant="secondary" :disabled="!connectivity.online || !connectivity.pending" @click="syncNow">Sync now</PosButton>
    </div>
  </section>
</template>

<style scoped>
.sync-panel { display:flex; justify-content:space-between; align-items:center; gap:16px; padding:14px 16px; border:1px solid #dfe9e5; border-radius:12px; background:#f8fcfa; }
.sync-panel > div:first-child { display:grid; gap:5px; }
.sync-panel small { color:#6f7e82; font-size:11px; }
.sync-dot { display:inline-block; width:8px; height:8px; border-radius:50%; background:#d27832; }
.sync-dot[data-online=true] { background:#20a36f; box-shadow:0 0 0 4px #dff4e9; }
.sync-actions { display:flex; align-items:center; gap:10px; }
.sync-count { color:#176c57; font-size:11px; font-weight:700; }
@media (max-width:600px) { .sync-panel { align-items:flex-start; flex-direction:column; } }
</style>
