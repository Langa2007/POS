<template>
  <section class="table-service-manager">
    <header class="manager-header">
      <div><h3><PosIcon name="layout-grid" size="20" /> Floor Plan &amp; Table Service</h3><p class="subtitle">Live table occupancy, guest covers, and active order tabs</p></div>
      <div class="section-filter"><button v-for="section in sections" :key="section.id" type="button" class="filter-pill" :class="{ active: activeSection === section.id }" @click="activeSection = section.id">{{ section.name }} ({{ countTablesBySection(section.id) }})</button></div>
    </header>

    <div v-if="loading" class="state">Loading table service…</div>
    <div v-else-if="error" class="state error">{{ error }}</div>
    <template v-else>
      <div class="occupancy-stats">
        <div class="stat-box"><span>Total Tables</span><strong>{{ tables.length }}</strong></div>
        <div class="stat-box occupied"><span>Occupied / Dining</span><strong>{{ countStatus('occupied') }}</strong></div>
        <div class="stat-box billing"><span>Awaiting Bill</span><strong>{{ countStatus('billing') }}</strong></div>
        <div class="stat-box available"><span>Available</span><strong>{{ countStatus('available') }}</strong></div>
        <div class="stat-box covers"><span>Active Covers</span><strong>{{ totalGuests }} Guests</strong></div>
      </div>
      <div v-if="!filteredTables.length" class="state">No tables are configured for this area. Please contact your administrator.</div>
      <div v-else class="tables-grid">
        <button v-for="table in filteredTables" :key="table.id" type="button" class="table-card" :class="[table.status, { selected: selectedTable?.id === table.id }]" @click="selectTable(table)">
          <div class="table-top"><span>Table {{ table.number }}</span><span class="status-badge" :class="table.status">{{ formatStatus(table.status) }}</span></div>
          <div class="table-details"><span><PosIcon name="users" size="14" /> {{ table.guests }}/{{ table.capacity }} seats</span><span v-if="table.server"><PosIcon name="user" size="14" /> {{ table.server }}</span><span v-if="table.seated_at"><PosIcon name="clock" size="14" /> {{ table.seated_at }}</span></div>
          <div v-if="table.status !== 'available'" class="table-amount"><small>Current Tab</small><strong>KES {{ formatMoney(table.total_minor) }}</strong></div><div v-else class="table-empty-hint">Tap to seat guests</div>
        </button>
      </div>
    </template>

    <div v-if="selectedTable" class="table-detail-drawer">
      <header class="drawer-header"><h4>Table {{ selectedTable.number }}<span v-if="selectedTable.section_name"> &mdash; {{ selectedTable.section_name }}</span></h4><button type="button" class="close-btn" @click="selectedTable = null"><PosIcon name="x" size="18" /></button></header>
      <div v-if="selectedTable.status === 'available'" class="seat-form">
        <h5>Seat New Party</h5><label>Number of Guests (Max {{ selectedTable.capacity }})<input v-model.number="seatForm.guests" type="number" min="1" :max="selectedTable.capacity" /></label>
        <label v-if="servers.length">Assigned Server<select v-model="seatForm.server"><option v-for="server in servers" :key="server.id" :value="server.name">{{ server.name }}</option></select></label>
        <label>Guest / Booking Name (Optional)<input v-model="seatForm.guest_name" type="text" /></label>
        <button type="button" class="btn-primary" @click="seatTableAction"><PosIcon name="check" size="16" /> Open Table &amp; Start Tab</button>
      </div>
      <div v-else class="tab-management">
        <div class="tab-info-strip"><span>Server: <strong>{{ selectedTable.server || 'Unassigned' }}</strong></span><span>Guests: <strong>{{ selectedTable.guests }}</strong></span><span>Duration: <strong>{{ selectedTable.seated_at || 'Not available' }}</strong></span></div>
        <h5>Ordered Items</h5><p v-if="!selectedTable.items?.length" class="empty-items">No items ordered yet. Add items via checkout.</p>
        <div v-for="(item, index) in selectedTable.items || []" :key="item.id || index" class="tab-item-row"><span>{{ item.quantity }}x {{ item.name }} <small v-if="item.modifiers">({{ item.modifiers }})</small></span><strong>KES {{ formatMoney(Number(item.price_minor ?? item.priceMinor) * item.quantity) }}</strong></div>
        <div class="tab-totals"><div><span>Subtotal</span><span>KES {{ formatMoney(selectedTable.total_minor) }}</span></div><div><span>Catering Levy &amp; Service</span><span>KES {{ formatMoney(Math.round(Number(selectedTable.total_minor || 0) * 0.1)) }}</span></div><div class="grand"><strong>Total Bill</strong><strong>KES {{ formatMoney(Math.round(Number(selectedTable.total_minor || 0) * 1.1)) }}</strong></div></div>
        <div class="drawer-actions"><button type="button" class="btn-primary" @click="transferToPosCart"><PosIcon name="shopping-cart" size="16" /> Send to Checkout Till</button><button v-if="selectedTable.status === 'occupied'" type="button" class="btn-secondary" @click="markAwaitingBill"><PosIcon name="receipt" size="16" /> Request Bill</button><button type="button" class="btn-danger-outline" @click="clearTableAction"><PosIcon name="trash" size="16" /> Vacate &amp; Reset</button></div>
      </div>
    </div>
  </section>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { PosIcon } from '@pos/ui'
