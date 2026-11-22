<template>
  <main class="membership">
    <header><span>MEMBER OPERATIONS</span><h1>Membership desk</h1><p>Manage plans, subscriptions, attendance and recurring billing.</p></header>
    <p v-if="error" class="error">{{ error }}</p>
    <nav><button v-for="tab in tabs" :key="tab.id" :class="{ active: active === tab.id }" @click="active = tab.id">{{ tab.label }}</button></nav>
    <section>
      <div class="toolbar"><strong>{{ activeLabel }}</strong><button @click="createRecord">Add record</button></div>
      <div v-if="loading" class="muted">Loading records…</div>
      <table v-else>
        <thead><tr><th>Reference</th><th>Status</th><th>Details</th><th></th></tr></thead>
        <tbody>
          <tr v-for="record in records" :key="record.id">
            <td>{{ record.name || record.member || record.reference || record.id }}</td>
            <td>{{ record.status || 'Active' }}</td>
            <td>{{ record.plan || record.period || record.location || record.reason || '-' }}</td>
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
  { id: 'plans', label: 'Membership plans' }, { id: 'members', label: 'Members' },
  { id: 'attendance', label: 'Attendance' }, { id: 'billing', label: 'Billing' },
  { id: 'freezes', label: 'Freezes' },
]
const active = ref('members')
const records = ref([])
const loading = ref(false)
const error = ref('')
const activeLabel = computed(() => tabs.find((tab) => tab.id === active.value)?.label)

async function load() {
  loading.value = true
  error.value = ''
  try {
    const data = await api.get(`/api/v1/membership/${active.value}`)
    records.value = data[active.value] ?? []
  } catch (err) {
    error.value = `Could not load membership records: ${err.message}`
  } finally {
    loading.value = false
  }
}

async function createRecord() {
  const reference = window.prompt(`New ${activeLabel.value} reference`)
  if (!reference) return
  try {
    const payload = active.value === 'members'
      ? { name: reference, plan: 'Standard', status: 'Active' }
      : active.value === 'plans'
        ? { name: reference, price: 0, period: 'monthly' }
        : active.value === 'billing'
          ? { member: reference, amount: 0, period: 'monthly', status: 'Pending' }
          : active.value === 'attendance'
            ? { member: reference, entered: new Date().toISOString(), location: 'Main' }
            : { member: reference, start: new Date().toISOString().slice(0, 10), end: new Date().toISOString().slice(0, 10), reason: 'Requested' }
    const saved = await api.post(`/api/v1/membership/${active.value}`, payload)
    records.value.unshift(saved)
  } catch (err) {
    error.value = `Could not save record: ${err.message}`
  }
}

async function remove(id) {
  try {
    await api.delete(`/api/v1/membership/${active.value}/${id}`)
    records.value = records.value.filter((record) => record.id !== id)
  } catch (err) {
    error.value = `Could not remove record: ${err.message}`
  }
}

watch(active, load)
onMounted(load)
</script>

<style scoped>
.membership{display:grid;gap:16px}.membership header{padding:24px;border-radius:12px;background:#4c1d95;color:white}.membership header span{font-size:10px;font-weight:800;letter-spacing:1.5px;color:#c4b5fd}.membership h1{margin:7px 0}.membership header p{margin:0;color:#ddd6fe}.membership nav{display:flex;gap:4px;overflow:auto}.membership nav button{padding:9px 12px;border:0;border-bottom:2px solid transparent;background:none;color:#64748b;font-weight:700;cursor:pointer}.membership nav button.active{border-color:#7c3aed;color:#6d28d9}.membership section{border:1px solid #e2e8f0;border-radius:10px;background:white;overflow:hidden}.toolbar{display:flex;justify-content:space-between;padding:14px;border-bottom:1px solid #e2e8f0}.toolbar button,.membership td button{border:0;border-radius:7px;padding:7px 10px;background:#7c3aed;color:white;cursor:pointer}.membership table{width:100%;border-collapse:collapse}.membership th,.membership td{padding:11px 14px;border-bottom:1px solid #edf2f7;text-align:left;font-size:12px}.membership th{color:#64748b}.membership .remove{background:#fee2e2;color:#991b1b}.muted{color:#64748b}.error{padding:10px;border-radius:7px;background:#fff1f2;color:#9f1239}
</style>
