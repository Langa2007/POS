<template>
  <section class="manager">
    <header class="manager-head">
      <div>
        <h3><PosIcon name="calendar" size="20" /> Appointment Book</h3>
        <p>Manage today's bookings, walk-ins, and stylist availability.</p>
      </div>
      <button class="primary" @click="book">New appointment</button>
    </header>

    <p v-if="loading" class="state">Loading appointments…</p>
    <p v-else-if="error" class="state error" role="alert">{{ error }}</p>
    <p v-else-if="!appointments.length" class="state">No appointments scheduled for today.</p>
    <div v-else class="schedule">
      <article
        v-for="appointment in appointments"
        :key="appointment.id"
        class="appointment"
        :class="appointment.status"
      >
        <div class="time">{{ formatTime(appointment.starts_at) }}</div>
        <div class="client">
          <strong>{{ appointment.client_name || 'Walk-in client' }}</strong>
          <p>{{ appointment.service_type }} · {{ appointment.staff_name || appointment.staff_id || 'Unassigned' }}</p>
          <small v-if="appointment.client_phone">{{ appointment.client_phone }}</small>
        </div>
        <span class="status">{{ formatStatus(appointment.status) }}</span>
        <button
          v-if="nextStatus(appointment.status)"
          class="secondary"
          :disabled="updatingId === appointment.id"
          @click="advance(appointment)"
        >
          {{ updatingId === appointment.id ? 'Updating…' : actionLabel(appointment.status) }}
        </button>
      </article>
    </div>
  </section>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { PosIcon } from '@pos/ui'
import { useApi } from '../../../../apps/pos/src/composables/useApi.js'

const emit = defineEmits(['notification'])
const api = useApi()
const appointments = ref([])
const loading = ref(true)
const error = ref('')
const updatingId = ref('')

onMounted(loadAppointments)

async function loadAppointments() {
  loading.value = true
  error.value = ''
  try {
    const data = await api.get('/api/v1/salon/appointments')
    appointments.value = data.appointments ?? []
  } catch (err) {
    error.value = err.message
    emit('notification', { message: `Appointments could not be loaded: ${err.message}`, type: 'error' })
  } finally {
    loading.value = false
  }
}

function nextStatus(status) {
  return { booked: 'completed', no_show: 'booked', completed: null, cancelled: null }[status] ?? null
}

function actionLabel(status) {
  return status === 'booked' ? 'Complete' : 'Update'
}

async function advance(appointment) {
  const status = nextStatus(appointment.status)
  if (!status) return
  updatingId.value = appointment.id
  try {
    await api.patch(`/api/v1/salon/appointments/${appointment.id}?status=${status}`)
    appointment.status = status
    emit('notification', { message: `${appointment.client_name || 'Appointment'} marked ${formatStatus(status)}`, type: 'success' })
  } catch (err) {
    emit('notification', { message: `Appointment update failed: ${err.message}`, type: 'error' })
  } finally {
    updatingId.value = ''
  }
}

function formatTime(value) {
  if (!value) return '—'
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

function formatStatus(status) {
  return String(status || '').replace('_', ' ')
}

function book() {
  emit('notification', { message: 'New appointment creation is available through the appointment API.', type: 'info' })
}
</script>

<style scoped>
.manager { display:flex; flex-direction:column; gap:16px; }.manager-head { display:flex; justify-content:space-between; gap:12px; }.manager-head h3 { display:flex; align-items:center; gap:8px; margin:0; color:#1e293b; font-size:17px; }.manager-head p { margin:5px 0 0; color:#64748b; font-size:12px; }.primary,.secondary { border:0; border-radius:6px; padding:8px 11px; font-size:11px; font-weight:700; cursor:pointer; }.primary { background:#db2777; color:#fff; }.secondary { background:#fce7f3; color:#be185d; }.secondary:disabled { opacity:.6; cursor:wait; }.state { padding:24px; border:1px dashed #cbd5e1; border-radius:8px; color:#64748b; text-align:center; font-size:12px; }.state.error { border-color:#fecaca; background:#fff7f7; color:#b91c1c; }.schedule { display:flex; flex-direction:column; gap:8px; }.appointment { display:grid; grid-template-columns:70px 1fr auto auto; gap:14px; align-items:center; padding:12px 14px; border-left:4px solid #cbd5e1; border-radius:7px; background:#fff; box-shadow:0 1px 3px #0f172a10; }.appointment.booked { border-left-color:#db2777; }.appointment.completed { border-left-color:#16a34a; opacity:.72; }.appointment.cancelled { border-left-color:#94a3b8; opacity:.6; }.appointment.no_show { border-left-color:#dc2626; }.time { color:#334155; font-size:15px; font-weight:800; }.client strong { color:#1e293b; font-size:13px; }.client p { margin:3px 0; color:#64748b; font-size:11px; }.client small { color:#94a3b8; font-size:10px; }.status { padding:4px 7px; border-radius:5px; background:#f1f5f9; color:#475569; font-size:10px; text-transform:capitalize; }
</style>
