<template>
  <section class="manager">
    <header class="manager-head"><div><h3><PosIcon name="dollar-sign" size="20" /> Stylist Commissions</h3><p>Review service and product performance before the next payroll cut-off.</p></div><button class="primary" @click="closePeriod">Close period</button></header>
    <p v-if="loading" class="state">Loading commission data…</p><p v-else-if="error" class="state error">{{ error }}</p><p v-else-if="!stylists.length" class="state">No commission records are available for the current period.</p>
    <div v-else class="commission-grid"><article v-for="stylist in stylists" :key="stylist.staff_id" class="commission-card"><div class="avatar">{{ stylist.name[0] }}</div><div class="details"><h4>{{ stylist.name }}</h4><small>{{ stylist.services }} services</small><div class="line"><span>Service sales</span><strong>{{ formatMoney(stylist.sales_minor) }}</strong></div><div class="line"><span>Commission due</span><strong class="pink">{{ formatMoney(stylist.commission_minor) }}</strong></div></div></article></div>
  </section>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { PosIcon } from '@pos/ui'
import { useApi } from '../../../../apps/pos/src/composables/useApi.js'
const emit = defineEmits(['notification'])
const api = useApi()
const stylists = ref([])
const loading = ref(true)
const error = ref('')
onMounted(loadCommissions)
async function loadCommissions() { try { const data = await api.get('/api/v1/salon/commissions'); stylists.value = data.commissions ?? [] } catch (err) { error.value = err.message; emit('notification', { message:`Commissions could not be loaded: ${err.message}`, type:'error' }) } finally { loading.value = false } }
function formatMoney(minor) { return `KES ${(Number(minor) / 100).toLocaleString()}` }
function closePeriod() { emit('notification', { message:'Commission period closing is managed by the admin.', type:'info' }) }
</script>

<style scoped>
.manager { display:flex; flex-direction:column; gap:16px; }.manager-head { display:flex; justify-content:space-between; gap:12px; }.manager-head h3 { display:flex; align-items:center; gap:8px; margin:0; color:#1e293b; font-size:17px; }.manager-head p { margin:5px 0 0; color:#64748b; font-size:12px; }.primary { border:0; border-radius:6px; padding:8px 11px; background:#db2777; color:#fff; font-size:11px; font-weight:700; cursor:pointer; }.state { padding:24px; border:1px dashed #cbd5e1; border-radius:8px; color:#64748b; text-align:center; font-size:12px; }.state.error { border-color:#fecaca; color:#b91c1c; }.commission-grid { display:grid; grid-template-columns:repeat(auto-fit,minmax(230px,1fr)); gap:12px; }.commission-card { display:flex; gap:12px; padding:15px; border:1px solid #e2e8f0; border-radius:9px; }.avatar { display:grid; place-items:center; flex:0 0 38px; height:38px; border-radius:50%; background:#fce7f3; color:#be185d; font-weight:800; }.details { flex:1; }.details h4 { margin:0; color:#1e293b; font-size:14px; }.details small { color:#94a3b8; font-size:10px; }.line { display:flex; justify-content:space-between; margin-top:12px; color:#64748b; font-size:11px; }.line strong { color:#1e293b; }.pink { color:#db2777 !important; }
</style>
