<template>
  <section class="manager">
    <header class="manager-head"><div><h3><PosIcon name="refresh-cw" size="20" /> Publisher Returns</h3><p>Build return batches from slow-moving and damaged stock before supplier cut-offs.</p></div><button class="primary" @click="createBatch">Create return batch</button></header>
    <div class="return-grid"><article v-for="item in items" :key="item.id" class="return-card"><label class="select"><input v-model="item.selected" type="checkbox" /> Select</label><h4>{{ item.title }}</h4><p>{{ item.publisher }} · {{ item.reason }}</p><div class="line"><span>Units available</span><strong>{{ item.quantity }}</strong></div><div class="line"><span>Credit value</span><strong>KES {{ item.value.toLocaleString() }}</strong></div></article></div>
    <footer class="return-footer"><span>{{ selectedCount }} titles selected · KES {{ selectedValue.toLocaleString() }} expected credit</span><button class="secondary" :disabled="!selectedCount" @click="submitBatch">Submit batch</button></footer>
  </section>
</template>

<script setup>
import { computed, ref } from 'vue'
import { PosIcon } from '@pos/ui'
const emit = defineEmits(['notification'])
const items = ref([
  { id: 1, title: 'Introduction to Economics', publisher: 'Longhorn', reason: 'Overstock', quantity: 12, value: 8400, selected: true },
  { id: 2, title: 'Atlas of East Africa', publisher: 'Macmillan', reason: 'Damaged cover', quantity: 3, value: 2100, selected: false },
  { id: 3, title: 'KCSE Biology Revision', publisher: 'Oxford', reason: 'Edition superseded', quantity: 8, value: 5200, selected: false },
])
const selectedCount = computed(() => items.value.filter(i => i.selected).length)
const selectedValue = computed(() => items.value.filter(i => i.selected).reduce((sum, i) => sum + i.value, 0))
function createBatch() { emit('notification', { message: 'Return batch workspace opened', type: 'info' }) }
function submitBatch() { emit('notification', { message: `${selectedCount.value} titles submitted for publisher return`, type: 'success' }) }
</script>

<style scoped>
.manager { display:flex; flex-direction:column; gap:16px; }.manager-head { display:flex; justify-content:space-between; gap:12px; }.manager-head h3 { display:flex; align-items:center; gap:8px; margin:0; font-size:17px; color:#1e293b; }.manager-head p { margin:5px 0 0; color:#64748b; font-size:12px; }.primary,.secondary { border:0; border-radius:6px; padding:8px 11px; font-size:11px; font-weight:700; cursor:pointer; }.primary { background:#7c3aed; color:#fff; }.secondary { background:#ede9fe; color:#6d28d9; }.secondary:disabled { opacity:.5; cursor:not-allowed; }.return-grid { display:grid; grid-template-columns:repeat(auto-fit,minmax(220px,1fr)); gap:10px; }.return-card { padding:14px; border:1px solid #e2e8f0; border-radius:8px; }.select { display:flex; gap:6px; color:#64748b; font-size:11px; }.return-card h4 { margin:12px 0 4px; font-size:14px; color:#1e293b; }.return-card p { margin:0 0 12px; color:#64748b; font-size:11px; }.line { display:flex; justify-content:space-between; padding-top:7px; color:#64748b; font-size:11px; }.line strong { color:#1e293b; }.return-footer { display:flex; justify-content:space-between; align-items:center; padding-top:12px; border-top:1px solid #e2e8f0; color:#64748b; font-size:12px; }
</style>
