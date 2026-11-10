<template>
  <aside class="salon-sidebar">
    <header><div class="title"><PosIcon name="sparkles" size="18" /><strong>Salon Service</strong></div><span class="badge">Appointments</span></header>
    <p v-if="loading" class="state">Loading salon options…</p>
    <p v-else-if="error" class="state error">{{ error }}</p>
    <template v-else>
      <label>Assigned stylist
        <select v-model="stylistId" @change="notify('Stylist assigned to this sale')"><option value="">Unassigned</option><option v-for="person in staff" :key="person.id" :value="person.id">{{ person.name }}</option></select>
      </label>
      <label>Service package
        <select v-model="packageId" @change="notify('Service package selected')"><option value="">No package</option><option v-for="pkg in packages" :key="pkg.id" :value="pkg.id">{{ pkg.name }}</option></select>
      </label>
      <label class="check"><input v-model="loyalty" type="checkbox" @change="notify(loyalty ? 'Loyalty points will be earned' : 'Loyalty points disabled')" /> Apply loyalty rewards</label>
    </template>
    <div class="actions"><button type="button" @click="notify('Appointment receipt prepared')"><PosIcon name="calendar" size="14" /> Appointment slip</button><button type="button" @click="notify('Commission preview opened')"><PosIcon name="dollar-sign" size="14" /> Commission</button></div>
  </aside>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { PosIcon } from '@pos/ui'
import { useApi } from '../../../../apps/pos/src/composables/useApi.js'
defineProps({ cart: { type: Array, default: () => [] } })
const api = useApi()
const emit = defineEmits(['notification'])
const staff = ref([])
const packages = ref([])
const stylistId = ref('')
const packageId = ref('')
const loyalty = ref(true)
const loading = ref(true)
const error = ref('')
onMounted(async () => {
  try {
    const [staffData, packageData] = await Promise.all([api.get('/api/v1/salon/staff'), api.get('/api/v1/salon/packages')])
    staff.value = staffData.staff ?? []
    packages.value = packageData.packages ?? []
  } catch (err) {
    error.value = err.message
    emit('notification', { message: `Salon options could not be loaded: ${err.message}`, type: 'error' })
  } finally { loading.value = false }
})
function notify(message) { window.dispatchEvent(new CustomEvent('pos:notification', { detail: { message, type: 'info' } })) }
</script>

<style scoped>
.salon-sidebar { display:flex; flex-direction:column; gap:14px; padding:16px; background:#fff; border:1px solid #e2e8f0; border-radius:12px; } header,.title,.actions,.check { display:flex; align-items:center; gap:8px; } header { justify-content:space-between; border-bottom:1px solid #f1f5f9; padding-bottom:8px; }.title { color:#334155; font-size:13px; }.badge { padding:3px 7px; border-radius:5px; background:#fce7f3; color:#be185d; font-size:10px; font-weight:700; } label { display:flex; flex-direction:column; gap:6px; color:#64748b; font-size:11px; font-weight:700; text-transform:uppercase; } select { padding:8px; border:1px solid #cbd5e1; border-radius:6px; background:#f8fafc; font-size:12px; }.check { flex-direction:row; color:#475569; text-transform:none; }.actions { border-top:1px solid #f1f5f9; padding-top:10px; } button { flex:1; display:inline-flex; align-items:center; justify-content:center; gap:5px; padding:8px; border:1px solid #cbd5e1; border-radius:6px; background:#fff; color:#334155; font-size:11px; font-weight:600; cursor:pointer; }.state { color:#64748b; font-size:11px; }.state.error { color:#b91c1c; }
</style>
