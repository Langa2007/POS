<template>
  <section class="manager">
    <header><div><h3>Meter shift closeout</h3><p>Compare dispenser movement plus deliveries against metered till sales.</p></div><button class="primary" @click="open = !open">{{ open ? 'Cancel' : 'Close a shift' }}</button></header>
    <form v-if="open" class="form" @submit.prevent="submit"><label>Pump / product<input v-model.trim="form.pump" required placeholder="Pump 1 · Diesel" /></label><label>Attendant<input v-model.trim="form.operator" required /></label><label>Opening meter (L)<input v-model.number="form.open" type="number" min="0" step="0.001" required /></label><label>Closing meter (L)<input v-model.number="form.close" type="number" min="0" step="0.001" required /></label><label>Received during shift (L)<input v-model.number="form.deliveries" type="number" min="0" step="0.001" /></label><label>Till volume (L)<input v-model.number="form.sales" type="number" min="0" step="0.001" required /></label><label>Allowed variance (L)<input v-model.number="form.tolerance" type="number" min="0" step="0.001" /></label><button class="primary">Save closeout</button></form>
    <p v-if="message" class="message">{{ message }}</p>
    <div class="table-wrap"><table><thead><tr><th>Shift</th><th>Pump</th><th>Attendant</th><th>Meter movement</th><th>Till volume</th><th>Variance</th><th>Closeout</th></tr></thead><tbody><tr v-for="shift in shifts" :key="shift.id"><td>{{ shift.time }}</td><td>{{ shift.pump }}</td><td>{{ shift.operator }}</td><td>{{ fmt(shift.volume) }} L</td><td>{{ fmt(shift.sales) }} L</td><td>{{ fmt(shift.variance) }} L</td><td><span class="status" :class="shift.ok ? 'good' : 'review'">{{ shift.ok ? 'Balanced' : 'Investigate' }}</span></td></tr><tr v-if="!shifts.length"><td colspan="7" class="empty">Shift closeouts will appear here.</td></tr></tbody></table></div>
    <aside class="rule"><strong>Reconciliation rule</strong><span>Closing − opening + received volume = expected dispensed volume. Variance outside tolerance is flagged for investigation.</span></aside>
  </section>
</template>
<script setup>
import { reactive, ref } from 'vue'
const props = defineProps({ shifts: { type: Array, default: () => [] } })
const emit = defineEmits(['record'])
const open = ref(false)
const message = ref('')
const form = reactive({ pump: '', operator: '', open: 0, close: 0, deliveries: 0, sales: 0, tolerance: 1 })
function fmt(value) { return Number(value || 0).toLocaleString(undefined, { maximumFractionDigits: 3 }) }
function submit() {
  const volume = Number(form.close) - Number(form.open) + Number(form.deliveries)
  const variance = volume - Number(form.sales)
  const readingsValid = Number(form.close) >= Number(form.open) && Number(form.deliveries) >= 0 && Number(form.sales) >= 0
  const ok = readingsValid && Math.abs(variance) <= Number(form.tolerance)
  emit('record', { ...form, volume, variance, ok, time: new Date().toLocaleString() })
  message.value = !readingsValid ? 'Closeout saved for review: closing meter cannot be lower than opening.' : ok ? 'Closeout balanced and saved.' : 'Closeout saved with a variance. Review readings before approving.'
  Object.assign(form, { pump: '', operator: '', open: 0, close: 0, deliveries: 0, sales: 0 })
  open.value = false
}
</script>
<style scoped>
.manager{display:grid;gap:16px}header{display:flex;justify-content:space-between;align-items:center;gap:14px}h3{margin:0;font-size:18px}header p{margin:5px 0 0;color:#6c7871;font-size:13px}.primary{padding:9px 13px;border:0;border-radius:7px;background:#16734e;color:#fff;font-weight:700;cursor:pointer}.form{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:12px;padding:16px;background:#f5f8f5;border-radius:10px}.form label{display:grid;gap:6px;color:#56635b;font-size:11px;font-weight:700}.form input{padding:9px;border:1px solid #d4ddd6;border-radius:6px}.table-wrap{overflow:auto}table{width:100%;border-collapse:collapse;white-space:nowrap;font-size:12px}th,td{padding:11px 9px;border-bottom:1px solid #edf0ed;text-align:left}th{color:#748078;font-size:10px;text-transform:uppercase}.status{padding:4px 8px;border-radius:20px;font-size:10px;font-weight:800}.good{background:#e3f5e9;color:#17623b}.review{background:#fff0e4;color:#9a4e14}.empty{text-align:center;color:#87928b;padding:24px}.message{margin:0;color:#17623b;font-size:12px}.rule{display:flex;gap:10px;padding:12px;border-radius:8px;background:#f2f7f3;color:#516056;font-size:12px}.rule strong{white-space:nowrap;color:#17623b}@media(max-width:760px){.form{grid-template-columns:repeat(2,minmax(0,1fr))}}@media(max-width:480px){.form{grid-template-columns:1fr}}
</style>
