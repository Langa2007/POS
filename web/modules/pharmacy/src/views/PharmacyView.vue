<template>
  <div class="pharmacy-view">
    <header class="view-header">
      <div><span class="shop-badge"><PosIcon name="pill" size="14" /> PHARMACY DISPENSARY SUITE</span><h1>Pharmacy Dispensary &amp; Clinical Counter</h1><p class="view-description">Prescription intake, batch traceability, controlled-drug records, substitutions, and clinical labels.</p></div>
      <div class="quick-kpis"><div class="kpi-card"><span>Catalog Items</span><strong>{{ summary.catalog }}</strong></div><div class="kpi-card"><span>Prescription Queue</span><strong>{{ summary.prescriptions }}</strong></div><div class="kpi-card"><span>Near-Expiry Batches</span><strong>{{ summary.near_expiry }}</strong></div></div>
    </header>
    <div v-if="notification" class="inline-notice" :class="notification.type"><PosIcon name="info" size="15" /><span>{{ notification.message }}</span><button @click="notification = null"><PosIcon name="x" size="14" /></button></div>
    <nav class="pharmacy-tabs"><button v-for="tab in tabs" :key="tab.id" type="button" class="tab-btn" :class="{ active: activeTab === tab.id }" @click="activeTab = tab.id"><PosIcon :name="tab.icon" size="16" />{{ tab.label }}</button></nav>
    <main class="tab-content-panel">
      <PrescriptionStation v-if="activeTab === 'prescriptions'" @add-item="handleAddItem" @notification="showNotice" />
      <BatchFefoManager v-else-if="activeTab === 'batches'" @notification="showNotice" />
      <ControlledDrugsRegister v-else-if="activeTab === 'controlled'" @add-item="handleAddItem" @notification="showNotice" />
      <GenericSubstitutionFinder v-else-if="activeTab === 'generics'" @add-item="handleAddItem" @notification="showNotice" />
      <DispensingLabelGenerator v-else @notification="showNotice" />
    </main>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { PosIcon } from '@pos/ui'
import { useApi } from '../../../../apps/pos/src/composables/useApi.js'
import PrescriptionStation from '../components/PrescriptionStation.vue'
import BatchFefoManager from '../components/BatchFefoManager.vue'
import ControlledDrugsRegister from '../components/ControlledDrugsRegister.vue'
import GenericSubstitutionFinder from '../components/GenericSubstitutionFinder.vue'
import DispensingLabelGenerator from '../components/DispensingLabelGenerator.vue'
const api = useApi()
const activeTab = ref('prescriptions')
const notification = ref(null)
const summary = ref({ catalog: 0, prescriptions: 0, near_expiry: 0 })
const tabs = [{ id: 'prescriptions', label: 'Prescriptions & Intake', icon: 'clipboard-list' }, { id: 'batches', label: 'Batch Tracking & FEFO', icon: 'layers' }, { id: 'controlled', label: 'Controlled Drugs Register', icon: 'lock' }, { id: 'generics', label: 'Generic Substitution', icon: 'pill' }, { id: 'labels', label: 'Dispensing Labels', icon: 'tag' }]
onMounted(async () => { try { const [catalog, prescriptions, batches] = await Promise.all([api.get('/api/v1/pharmacy/catalog'), api.get('/api/v1/pharmacy/prescriptions'), api.get('/api/v1/pharmacy/batches')]); summary.value = { catalog: catalog.items?.length ?? 0, prescriptions: prescriptions.items?.length ?? 0, near_expiry: batches.items?.filter(item => item.days_left > 0 && item.days_left <= 60).length ?? 0 } } catch { showNotice({ message: 'Pharmacy summary is unavailable. Please contact your administrator.', type: 'error' }) } })
function showNotice(payload) { notification.value = payload; setTimeout(() => { if (notification.value === payload) notification.value = null }, 5000) }
function handleAddItem(item) { window.dispatchEvent(new CustomEvent('pos:add-item', { detail: item })); showNotice({ message: 'Item added to the active sale', type: 'success' }) }
</script>

<style scoped>
.pharmacy-view{display:flex;flex-direction:column;gap:1.5rem}.view-header{display:flex;justify-content:space-between;align-items:flex-start;flex-wrap:wrap;gap:1rem;border-bottom:1px solid #e2e8f0;padding-bottom:1.25rem}.shop-badge{display:inline-flex;align-items:center;gap:5px;font-size:.75rem;letter-spacing:.08em;font-weight:800;color:#0369a1;background:#e0f2fe;padding:.2rem .6rem;border-radius:4px}.header-main h1{margin:.5rem 0 .4rem;font-size:1.85rem;color:#0f172a}.view-description{margin:0;color:#64748b;font-size:.95rem;max-width:680px}.quick-kpis{display:flex;gap:.75rem;flex-wrap:wrap}.kpi-card{background:#f8fafc;border:1px solid #e2e8f0;padding:.75rem 1rem;border-radius:8px;display:flex;flex-direction:column;gap:.25rem;min-width:120px}.kpi-card span{font-size:.75rem;color:#64748b}.kpi-card strong{font-size:1.15rem;color:#0f172a}.inline-notice{display:flex;align-items:center;gap:8px;padding:.75rem 1.25rem;border-radius:8px;font-size:.9rem}.inline-notice.success{background:#dcfce7;color:#166534}.inline-notice.error{background:#fee2e2;color:#991b1b}.inline-notice button{margin-left:auto;border:0;background:transparent;color:inherit;cursor:pointer}.pharmacy-tabs{display:flex;gap:.5rem;border-bottom:2px solid #e2e8f0;overflow-x:auto}.tab-btn{display:inline-flex;align-items:center;gap:7px;background:none;border:0;padding:.75rem 1.15rem;font-size:.95rem;font-weight:700;color:#64748b;cursor:pointer;white-space:nowrap}.tab-btn.active{color:#0284c7;border-bottom:2px solid #0284c7}.tab-content-panel{min-height:400px}
</style>
