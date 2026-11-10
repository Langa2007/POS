<template>
  <section class="manager">
    <header class="manager-head">
      <div><h3><PosIcon name="heart-pulse" size="20" /> Client Loyalty</h3><p>Find returning clients, apply rewards, and keep service history visible.</p></div>
      <button class="primary" @click="enroll">Enroll client</button>
    </header>
    <div class="lookup"><input v-model="query" placeholder="Search client name, phone, or membership number" @keydown.enter="search" /><button class="primary" :disabled="loading" @click="search">Find client</button></div>
    <p v-if="loading" class="state">Loading loyalty members…</p>
    <p v-else-if="error" class="state error" role="alert">{{ error }}</p>
    <p v-else-if="!visibleClients.length" class="state">No loyalty members found.</p>
    <div v-else class="client-grid">
      <article v-for="client in visibleClients" :key="client.id" class="client-card">
        <div class="client-head"><div class="avatar">{{ client.name[0] }}</div><div><h4>{{ client.name }}</h4><small>{{ client.phone }} · {{ client.membership_no }}</small></div><span class="tier">{{ client.tier }}</span></div>
        <div class="points"><strong>{{ Number(client.points).toLocaleString() }}</strong><span>points available</span></div>
        <div class="client-footer"><span>Last visit {{ formatDate(client.last_visit) }}</span><button class="secondary" :disabled="redeemingId === client.id || client.points <= 0" @click="redeem(client)">{{ redeemingId === client.id ? 'Applying…' : 'Apply reward' }}</button></div>
      </article>
    </div>
  </section>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { PosIcon } from '@pos/ui'
import { useApi } from '../../../../apps/pos/src/composables/useApi.js'

const emit = defineEmits(['notification'])
const api = useApi()
const query = ref('')
const clients = ref([])
const loading = ref(true)
const error = ref('')
const redeemingId = ref('')
const visibleClients = computed(() => clients.value)

onMounted(loadMembers)

async function loadMembers() {
  loading.value = true
  error.value = ''
  try {
    const data = await api.get(`/api/v1/salon/loyalty${query.value ? `?q=${encodeURIComponent(query.value)}` : ''}`)
    clients.value = data.members ?? []
  } catch (err) {
    error.value = err.message
    emit('notification', { message: `Loyalty members could not be loaded: ${err.message}`, type: 'error' })
  } finally {
    loading.value = false
  }
}

async function search() { await loadMembers() }

async function redeem(client) {
  const points = Math.min(1000, Number(client.points))
  if (!points) return
  redeemingId.value = client.id
  try {
    await api.post(`/api/v1/salon/loyalty/${client.id}/redeem`, { points })
    client.points -= points
    emit('notification', { message: `${points.toLocaleString()} points redeemed for ${client.name}`, type: 'success' })
  } catch (err) {
    emit('notification', { message: `Reward could not be applied: ${err.message}`, type: 'error' })
  } finally {
    redeemingId.value = ''
  }
}

async function enroll() {
  emit('notification', { message: 'Client enrollment requires a name and phone through the loyalty API.', type: 'info' })
}

function formatDate(value) {
  if (!value) return 'not recorded'
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleDateString()
}
</script>

<style scoped>
.manager { display:flex; flex-direction:column; gap:16px; }.manager-head { display:flex; justify-content:space-between; gap:12px; }.manager-head h3 { display:flex; align-items:center; gap:8px; margin:0; color:#1e293b; font-size:17px; }.manager-head p { margin:5px 0 0; color:#64748b; font-size:12px; }.primary,.secondary { border:0; border-radius:6px; padding:8px 11px; font-size:11px; font-weight:700; cursor:pointer; }.primary { background:#db2777; color:#fff; }.secondary { background:#fce7f3; color:#be185d; }.primary:disabled,.secondary:disabled { opacity:.6; cursor:wait; }.lookup { display:flex; gap:8px; }.lookup input { flex:1; padding:10px; border:1px solid #cbd5e1; border-radius:7px; }.state { padding:24px; border:1px dashed #cbd5e1; border-radius:8px; color:#64748b; text-align:center; font-size:12px; }.state.error { border-color:#fecaca; background:#fff7f7; color:#b91c1c; }.client-grid { display:grid; grid-template-columns:repeat(auto-fit,minmax(240px,1fr)); gap:10px; }.client-card { padding:14px; border:1px solid #e2e8f0; border-radius:9px; }.client-head { display:flex; align-items:center; gap:9px; }.avatar { display:grid; place-items:center; flex:0 0 34px; height:34px; border-radius:50%; background:#fce7f3; color:#be185d; font-weight:800; }.client-head h4 { margin:0; color:#1e293b; font-size:13px; }.client-head small { color:#94a3b8; font-size:10px; }.tier { margin-left:auto; padding:3px 6px; border-radius:4px; background:#fef3c7; color:#92400e; font-size:9px; font-weight:700; }.points { display:flex; align-items:baseline; gap:6px; margin:18px 0 12px; }.points strong { color:#db2777; font-size:25px; }.points span { color:#64748b; font-size:10px; }.client-footer { display:flex; justify-content:space-between; align-items:center; border-top:1px solid #f1f5f9; padding-top:10px; color:#94a3b8; font-size:10px; }
</style>
