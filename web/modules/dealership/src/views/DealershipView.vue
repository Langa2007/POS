<template>
  <main class="dealership">
    <header><span>VEHICLE OPERATIONS</span><h1>Dealership desk</h1><p>Manage serial inventory, warranties, trade-ins, deposits and payment plans.</p></header>
    <p v-if="error" class="error">{{ error }}</p>
    <nav><button v-for="tab in tabs" :key="tab.id" :class="{ active: active === tab.id }" @click="active = tab.id">{{ tab.label }}</button></nav>
    <section>
      <div class="toolbar"><strong>{{ activeLabel }}</strong><button @click="createRecord">Add record</button></div>
      <div v-if="loading" class="muted">Loading records…</div>
      <table v-else>
        <thead><tr><th>Reference</th><th>Status</th><th>Details</th><th></th></tr></thead>
        <tbody>
          <tr v-for="record in records" :key="record.id">
            <td>{{ record.registration || record.serial || record.reference || record.id }}</td>
            <td>{{ record.status || 'Active' }}</td>
            <td>{{ record.make || record.provider || record.customer || record.type || '-' }}</td>
            <td><button class="remove" @click="remove(record.id)">Remove</button></td>
          </tr>
          <tr v-if="!records.length"><td colspan="4" class="muted">No records yet</td></tr>
        </tbody>
      </table>
    </section>
  </main>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useApi } from '../../../../apps/pos/src/composables/useApi.js'

const api = useApi()
const tabs = [
  { id: 'vehicles', label: 'Vehicles' }, { id: 'warranties', label: 'Warranties' },
  { id: 'trade_ins', label: 'Trade-ins' }, { id: 'installments', label: 'Installments' },
  { id: 'deposits', label: 'Deposits' }, { id: 'documents', label: 'Documents' },
]
const active = ref('vehicles')
const records = ref([])
const loading = ref(false)
const error = ref('')
const activeLabel = computed(() => tabs.find((tab) => tab.id === active.value)?.label)

async function load() {
  loading.value = true
  error.value = ''
  try {
    const data = await api.get(`/api/v1/dealership/${active.value}`)
    records.value = data[active.value] ?? []
  } catch (err) {
    error.value = `Could not load dealership records: ${err.message}`
  } finally {
    loading.value = false
  }
}

async function createRecord() {
  const reference = window.prompt(`New ${activeLabel.value} reference`)
  if (!reference) return
  try {
    const payload = active.value === 'vehicles'
      ? { registration: reference, serial: '', make: '', model: '' }
      : active.value === 'warranties'
        ? { vehicle: reference, provider: '', expires: '' }
        : active.value === 'trade_ins'
          ? { customer: '', vehicle: reference, allowance: 0 }
          : active.value === 'installments'
            ? { customer: '', vehicle: reference, deposit: 0, balance: 0 }
            : active.value === 'deposits'
              ? { customer: '', vehicle: reference, amount: 0 }
              : { vehicle: reference, type: 'Document', reference }
    const saved = await api.post(`/api/v1/dealership/${active.value}`, payload)
    records.value.unshift(saved)
  } catch (err) {
    error.value = `Could not save record: ${err.message}`
  }
}

async function remove(id) {
  try {
    await api.delete(`/api/v1/dealership/${active.value}/${id}`)
    records.value = records.value.filter((record) => record.id !== id)
  } catch (err) {
    error.value = `Could not remove record: ${err.message}`
  }
}

watch(active, load)
onMounted(load)
</script>

<style scoped>
.dealership{display:grid;gap:16px}.dealership header{padding:24px;border-radius:12px;background:#172554;color:white}.dealership header span{font-size:10px;font-weight:800;letter-spacing:1.5px;color:#93c5fd}.dealership h1{margin:7px 0}.dealership header p{margin:0;color:#cbd5e1}.dealership nav{display:flex;gap:4px;overflow:auto}.dealership nav button{padding:9px 12px;border:0;border-bottom:2px solid transparent;background:none;color:#64748b;font-weight:700;cursor:pointer}.dealership nav button.active{border-color:#2563eb;color:#1d4ed8}.dealership section{border:1px solid #e2e8f0;border-radius:10px;background:white;overflow:hidden}.toolbar{display:flex;justify-content:space-between;padding:14px;border-bottom:1px solid #e2e8f0}.toolbar button,.dealership td button{border:0;border-radius:7px;padding:7px 10px;background:#2563eb;color:white;cursor:pointer}.dealership table{width:100%;border-collapse:collapse}.dealership th,.dealership td{padding:11px 14px;border-bottom:1px solid #edf2f7;text-align:left;font-size:12px}.dealership th{color:#64748b}.dealership .remove{background:#fee2e2;color:#991b1b}.muted{color:#64748b}.error{padding:10px;border-radius:7px;background:#fff1f2;color:#9f1239}
</style>
