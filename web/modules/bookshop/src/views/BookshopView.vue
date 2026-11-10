<template>
  <div class="bookshop-view">
    <header class="topbar"><div><div class="eyebrow"><PosIcon name="file-text" size="16" /> BOOKSHOP WORKSTATION</div><h2>Books, orders, and returns</h2><p>ISBN-aware catalog management, institutional ordering, consignment, and publisher operations.</p></div><div class="kpis"><div><small>Today's sales</small><strong>KES 86,400</strong></div><div><small>Open orders</small><strong>12</strong></div><div><small>Titles in stock</small><strong>4,281</strong></div><div><small>Return value</small><strong>KES 15.7k</strong></div></div></header>
    <div v-if="notification" class="notice"><PosIcon name="check-circle" size="15" /> {{ notification.message }}<button @click="notification = null">×</button></div>
    <nav class="tabs"><button v-for="tab in tabs" :key="tab.id" :class="{ active: activeTab === tab.id }" @click="activeTab = tab.id"><PosIcon :name="tab.icon" size="16" />{{ tab.label }}</button></nav>
    <main class="panel">
      <BookCatalogManager v-if="activeTab === 'catalog'" @notification="showNotification" />
      <BookOrderManager v-else-if="activeTab === 'orders'" @notification="showNotification" />
      <PublisherReturnsManager v-else-if="activeTab === 'returns'" @notification="showNotification" />
      <ConsignmentManager v-else @notification="showNotification" />
    </main>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { PosIcon } from '@pos/ui'
import BookCatalogManager from '../components/BookCatalogManager.vue'
import BookOrderManager from '../components/BookOrderManager.vue'
import PublisherReturnsManager from '../components/PublisherReturnsManager.vue'
import ConsignmentManager from '../components/ConsignmentManager.vue'
const activeTab = ref('catalog')
const notification = ref(null)
const tabs = [{ id:'catalog', label:'ISBN Catalog', icon:'search' }, { id:'orders', label:'Special & School Orders', icon:'clipboard-list' }, { id:'returns', label:'Publisher Returns', icon:'refresh-cw' }, { id:'stock', label:'Consignment & Second-hand', icon:'package' }]
function showNotification(value) { notification.value = value; setTimeout(() => { if (notification.value === value) notification.value = null }, 4000) }
</script>

<style scoped>
.bookshop-view { display:flex; flex-direction:column; gap:18px; padding:8px 0; }.topbar { display:flex; justify-content:space-between; align-items:flex-start; flex-wrap:wrap; gap:18px; padding:18px 20px; border:1px solid #e2e8f0; border-radius:12px; background:linear-gradient(135deg,#faf5ff,#fff); }.eyebrow { display:flex; align-items:center; gap:7px; color:#7c3aed; font-size:11px; font-weight:800; letter-spacing:.06em; }.topbar h2 { margin:7px 0 3px; color:#1e293b; font-size:20px; }.topbar p { margin:0; color:#64748b; font-size:12px; }.kpis { display:flex; gap:9px; flex-wrap:wrap; }.kpis div { min-width:105px; padding:10px 12px; border:1px solid #ede9fe; border-radius:8px; background:#fff; }.kpis small,.kpis strong { display:block; }.kpis small { color:#64748b; font-size:10px; }.kpis strong { margin-top:4px; color:#7c3aed; font-size:16px; }.notice { display:flex; align-items:center; gap:8px; padding:10px 13px; border:1px solid #bbf7d0; border-radius:7px; background:#f0fdf4; color:#166534; font-size:12px; }.notice button { margin-left:auto; border:0; background:none; color:inherit; font-size:18px; cursor:pointer; }.tabs { display:flex; gap:8px; overflow-x:auto; border-bottom:2px solid #e2e8f0; }.tabs button { display:inline-flex; align-items:center; gap:7px; padding:10px 14px; border:0; border-bottom:2px solid transparent; margin-bottom:-2px; background:transparent; color:#64748b; font-size:12px; font-weight:700; white-space:nowrap; cursor:pointer; }.tabs button.active { border-bottom-color:#7c3aed; color:#7c3aed; }.panel { min-height:420px; padding:20px; border:1px solid #e2e8f0; border-radius:12px; background:#fff; }
</style>
