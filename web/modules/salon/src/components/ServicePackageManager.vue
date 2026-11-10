<template>
  <section class="manager">
    <header class="manager-head"><div><h3><PosIcon name="sparkles" size="20" /> Services &amp; Packages</h3><p>Configure treatment durations, bundled visits, add-ons, and package balances.</p></div><button class="primary" @click="addPackage">New package</button></header>
    <p v-if="loading" class="state">Loading service packages…</p><p v-else-if="error" class="state error">{{ error }}</p><p v-else-if="!packages.length" class="state">No active service packages configured.</p>
    <div v-else class="package-grid"><article v-for="pkg in packages" :key="pkg.id" class="package-card"><div class="package-icon"><PosIcon name="sparkles" size="20" /></div><div class="package-main"><div class="package-head"><h4>{{ pkg.name }}</h4><span class="active" v-if="pkg.active">Active</span></div><p>{{ pkg.description }}</p><div class="package-footer"><strong>{{ formatMoney(pkg.price_minor) }}</strong><span>{{ pkg.visits }} visits · {{ pkg.duration_minutes }} min each</span></div><button class="secondary" @click="sell(pkg)">Sell package</button></div></article></div>
  </section>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { PosIcon } from '@pos/ui'
import { useApi } from '../../../../apps/pos/src/composables/useApi.js'
const emit = defineEmits(['notification'])
const api = useApi()
const packages = ref([])
const loading = ref(true)
const error = ref('')
onMounted(loadPackages)
async function loadPackages() { try { const data = await api.get('/api/v1/salon/packages'); packages.value = data.packages ?? [] } catch (err) { error.value = err.message; emit('notification', { message: `Packages could not be loaded: ${err.message}`, type: 'error' }) } finally { loading.value = false } }
function formatMoney(minor) { return `KES ${(Number(minor) / 100).toLocaleString()}` }
function sell(pkg) { emit('notification', { message:`${pkg.name} added to checkout`, type:'success' }); window.dispatchEvent(new CustomEvent('pos:add-item', { detail:{ id:pkg.id, sku:`PKG-${pkg.id}`, name:pkg.name, price_minor:Number(pkg.price_minor), unit_price_minor:Number(pkg.price_minor), quantity:1, kind:'service' } })) }
function addPackage() { emit('notification', { message:'Create the package through the salon package API.', type:'info' }) }
</script>

<style scoped>
.manager { display:flex; flex-direction:column; gap:16px; }.manager-head { display:flex; justify-content:space-between; gap:12px; }.manager-head h3 { display:flex; align-items:center; gap:8px; margin:0; color:#1e293b; font-size:17px; }.manager-head p { margin:5px 0 0; color:#64748b; font-size:12px; }.primary,.secondary { border:0; border-radius:6px; padding:8px 11px; font-size:11px; font-weight:700; cursor:pointer; }.primary { background:#db2777; color:#fff; }.secondary { background:#fce7f3; color:#be185d; }.state { padding:24px; border:1px dashed #cbd5e1; border-radius:8px; color:#64748b; text-align:center; font-size:12px; }.state.error { border-color:#fecaca; color:#b91c1c; }.package-grid { display:grid; grid-template-columns:repeat(auto-fit,minmax(240px,1fr)); gap:12px; }.package-card { display:flex; gap:12px; padding:14px; border:1px solid #f1f5f9; border-radius:9px; }.package-icon { display:grid; place-items:center; flex:0 0 40px; height:40px; border-radius:8px; background:#fce7f3; color:#db2777; }.package-main { flex:1; }.package-head { display:flex; justify-content:space-between; gap:8px; }.package-head h4 { margin:0; color:#1e293b; font-size:14px; }.active { padding:3px 6px; border-radius:4px; background:#dcfce7; color:#166534; font-size:9px; font-weight:700; }.package-main p { margin:6px 0 12px; color:#64748b; font-size:11px; line-height:1.4; }.package-footer { display:flex; justify-content:space-between; gap:8px; align-items:center; margin-bottom:10px; }.package-footer strong { color:#db2777; font-size:15px; }.package-footer span { color:#94a3b8; font-size:10px; }
</style>
