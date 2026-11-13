<template>
  <section class="controlled-register">
    <header><div><h3><PosIcon name="lock" size="20" /> Controlled Substances Register</h3><p>Controlled-drug records and pharmacist authorization are loaded from the pharmacy service.</p></div><button type="button" class="btn-secondary" @click="exportInspectionReport"><PosIcon name="download" size="14" /> Export report</button></header>
    <div class="regulatory-notice"><PosIcon name="shield-check" size="22" /><span>Dispensing requires a valid prescription, patient identification, and pharmacist authorization.</span></div>
    <p v-if="loading" class="state">Loading controlled-drug records…</p><p v-else-if="error" class="state error">{{ error }}</p>
    <template v-else>
      <form class="auth-card" @submit.prevent="authorize"><h4>Authorize Dispensation</h4><div class="form-grid"><label>Controlled drug<select v-model="form.drug_id" required><option value="" disabled>Select a medication</option><option v-for="drug in drugs" :key="drug.id" :value="drug.id">{{ drug.name }} {{ drug.strength }}</option></select></label><label>Batch<select v-model="form.batch_no" required><option value="" disabled>Select a batch</option><option v-for="batch in batches" :key="batch.id" :value="batch.batch_no">{{ batch.batch_no }} · {{ batch.exp_date }}</option></select></label><label>Quantity<input v-model.number="form.quantity" type="number" min="1" required /></label><label>Patient name<input v-model="form.patient_name" required /></label><label>Patient ID / Passport<input v-model="form.patient_id" required /></label><label>Prescriber and license<input v-model="form.prescriber" required /></label></div><div class="auth-row"><label>Pharmacist PIN<input v-model="form.pin" type="password" maxlength="6" required /></label><button class="btn-primary" type="submit"><PosIcon name="check" size="14" /> Sign and authorize</button></div></form>
      <div class="table-section"><h4>Controlled substance disposal records <span>{{ entries.length }}</span></h4><div v-if="!entries.length" class="state">No controlled-drug records are available.</div><table v-else><thead><tr><th>Timestamp</th><th>Medication</th><th>Batch</th><th>Qty</th><th>Patient</th><th>Prescriber</th><th>Audit ref</th></tr></thead><tbody><tr v-for="entry in entries" :key="entry.id"><td>{{ entry.timestamp }}</td><td>{{ entry.drug_name }}</td><td>{{ entry.batch_no }}</td><td>{{ entry.quantity }}</td><td>{{ entry.patient_name }}<small>{{ entry.patient_id }}</small></td><td>{{ entry.prescriber }}</td><td>{{ entry.audit_ref }}</td></tr></tbody></table></div>
    </template>
  </section>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { PosIcon } from '@pos/ui'
import { useApi } from '../../../../apps/pos/src/composables/useApi.js'
const emit = defineEmits(['notification'])
const api = useApi()
const drugs = ref([])
const batches = ref([])
const entries = ref([])
const loading = ref(true)
const error = ref('')
const form = ref({ drug_id: '', batch_no: '', quantity: 1, patient_name: '', patient_id: '', prescriber: '', pin: '' })
onMounted(async () => { try { const [catalog, batchData, register] = await Promise.all([api.get('/api/v1/pharmacy/catalog'), api.get('/api/v1/pharmacy/batches'), api.get('/api/v1/pharmacy/controlled')]); drugs.value = (catalog.items ?? []).filter(item => item.kind === 'stocked'); batches.value = batchData.items ?? []; entries.value = register.items ?? [] } catch { error.value = 'Controlled-drug records could not be loaded. Please contact your administrator.' } finally { loading.value = false } })
function authorize() { emit('notification', { message: 'Controlled-drug authorization is managed by the pharmacy service. Please contact your administrator if it cannot be completed.', type: 'warning' }) }
function exportInspectionReport() { emit('notification', { message: 'The regulatory report is being prepared by the pharmacy service.', type: 'info' }) }
</script>

<style scoped>
.controlled-register{display:flex;flex-direction:column;gap:16px}header,.auth-row,.regulatory-notice{display:flex;align-items:center;gap:10px}header{justify-content:space-between}h3{display:flex;align-items:center;gap:8px;margin:0;color:#1e293b}header p{margin:4px 0 0;color:#64748b;font-size:12px}.btn-primary,.btn-secondary{display:inline-flex;align-items:center;gap:6px;padding:8px 11px;border:0;border-radius:6px;font-size:11px;font-weight:700;cursor:pointer}.btn-primary{background:#0284c7;color:#fff}.btn-secondary{border:1px solid #cbd5e1;background:#fff;color:#334155}.regulatory-notice{padding:12px;border:1px solid #bae6fd;border-radius:8px;background:#f0f9ff;color:#075985;font-size:12px}.state{padding:24px;border:1px dashed #cbd5e1;border-radius:8px;color:#64748b;text-align:center;font-size:12px}.state.error{border-color:#fecaca;color:#b91c1c}.auth-card,.table-section{padding:16px;border:1px solid #e2e8f0;border-radius:9px}.auth-card h4,.table-section h4{margin:0 0 14px;color:#1e293b}.form-grid{display:grid;grid-template-columns:repeat(3,1fr);gap:12px}.form-grid label,.auth-row label{display:flex;flex-direction:column;gap:5px;color:#475569;font-size:11px;font-weight:700}.form-grid input,.form-grid select,.auth-row input{padding:8px;border:1px solid #cbd5e1;border-radius:6px}.auth-row{justify-content:space-between;margin-top:14px}.table-section h4 span{margin-left:8px;color:#64748b;font-size:11px}table{width:100%;border-collapse:collapse;font-size:11px}th,td{padding:9px;text-align:left;border-bottom:1px solid #f1f5f9}th{color:#64748b}td small{display:block;color:#94a3b8}@media(max-width:700px){.form-grid{grid-template-columns:1fr}table{display:block;overflow:auto}}
</style>
