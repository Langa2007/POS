<template>
  <section class="manager"><header><div><h3>Tank inventory &amp; dip log</h3><p>Record calibrated manual dips to track estimated on-hand fuel by tank.</p></div></header>
    <div class="summary"><article v-for="tank in tankSummary" :key="tank.key"><span>{{ tank.product }} · {{ tank.name }}</span><strong>{{ fmt(tank.volume) }} L</strong><small>{{ tank.time || 'No dip recorded' }}</small></article><article class="add"><strong>New reading</strong><small>Record the dip and measured volume below.</small></article></div>
    <form class="form" @submit.prevent="submit"><label>Tank identifier<input v-model.trim="form.tank" required placeholder="Tank 1" /></label><label>Fuel grade<select v-model="form.product"><option>Petrol</option><option>Diesel</option><option>Kerosene</option></select></label><label>Dip (mm)<input v-model.number="form.reading" type="number" min="0" step="0.1" required /></label><label>Calibrated volume (L)<input v-model.number="form.volume" type="number" min="0" step="0.1" required /></label><label>Recorded by<input v-model.trim="form.operator" required /></label><button class="primary">Save dip reading</button></form>
    <div class="table-wrap"><table><thead><tr><th>Time</th><th>Tank</th><th>Product</th><th>Dip</th><th>Volume</th><th>Operator</th></tr></thead><tbody><tr v-for="dip in dips" :key="dip.id"><td>{{ dip.time }}</td><td>{{ dip.tank }}</td><td>{{ dip.product }}</td><td>{{ fmt(dip.reading) }} mm</td><td>{{ fmt(dip.volume) }} L</td><td>{{ dip.operator }}</td></tr><tr v-if="!dips.length"><td colspan="6" class="empty">No tank dips have been recorded.</td></tr></tbody></table></div>
  </section>
</template>
<script setup>
import { computed, reactive } from 'vue'
const props = defineProps({ dips: { type: Array, default: () => [] } })
const emit = defineEmits(['record'])
const form = reactive({ tank: '', product: 'Petrol', reading: 0, volume: 0, operator: '' })
const tankSummary = computed(() => { const latest = new Map(); for (const dip of props.dips) if (!latest.has(dip.tank)) latest.set(dip.tank, dip); return [...latest.values()] })
function fmt(value) { return Number(value || 0).toLocaleString(undefined, { maximumFractionDigits: 1 }) }
function submit() { emit('record', { ...form, id: crypto.randomUUID(), time: new Date().toLocaleString() }); Object.assign(form, { tank: '', product: 'Petrol', reading: 0, volume: 0, operator: '' }) }
</script>
<style scoped>
.manager{display:grid;gap:16px}header h3{margin:0;font-size:18px}header p{margin:5px 0 0;color:#6c7871;font-size:13px}.summary{display:grid;grid-template-columns:repeat(auto-fit,minmax(160px,1fr));gap:10px}.summary article{display:grid;gap:5px;padding:13px;border:1px solid #e2e8e0;border-radius:9px}.summary span,.summary small{color:#6b786f;font-size:11px}.summary strong{font-size:18px}.summary .add{background:#f5f8f5}.form{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:12px;padding:16px;background:#f5f8f5;border-radius:10px}.form label{display:grid;gap:6px;color:#56635b;font-size:11px;font-weight:700}.form input,.form select{padding:9px;border:1px solid #d4ddd6;border-radius:6px;background:white}.primary{align-self:end;padding:10px 12px;border:0;border-radius:7px;background:#16734e;color:#fff;font-weight:700;cursor:pointer}.table-wrap{overflow:auto}table{width:100%;border-collapse:collapse;white-space:nowrap;font-size:12px}th,td{padding:11px 9px;border-bottom:1px solid #edf0ed;text-align:left}th{color:#748078;font-size:10px;text-transform:uppercase}.empty{text-align:center;color:#87928b;padding:24px}@media(max-width:650px){.form{grid-template-columns:repeat(2,minmax(0,1fr))}}@media(max-width:450px){.form{grid-template-columns:1fr}}
</style>
