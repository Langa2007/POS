<template>
  <section class="manager"><header><div><h3>Fuel receipt &amp; stock movement</h3><p>Log supplier deliveries and compare receipts with tank stock movement.</p></div><button class="primary" @click="open = !open">{{ open ? 'Close form' : 'Record delivery' }}</button></header>
    <form v-if="open" class="form" @submit.prevent="submit"><label>Supplier<input v-model.trim="form.supplier" required /></label><label>Delivery note<input v-model.trim="form.note" required /></label><label>Fuel grade<select v-model="form.product"><option>Petrol</option><option>Diesel</option><option>Kerosene</option></select></label><label>Tank<input v-model.trim="form.tank" required /></label><label>Volume received (L)<input v-model.number="form.volume" type="number" min="0.001" step="0.001" required /></label><label>Received by<input v-model.trim="form.operator" required /></label><button class="primary">Save receipt</button></form>
    <div class="table-wrap"><table><thead><tr><th>Received</th><th>Supplier</th><th>Delivery note</th><th>Product</th><th>Tank</th><th>Volume</th><th>Received by</th></tr></thead><tbody><tr v-for="item in deliveries" :key="item.id"><td>{{ item.time }}</td><td>{{ item.supplier }}</td><td>{{ item.note }}</td><td>{{ item.product }}</td><td>{{ item.tank }}</td><td>{{ Number(item.volume).toLocaleString() }} L</td><td>{{ item.operator }}</td></tr><tr v-if="!deliveries.length"><td colspan="7" class="empty">Supplier receipts will appear here.</td></tr></tbody></table></div>
  </section>
</template>
<script setup>
import { reactive, ref } from 'vue'
defineProps({ deliveries: { type: Array, default: () => [] } })
const emit = defineEmits(['record'])
const open = ref(false)
const form = reactive({ supplier: '', note: '', product: 'Petrol', tank: '', volume: 0, operator: '' })
function submit() { emit('record', { ...form, id: crypto.randomUUID(), time: new Date().toLocaleString() }); Object.assign(form, { supplier: '', note: '', product: 'Petrol', tank: '', volume: 0, operator: '' }); open.value = false }
</script>
<style scoped>
.manager{display:grid;gap:16px}header{display:flex;justify-content:space-between;align-items:center}h3{margin:0;font-size:18px}header p{margin:5px 0 0;color:#6c7871;font-size:13px}.primary{padding:9px 13px;border:0;border-radius:7px;background:#16734e;color:#fff;font-weight:700;cursor:pointer}.form{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:12px;padding:16px;background:#f5f8f5;border-radius:10px}.form label{display:grid;gap:6px;color:#56635b;font-size:11px;font-weight:700}.form input,.form select{padding:9px;border:1px solid #d4ddd6;border-radius:6px;background:white}.table-wrap{overflow:auto}table{width:100%;border-collapse:collapse;white-space:nowrap;font-size:12px}th,td{padding:11px 9px;border-bottom:1px solid #edf0ed;text-align:left}th{color:#748078;font-size:10px;text-transform:uppercase}.empty{text-align:center;color:#87928b;padding:24px}@media(max-width:650px){.form{grid-template-columns:repeat(2,minmax(0,1fr))}}@media(max-width:450px){.form{grid-template-columns:1fr}}
</style>
