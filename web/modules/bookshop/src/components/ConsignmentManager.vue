<template>
  <section class="manager">
    <header class="manager-head"><div><h3><PosIcon name="package" size="20" /> Consignment &amp; Second-hand Stock</h3><p>Reconcile supplier-owned stock and grade used books for resale.</p></div><button class="primary" @click="intake">Record intake</button></header>
    <div class="consignment-table"><div class="table-head"><span>Supplier / source</span><span>Units</span><span>Retail value</span><span>Payable</span><span>Status</span></div><div v-for="row in rows" :key="row.id" class="table-row"><strong>{{ row.source }}</strong><span>{{ row.units }}</span><span>KES {{ row.retail.toLocaleString() }}</span><span>KES {{ row.payable.toLocaleString() }}</span><span class="status" :class="row.status">{{ row.status }}</span></div></div>
    <div class="intake-note"><PosIcon name="info" size="16" /><span>Consignment settlements are calculated after sold units are reconciled. Used-book grades are recorded at intake.</span></div>
  </section>
</template>

<script setup>
import { ref } from 'vue'
import { PosIcon } from '@pos/ui'
const emit = defineEmits(['notification'])
const rows = ref([{ id:1, source:'East African Educational', units:146, retail:128400, payable:77100, status:'open' }, { id:2, source:'Nairobi Book Exchange', units:38, retail:22400, payable:0, status:'second-hand' }, { id:3, source:'Moran Publishers', units:82, retail:65400, payable:39240, status:'reconcile' }])
function intake() { emit('notification', { message: 'Stock intake form opened for grading and supplier assignment', type: 'info' }) }
</script>

<style scoped>
.manager { display:flex; flex-direction:column; gap:16px; }.manager-head { display:flex; justify-content:space-between; gap:12px; }.manager-head h3 { display:flex; align-items:center; gap:8px; margin:0; font-size:17px; color:#1e293b; }.manager-head p { margin:5px 0 0; color:#64748b; font-size:12px; }.primary { border:0; border-radius:6px; padding:8px 11px; background:#7c3aed; color:#fff; font-size:11px; font-weight:700; cursor:pointer; }.consignment-table { overflow:auto; border:1px solid #e2e8f0; border-radius:8px; }.table-head,.table-row { display:grid; grid-template-columns:2fr .7fr 1fr 1fr 1fr; gap:10px; min-width:620px; padding:12px 14px; align-items:center; }.table-head { background:#faf5ff; color:#64748b; font-size:10px; font-weight:800; text-transform:uppercase; }.table-row { border-top:1px solid #f1f5f9; color:#475569; font-size:12px; }.table-row strong { color:#1e293b; }.status { width:max-content; padding:3px 6px; border-radius:5px; background:#f1f5f9; color:#475569; font-size:10px; }.status.reconcile { background:#fef3c7; color:#92400e; }.status.second-hand { background:#ede9fe; color:#6d28d9; }.intake-note { display:flex; gap:8px; padding:11px; border:1px solid #bae6fd; border-radius:7px; background:#f0f9ff; color:#0369a1; font-size:11px; }
</style>
