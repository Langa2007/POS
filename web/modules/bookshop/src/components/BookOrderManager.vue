<template>
  <section class="manager">
    <header class="manager-head"><div><h3><PosIcon name="clipboard-list" size="20" /> Special &amp; School Orders</h3><p>Track deposits, supplier fulfilment, and customer collection status.</p></div><button class="primary" @click="newOrder">New order</button></header>
    <div class="stats"><div><small>Open orders</small><strong>{{ orders.filter(o => o.status !== 'collected').length }}</strong></div><div><small>Deposits held</small><strong>KES {{ deposits.toLocaleString() }}</strong></div><div><small>Ready today</small><strong>{{ orders.filter(o => o.status === 'ready').length }}</strong></div></div>
    <div class="order-list"><article v-for="order in orders" :key="order.id" class="order-card"><div><strong>{{ order.customer }}</strong><p>{{ order.id }} · {{ order.type }} · {{ order.lines }} titles</p></div><div class="order-meta"><span class="status" :class="order.status">{{ formatStatus(order.status) }}</span><strong>KES {{ order.total.toLocaleString() }}</strong><button class="secondary" @click="advance(order)">Advance</button></div></article></div>
  </section>
</template>

<script setup>
import { ref } from 'vue'
import { PosIcon } from '@pos/ui'
const emit = defineEmits(['notification'])
const orders = ref([
  { id: 'ORD-2401', customer: 'Greenfield Academy', type: 'School order', lines: 38, total: 48600, status: 'awaiting_supplier' },
  { id: 'ORD-2402', customer: 'Mary Wanjiku', type: 'Special order', lines: 2, total: 2800, status: 'ready' },
  { id: 'ORD-2403', customer: 'Nairobi Reading Club', type: 'School order', lines: 16, total: 19400, status: 'deposit_paid' },
  { id: 'ORD-2398', customer: 'Peter Omondi', type: 'Special order', lines: 1, total: 1200, status: 'collected' },
])
const deposits = 29800
function formatStatus(status) { return status.replace('_', ' ') }
function advance(order) { const states = ['deposit_paid', 'awaiting_supplier', 'ready', 'collected']; order.status = states[(states.indexOf(order.status) + 1) % states.length]; emit('notification', { message: `${order.id} moved to ${formatStatus(order.status)}`, type: 'success' }) }
function newOrder() { emit('notification', { message: 'New book order form opened', type: 'info' }) }
</script>

<style scoped>
.manager { display:flex; flex-direction:column; gap:16px; }.manager-head { display:flex; justify-content:space-between; gap:12px; }.manager-head h3 { display:flex; align-items:center; gap:8px; margin:0; font-size:17px; color:#1e293b; }.manager-head p { margin:5px 0 0; color:#64748b; font-size:12px; }.primary,.secondary { border:0; border-radius:6px; padding:8px 11px; font-size:11px; font-weight:700; cursor:pointer; }.primary { background:#7c3aed; color:#fff; }.secondary { background:#f3e8ff; color:#6d28d9; }.stats { display:grid; grid-template-columns:repeat(3,1fr); gap:10px; }.stats div { padding:12px; border:1px solid #ede9fe; border-radius:8px; background:#faf5ff; }.stats small,.stats strong { display:block; }.stats small { color:#64748b; font-size:10px; }.stats strong { margin-top:4px; color:#7c3aed; font-size:18px; }.order-list { display:flex; flex-direction:column; gap:8px; }.order-card { display:flex; justify-content:space-between; gap:12px; align-items:center; padding:13px; border:1px solid #e2e8f0; border-radius:8px; }.order-card p { margin:4px 0 0; color:#64748b; font-size:11px; }.order-meta { display:flex; align-items:center; gap:12px; }.status { padding:4px 7px; border-radius:5px; background:#f1f5f9; color:#475569; font-size:10px; font-weight:700; text-transform:capitalize; }.status.ready { background:#dcfce7; color:#166534; }.status.awaiting_supplier { background:#fef3c7; color:#92400e; }.status.collected { background:#e0f2fe; color:#0369a1; }
</style>
