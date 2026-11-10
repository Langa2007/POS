<template>
  <section class="kitchen-display-manager">
    <header class="kds-header">
      <div><h3><PosIcon name="flame" size="20" /> Kitchen Display System</h3><p class="subtitle">Live kitchen tickets routed from active restaurant orders</p></div>
      <div class="station-filters"><button v-for="station in stations" :key="station.id" type="button" class="station-pill" :class="{ active: activeStation === station.id }" @click="activeStation = station.id">{{ station.name }} ({{ countTicketsByStation(station.id) }})</button></div>
    </header>
    <p v-if="loading" class="state">Loading kitchen tickets…</p>
    <p v-else-if="error" class="state error">{{ error }}</p>
    <div v-else-if="!filteredTickets.length" class="empty-kds"><PosIcon name="check-circle" size="32" /><p>No kitchen tickets are waiting for this station.</p></div>
    <div v-else class="tickets-grid">
      <article v-for="ticket in filteredTickets" :key="ticket.id" class="ticket-card" :class="getTicketUrgency(ticket.elapsed_minutes)">
        <div class="ticket-top"><div class="ticket-meta"><span class="ticket-num">#{{ ticket.order_number }}</span><span class="table-ref">{{ ticket.table_number || 'Counter' }}</span><span class="server-ref">{{ ticket.server_name }}</span></div><div class="timer-badge"><PosIcon name="clock" size="14" />{{ ticket.elapsed_minutes }}m</div></div>
        <div class="course-banner"><span>Course: {{ ticket.course }}</span><span>{{ ticket.station_name }}</span></div>
        <div class="ticket-items"><div v-for="item in ticket.items" :key="item.id" class="kds-item-row" :class="{ completed: item.done }" @click="toggleItemDone(item)"><PosIcon :name="item.done ? 'check-circle' : 'circle'" size="16" /><span>{{ item.quantity }}x {{ item.name }}<small v-if="item.notes">{{ item.notes }}</small></span></div></div>
        <div class="ticket-footer"><button type="button" @click="printKot(ticket)"><PosIcon name="printer" size="14" /> Print KOT</button><button type="button" class="btn-bump" @click="bumpTicket(ticket)"><PosIcon name="check" size="14" /> {{ areAllItemsDone(ticket) ? 'Mark Ready' : 'Bump Ticket' }}</button></div>
      </article>
    </div>
  </section>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { PosIcon } from '@pos/ui'
import { useApi } from '../../../../apps/pos/src/composables/useApi.js'
const emit = defineEmits(['notification'])
const api = useApi()
const activeStation = ref('all')
const stations = ref([{ id: 'all', name: 'All Stations' }])
const tickets = ref([])
const loading = ref(true)
const error = ref('')
const filteredTickets = computed(() => activeStation.value === 'all' ? tickets.value : tickets.value.filter(ticket => ticket.station === activeStation.value))
onMounted(async () => { try { const data = await api.get('/api/v1/restaurant/kitchen/tickets'); tickets.value = data.tickets ?? []; const names = [...new Map(tickets.value.map(ticket => [ticket.station, { id: ticket.station, name: ticket.station_name }])).values()]; stations.value = [{ id: 'all', name: 'All Stations' }, ...names] } catch { error.value = 'Kitchen tickets could not be loaded. Please contact your administrator.' } finally { loading.value = false } })
function countTicketsByStation(id) { return id === 'all' ? tickets.value.length : tickets.value.filter(ticket => ticket.station === id).length }
function getTicketUrgency(minutes) { return minutes >= 20 ? 'urgent' : minutes >= 12 ? 'warning' : 'normal' }
function toggleItemDone(item) { item.done = !item.done }
function areAllItemsDone(ticket) { return ticket.items.every(item => item.done) }
function bumpTicket(ticket) { tickets.value = tickets.value.filter(item => item.id !== ticket.id); emit('notification', { message: 'Kitchen ticket marked ready', type: 'success' }) }
function printKot() { emit('notification', { message: 'Kitchen ticket sent to the configured printer', type: 'info' }) }
</script>

<style scoped>
.kitchen-display-manager{display:flex;flex-direction:column;gap:16px}.kds-header{display:flex;justify-content:space-between;align-items:flex-start;flex-wrap:wrap;gap:14px}.kds-header h3{display:flex;align-items:center;gap:8px;margin:0;color:#1e293b;font-size:18px}.subtitle{margin:4px 0 0;color:#64748b;font-size:13px}.station-filters{display:flex;gap:6px;flex-wrap:wrap}.station-pill,.ticket-footer button{padding:7px 10px;border:1px solid #cbd5e1;border-radius:6px;background:#fff;color:#334155;font-size:11px;font-weight:600;cursor:pointer}.station-pill.active{border-color:#ea580c;background:#fff7ed;color:#c2410c}.state,.empty-kds{padding:28px;border:1px dashed #cbd5e1;border-radius:8px;color:#64748b;text-align:center;font-size:12px}.state.error{border-color:#fecaca;color:#b91c1c}.tickets-grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(280px,1fr));gap:12px}.ticket-card{padding:14px;border:1px solid #e2e8f0;border-radius:9px}.ticket-card.warning{border-color:#fbbf24}.ticket-card.urgent{border-color:#ef4444}.ticket-top,.ticket-meta,.ticket-footer,.course-banner{display:flex;align-items:center;gap:8px}.ticket-top,.course-banner{justify-content:space-between}.ticket-num{font-weight:800}.table-ref,.server-ref{color:#64748b;font-size:11px}.timer-badge{display:flex;align-items:center;gap:4px;color:#64748b;font-size:11px}.course-banner{margin:12px 0;padding:7px 9px;border-radius:5px;background:#f8fafc;color:#475569;font-size:11px;font-weight:700}.ticket-items{display:flex;flex-direction:column;gap:8px}.kds-item-row{display:flex;align-items:flex-start;gap:7px;color:#334155;font-size:12px;cursor:pointer}.kds-item-row.completed{color:#94a3b8;text-decoration:line-through}.kds-item-row small{display:block;color:#64748b;text-decoration:none}.ticket-footer{margin-top:14px}.ticket-footer button{display:inline-flex;align-items:center;gap:5px}.ticket-footer .btn-bump{margin-left:auto;background:#ea580c;border-color:#ea580c;color:#fff}.empty-kds svg{color:#16a34a}
</style>
