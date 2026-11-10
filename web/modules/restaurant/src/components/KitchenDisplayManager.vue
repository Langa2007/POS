<template>
  <div class="kitchen-display-manager">
    <div class="kds-header">
      <div>
        <h3>
          <PosIcon name="flame" size="20" class="header-icon" />
          Kitchen Display System (KDS &amp; KOT)
        </h3>
        <p class="subtitle">Live station routing, course firing, and ticket prep timers</p>
      </div>

      <div class="station-filters">
        <button
          v-for="stn in stations"
          :key="stn.id"
          type="button"
          class="station-pill"
          :class="{ 'active': activeStation === stn.id }"
          @click="activeStation = stn.id"
        >
          {{ stn.name }} ({{ countTicketsByStation(stn.id) }})
        </button>
      </div>
    </div>

    <!-- Active Tickets Queue -->
    <div class="tickets-grid">
      <div
        v-for="ticket in filteredTickets"
        :key="ticket.id"
        class="ticket-card"
        :class="getTicketUrgency(ticket.elapsedMinutes)"
      >
        <div class="ticket-top">
          <div class="ticket-meta">
            <span class="ticket-num">#{{ ticket.orderNumber }}</span>
            <span class="table-ref">T-{{ ticket.tableNumber }}</span>
            <span class="server-ref">{{ ticket.server }}</span>
          </div>
          <div class="timer-badge">
            <PosIcon name="clock" size="14" />
            <span>{{ ticket.elapsedMinutes }}m</span>
          </div>
        </div>

        <div class="course-banner" :class="ticket.course.toLowerCase()">
          <span>Course: {{ ticket.course }}</span>
          <span class="station-label">{{ ticket.stationName }}</span>
        </div>

        <div class="ticket-items">
          <div
            v-for="(it, idx) in ticket.items"
            :key="idx"
            class="kds-item-row"
            :class="{ 'completed': it.done }"
            @click="toggleItemDone(it)"
          >
            <div class="check-box">
              <PosIcon :name="it.done ? 'check-circle' : 'circle'" size="16" />
            </div>
            <div class="item-text">
              <span class="item-qty">{{ it.quantity }}x</span>
              <span class="item-title">{{ it.name }}</span>
              <small v-if="it.notes" class="item-notes">Note: {{ it.notes }}</small>
            </div>
          </div>
        </div>

        <div class="ticket-footer">
          <button type="button" class="btn-print-kot" @click="printKot(ticket)">
            <PosIcon name="printer" size="14" />
            Print KOT
          </button>
          <button
            type="button"
            class="btn-bump"
            :class="{ 'ready': areAllItemsDone(ticket) }"
            @click="bumpTicket(ticket)"
          >
            <PosIcon name="check" size="14" />
            {{ areAllItemsDone(ticket) ? 'Mark Ready' : 'Bump Ticket' }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="!filteredTickets.length" class="empty-kds">
      <PosIcon name="check-circle" size="32" class="empty-icon" />
      <p>All kitchen tickets cleared for this station. Ready for incoming orders.</p>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { PosIcon } from '@pos/ui'

const emit = defineEmits(['notification'])

const activeStation = ref('all')

const stations = [
  { id: 'all', name: 'All Stations' },
  { id: 'hot', name: 'Hot Line & Grill' },
  { id: 'cold', name: 'Cold & Salads' },
  { id: 'bar', name: 'Bar & Drinks' },
]

const tickets = ref([
  {
    id: 'kds-101',
    orderNumber: 'KOT-401',
    tableNumber: '1',
    server: 'Mary W.',
    station: 'hot',
    stationName: 'Hot Line & Grill',
    course: 'Mains',
    elapsedMinutes: 18,
    items: [
      { name: 'Prime T-Bone Steak 400g', quantity: 2, notes: '1x Medium Rare, 1x Well Done', done: false },
      { name: 'Chips Masala', quantity: 2, notes: 'Extra spicy', done: true },
    ],
  },
  {
    id: 'kds-102',
    orderNumber: 'KOT-402',
    tableNumber: '5',
    server: 'Mary W.',
    station: 'hot',
    stationName: 'Hot Line & Grill',
    course: 'Mains',
    elapsedMinutes: 8,
    items: [
      { name: 'Whole Grilled Tilapia', quantity: 2, notes: 'Crispy skin, ugali well cooked', done: false },
      { name: 'Traditional Greens', quantity: 2, notes: 'No salt added', done: false },
    ],
  },
  {
    id: 'kds-103',
    orderNumber: 'KOT-403',
    tableNumber: 'VIP 1',
    server: 'Alice N.',
    station: 'cold',
    stationName: 'Cold & Salads',
    course: 'Starters',
    elapsedMinutes: 4,
    items: [
      { name: 'Gourmet Mezze Platter', quantity: 2, notes: 'Extra pita bread', done: false },
      { name: 'Classic Caesar Salad', quantity: 1, notes: 'Dressing on the side', done: true },
    ],
  },
  {
    id: 'kds-104',
    orderNumber: 'KOT-404',
    tableNumber: 'Bar 1',
    server: 'Brian M.',
    station: 'bar',
    stationName: 'Bar & Drinks',
    course: 'Drinks',
    elapsedMinutes: 2,
    items: [
      { name: 'Singleton 12Y Double', quantity: 1, notes: 'On the rocks with twist', done: false },
      { name: 'Fresh Passion Mojito', quantity: 2, notes: 'Virgin, extra mint', done: true },
    ],
  },
])

const filteredTickets = computed(() => {
  if (activeStation.value === 'all') return tickets.value
  return tickets.value.filter(t => t.station === activeStation.value)
})

function countTicketsByStation(stnId) {
  if (stnId === 'all') return tickets.value.length
  return tickets.value.filter(t => t.station === stnId).length
}

function getTicketUrgency(minutes) {
  if (minutes >= 20) return 'urgent'
  if (minutes >= 12) return 'warning'
  return 'normal'
}

function toggleItemDone(item) {
  item.done = !item.done
}

function areAllItemsDone(ticket) {
  return ticket.items.every(it => it.done)
}

function bumpTicket(ticket) {
  tickets.value = tickets.value.filter(t => t.id !== ticket.id)
  emit('notification', {
    message: `${ticket.orderNumber} for Table ${ticket.tableNumber} completed and bumped from KDS`,
    type: 'success',
  })
}

function printKot(ticket) {
  emit('notification', {
    message: `KOT Slip ${ticket.orderNumber} dispatched to station thermal printer (Table ${ticket.tableNumber})`,
    type: 'info',
  })
}
</script>

<style scoped>
.kitchen-display-manager {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.kds-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  flex-wrap: wrap;
  gap: 12px;
}

.kds-header h3 {
  margin: 0;
  font-size: 18px;
  color: #1e293b;
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-icon {
  color: #dc2626;
}

.subtitle {
  margin: 4px 0 0;
  font-size: 13px;
  color: #64748b;
}

.station-filters {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.station-pill {
  padding: 6px 12px;
  border-radius: 20px;
  border: 1px solid #cbd5e1;
  background: white;
  font-size: 12px;
  font-weight: 600;
  color: #475569;
  cursor: pointer;
  transition: all 0.15s ease;
}

.station-pill:hover {
  background: #f1f5f9;
}

.station-pill.active {
  background: #dc2626;
  color: white;
  border-color: #dc2626;
}

.tickets-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 16px;
}

.ticket-card {
  background: white;
  border: 2px solid #e2e8f0;
  border-radius: 10px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.04);
}

.ticket-card.normal { border-color: #94a3b8; }
.ticket-card.warning { border-color: #f59e0b; }
.ticket-card.urgent { border-color: #ef4444; }

.ticket-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  background: #f8fafc;
  border-bottom: 1px solid #e2e8f0;
}

.ticket-meta {
  display: flex;
  align-items: center;
  gap: 8px;
}

.ticket-num {
  font-weight: 800;
  font-size: 14px;
  color: #0f172a;
}

.table-ref {
  background: #0f172a;
  color: white;
  font-size: 11px;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 4px;
}

.server-ref {
  font-size: 11px;
  color: #64748b;
}

.timer-badge {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 4px;
}

.ticket-card.normal .timer-badge { background: #e2e8f0; color: #334155; }
.ticket-card.warning .timer-badge { background: #fef3c7; color: #b45309; }
.ticket-card.urgent .timer-badge { background: #fee2e2; color: #dc2626; }

.course-banner {
  display: flex;
  justify-content: space-between;
  padding: 4px 12px;
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  background: #f1f5f9;
  color: #475569;
}

.course-banner.mains { background: #ffedd5; color: #c2410c; }
.course-banner.starters { background: #dbeafe; color: #1d4ed8; }
.course-banner.drinks { background: #fae8ff; color: #86198f; }

.station-label {
  font-size: 10px;
  font-weight: 600;
  opacity: 0.8;
}

.ticket-items {
  padding: 10px 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  flex: 1;
}

.kds-item-row {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 6px 8px;
  border-radius: 6px;
  background: #f8fafc;
  cursor: pointer;
  transition: background 0.15s ease;
}

.kds-item-row:hover {
  background: #f1f5f9;
}

.kds-item-row.completed {
  opacity: 0.5;
  text-decoration: line-through;
}

.check-box {
  margin-top: 2px;
  color: #64748b;
}

.kds-item-row.completed .check-box {
  color: #16a34a;
}

.item-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.item-qty {
  font-weight: 800;
  color: #0f172a;
  margin-right: 4px;
}

.item-title {
  font-size: 13px;
  font-weight: 600;
  color: #1e293b;
}

.item-notes {
  font-size: 11px;
  color: #ea580c;
  font-weight: 500;
}

.ticket-footer {
  display: flex;
  padding: 10px 12px;
  gap: 8px;
  background: #f8fafc;
  border-top: 1px solid #e2e8f0;
}

.btn-print-kot, .btn-bump {
  padding: 8px 10px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  flex: 1;
  transition: all 0.15s ease;
}

.btn-print-kot {
  background: white;
  border: 1px solid #cbd5e1;
  color: #475569;
}

.btn-print-kot:hover { background: #f1f5f9; }

.btn-bump {
  background: #0f172a;
  color: white;
  border: 1px solid #0f172a;
}

.btn-bump:hover { background: #334155; }
.btn-bump.ready { background: #16a34a; border-color: #16a34a; }
.btn-bump.ready:hover { background: #15803d; }

.empty-kds {
  padding: 40px;
  text-align: center;
  background: white;
  border-radius: 12px;
  border: 1px dashed #cbd5e1;
  color: #64748b;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.empty-icon {
  color: #16a34a;
}
</style>