import { useApi } from '../../../../apps/pos/src/composables/useApi.js'

const emit = defineEmits(['notification'])
const api = useApi()
const tables = ref([])
const sections = ref([{ id: 'all', name: 'All Areas' }])
const servers = ref([])
const activeSection = ref('all')
const selectedTable = ref(null)
const loading = ref(true)
const error = ref('')
const seatForm = ref({ guests: 1, server: '', guest_name: '' })
const filteredTables = computed(() => activeSection.value === 'all' ? tables.value : tables.value.filter(table => table.section_id === activeSection.value || table.section === activeSection.value))
const totalGuests = computed(() => tables.value.reduce((total, table) => total + Number(table.guests || 0), 0))

onMounted(loadTables)
async function loadTables() {
  try {
    const data = await api.get('/api/v1/restaurant/tables')
    tables.value = data.tables ?? []
    const areas = [...new Map(tables.value.map(table => [table.section_id || table.section, { id: table.section_id || table.section, name: table.section_name || table.section }])).values()].filter(area => area.id)
    sections.value = [{ id: 'all', name: 'All Areas' }, ...areas]
    servers.value = data.servers ?? []
  } catch {
    error.value = 'Table service could not be loaded. Please contact your administrator.'
  } finally { loading.value = false }
}
function countTablesBySection(id) { return id === 'all' ? tables.value.length : tables.value.filter(table => (table.section_id || table.section) === id).length }
function countStatus(status) { return tables.value.filter(table => table.status === status).length }
function formatStatus(status) { return ({ available: 'Available', occupied: 'Occupied', billing: 'Bill Requested', reserved: 'Reserved' })[status] || status }
function formatMoney(minor) { return (Number(minor || 0) / 100).toLocaleString() }
function selectTable(table) { selectedTable.value = table; seatForm.value.guests = Math.min(1, table.capacity); seatForm.value.server = servers.value[0]?.name || '' }
async function updateTable(id, body, successMessage) {
  try { const data = await api.patch(`/api/v1/restaurant/tables/${id}`, body); const updated = data.table ?? data; const index = tables.value.findIndex(table => table.id === id); if (index >= 0) tables.value[index] = updated; selectedTable.value = updated; emit('notification', { message: successMessage, type: 'success' }) } catch { emit('notification', { message: 'Table update could not be completed. Please contact your administrator.', type: 'error' }) }
}
function seatTableAction() { if (selectedTable.value) updateTable(selectedTable.value.id, { status: 'occupied', guests: seatForm.value.guests, server: seatForm.value.server, guest_name: seatForm.value.guest_name }, `Table ${selectedTable.value.number} opened for service`) }
function markAwaitingBill() { if (selectedTable.value) updateTable(selectedTable.value.id, { status: 'billing' }, 'The bill has been requested') }
function clearTableAction() { if (selectedTable.value) { const number = selectedTable.value.number; updateTable(selectedTable.value.id, { status: 'available', guests: 0, server: null, guest_name: null }, `Table ${number} is available`) } }
function transferToPosCart() {
  if (!selectedTable.value?.items?.length) { emit('notification', { message: 'There are no items on this table to check out', type: 'error' }); return }
  const items = selectedTable.value.items.map((item, index) => ({ id: `${selectedTable.value.id}-${item.id || index}`, name: item.name, sku: item.sku || `TABLE-${selectedTable.value.number}`, price_minor: Number(item.price_minor ?? item.priceMinor), unit_price_minor: Number(item.price_minor ?? item.priceMinor), quantity: item.quantity, kind: item.kind || 'service' }))
  window.dispatchEvent(new CustomEvent('pos:add-item-batch', { detail: items }))
  emit('notification', { message: 'Table items sent to checkout', type: 'success' })
}
</script>

