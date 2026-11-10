<template>
  <div class="salon-view">
    <header class="topbar"><div><div class="eyebrow"><PosIcon name="sparkles" size="16" /> SALON &amp; SPA WORKSTATION</div><h2>Appointments, services, and client care</h2><p>Coordinate bookings, service packages, stylist commissions, and loyalty from the front desk.</p></div><div v-if="dashboardLoading" class="kpis"><div><small>Loading dashboard</small><strong>…</strong></div></div><div v-else-if="dashboardError" class="dashboard-error">{{ dashboardError }}</div><div v-else class="kpis"><div><small>Today's bookings</small><strong>{{ dashboard.bookings }}</strong></div><div><small>In service</small><strong>{{ dashboard.in_service }}</strong></div><div><small>Service revenue</small><strong>{{ formatMoney(dashboard.revenue_minor) }}</strong></div><div><small>Commission due</small><strong>{{ formatMoney(dashboard.commission_minor) }}</strong></div></div></header>
    <div v-if="notification" class="notice"><PosIcon name="check-circle" size="15" /> {{ notification.message }}<button @click="notification = null">×</button></div>
    <nav class="tabs"><button v-for="tab in tabs" :key="tab.id" :class="{ active: activeTab === tab.id }" @click="activeTab = tab.id"><PosIcon :name="tab.icon" size="16" />{{ tab.label }}</button></nav>
    <main class="panel">
      <AppointmentManager v-if="activeTab === 'appointments'" @notification="showNotification" />
      <ServicePackageManager v-else-if="activeTab === 'services'" @notification="showNotification" />
      <CommissionManager v-else-if="activeTab === 'commissions'" @notification="showNotification" />
      <ClientLoyaltyManager v-else @notification="showNotification" />
    </main>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { PosIcon } from '@pos/ui'
import { useApi } from '../../../../apps/pos/src/composables/useApi.js'
import AppointmentManager from '../components/AppointmentManager.vue'
import ServicePackageManager from '../components/ServicePackageManager.vue'
import CommissionManager from '../components/CommissionManager.vue'
import ClientLoyaltyManager from '../components/ClientLoyaltyManager.vue'
const activeTab = ref('appointments')
const notification = ref(null)
const api = useApi()
const dashboard = ref({ bookings: 0, in_service: 0, revenue_minor: 0, commission_minor: 0 })
const dashboardLoading = ref(true)
const dashboardError = ref('')
const tabs = [{ id:'appointments', label:'Appointments', icon:'calendar' }, { id:'services', label:'Services & Packages', icon:'sparkles' }, { id:'commissions', label:'Stylist Commissions', icon:'dollar-sign' }, { id:'loyalty', label:'Client Loyalty', icon:'heart-pulse' }]
onMounted(async () => { try { dashboard.value = await api.get('/api/v1/salon/dashboard') } catch (err) { dashboardError.value = err.message } finally { dashboardLoading.value = false } })
function formatMoney(minor) { return `KES ${(Number(minor) / 100).toLocaleString()}` }
function showNotification(value) { notification.value = value; setTimeout(() => { if (notification.value === value) notification.value = null }, 4000) }
</script>

<style scoped>
.salon-view { display:flex; flex-direction:column; gap:18px; padding:8px 0; }.topbar { display:flex; justify-content:space-between; align-items:flex-start; flex-wrap:wrap; gap:18px; padding:18px 20px; border:1px solid #e2e8f0; border-radius:12px; background:linear-gradient(135deg,#fff1f2,#fff); }.eyebrow { display:flex; align-items:center; gap:7px; color:#db2777; font-size:11px; font-weight:800; letter-spacing:.06em; }.topbar h2 { margin:7px 0 3px; color:#1e293b; font-size:20px; }.topbar p { margin:0; color:#64748b; font-size:12px; }.kpis { display:flex; gap:9px; flex-wrap:wrap; }.kpis div { min-width:105px; padding:10px 12px; border:1px solid #fce7f3; border-radius:8px; background:#fff; }.kpis small,.kpis strong { display:block; }.kpis small { color:#64748b; font-size:10px; }.kpis strong { margin-top:4px; color:#db2777; font-size:16px; }.notice { display:flex; align-items:center; gap:8px; padding:10px 13px; border:1px solid #bbf7d0; border-radius:7px; background:#f0fdf4; color:#166534; font-size:12px; }.notice button { margin-left:auto; border:0; background:none; color:inherit; font-size:18px; cursor:pointer; }.tabs { display:flex; gap:8px; overflow-x:auto; border-bottom:2px solid #e2e8f0; }.tabs button { display:inline-flex; align-items:center; gap:7px; padding:10px 14px; border:0; border-bottom:2px solid transparent; margin-bottom:-2px; background:transparent; color:#64748b; font-size:12px; font-weight:700; white-space:nowrap; cursor:pointer; }.tabs button.active { border-bottom-color:#db2777; color:#db2777; }.panel { min-height:420px; padding:20px; border:1px solid #e2e8f0; border-radius:12px; background:#fff; }
.dashboard-error { padding:12px; border:1px solid #fecaca; border-radius:8px; color:#b91c1c; font-size:12px; }
</style>