<style scoped>
.table-service-manager{display:flex;flex-direction:column;gap:16px}.manager-header{display:flex;justify-content:space-between;align-items:flex-start;flex-wrap:wrap;gap:12px}.manager-header h3{display:flex;align-items:center;gap:8px;margin:0;color:#1e293b;font-size:18px}.subtitle{margin:4px 0 0;color:#64748b;font-size:13px}.section-filter{display:flex;gap:6px;flex-wrap:wrap}.filter-pill,.btn-secondary,.btn-danger-outline{padding:7px 10px;border:1px solid #cbd5e1;border-radius:6px;background:#fff;color:#334155;font-size:11px;font-weight:600;cursor:pointer}.filter-pill.active{border-color:#ea580c;background:#fff7ed;color:#c2410c}.state{padding:28px;border:1px dashed #cbd5e1;border-radius:8px;color:#64748b;text-align:center;font-size:12px}.state.error{border-color:#fecaca;color:#b91c1c}.occupancy-stats{display:grid;grid-template-columns:repeat(5,1fr);gap:10px}.stat-box{padding:11px;border:1px solid #e2e8f0;border-radius:7px}.stat-box span,.stat-box strong{display:block}.stat-box span{color:#64748b;font-size:10px}.stat-box strong{margin-top:4px;color:#1e293b;font-size:16px}.stat-box.occupied strong{color:#ea580c}.stat-box.billing strong{color:#ca8a04}.stat-box.available strong{color:#16a34a}.tables-grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(210px,1fr));gap:12px}.table-card{padding:14px;border:1px solid #e2e8f0;border-radius:9px;background:#fff;text-align:left;cursor:pointer}.table-card.selected,.table-card.occupied{border-color:#fdba74}.table-top,.table-details,.table-amount,.drawer-header,.tab-info-strip,.drawer-actions,.tab-item-row,.tab-totals div{display:flex;align-items:center;gap:8px}.table-top,.table-amount,.drawer-header,.tab-info-strip,.tab-item-row,.tab-totals div{justify-content:space-between}.status-badge{padding:3px 6px;border-radius:4px;font-size:9px;font-weight:700}.status-badge.available{background:#dcfce7;color:#166534}.status-badge.occupied{background:#ffedd5;color:#c2410c}.status-badge.billing{background:#fef3c7;color:#92400e}.status-badge.reserved{background:#e0e7ff;color:#3730a3}.table-details{flex-direction:column;align-items:flex-start;margin:14px 0;color:#64748b;font-size:11px}.table-details span{display:flex;align-items:center;gap:5px}.table-amount{padding-top:10px;border-top:1px solid #f1f5f9}.table-amount small{color:#64748b}.table-amount strong{color:#ea580c}.table-empty-hint{padding-top:10px;border-top:1px solid #f1f5f9;color:#64748b;font-size:11px}.table-detail-drawer{padding:18px;border:1px solid #e2e8f0;border-radius:10px;background:#fff}.drawer-header{padding-bottom:12px;border-bottom:1px solid #f1f5f9}.drawer-header h4{margin:0;color:#1e293b}.close-btn{border:0;background:transparent;cursor:pointer}.seat-form,.tab-management{display:flex;flex-direction:column;gap:14px;padding-top:16px}.seat-form label{display:flex;flex-direction:column;gap:6px;color:#475569;font-size:11px;font-weight:700}.seat-form input,.seat-form select{padding:9px;border:1px solid #cbd5e1;border-radius:6px}.tab-info-strip{padding:10px;background:#f8fafc;color:#64748b;font-size:11px}.tab-item-row{padding:9px 0;border-bottom:1px solid #f1f5f9;color:#334155;font-size:12px}.tab-item-row small{color:#64748b}.tab-totals{display:flex;flex-direction:column;gap:8px;padding-top:10px;color:#64748b;font-size:12px}.tab-totals .grand{padding-top:8px;border-top:1px solid #cbd5e1;color:#1e293b}.drawer-actions{flex-wrap:wrap}.btn-primary{display:inline-flex;align-items:center;gap:6px;padding:9px 12px;border:0;border-radius:6px;background:#ea580c;color:#fff;font-size:11px;font-weight:700;cursor:pointer}.btn-danger-outline{color:#b91c1c;border-color:#fecaca}.empty-items{color:#64748b;font-size:12px}@media(max-width:700px){.occupancy-stats{grid-template-columns:repeat(2,1fr)}}
</style>
