<template>
  <div class="table-service-manager">
    <div class="manager-header">
      <div>
        <h3>
          <PosIcon name="layout-grid" size="20" class="header-icon" />
          Floor Plan & Table Service
        </h3>
        <p class="subtitle">Real-time table occupancy, guest covers, and active order tabs</p>
      </div>
      <div class="header-actions">
        <div class="section-filter">
          <button
            v-for="sec in sections"
            :key="sec.id"
            type="button"
            class="filter-pill"
            :class="{ 'active': activeSection === sec.id }"
            @click="activeSection = sec.id"
          >
            {{ sec.name }} ({{ countTablesBySection(sec.id) }})
          </button>
        </div>
      </div>
    </div>

    <!-- Occupancy stats summary bar -->
    <div class="occupancy-stats">
      <div class="stat-box">
        <span class="stat-label">Total Tables</span>
        <strong class="stat-value">{{ tables.length }}</strong>
      </div>
      <div class="stat-box occupied">
        <span class="stat-label">Occupied / Dining</span>
        <strong class="stat-value">{{ tables.filter(t => t.status === 'occupied').length }}</strong>
      </div>
      <div class="stat-box billing">
        <span class="stat-label">Awaiting Bill</span>
        <strong class="stat-value">{{ tables.filter(t => t.status === 'billing').length }}</strong>
      </div>
      <div class="stat-box available">
        <span class="stat-label">Available</span>
        <strong class="stat-value">{{ tables.filter(t => t.status === 'available').length }}</strong>
      </div>
      <div class="stat-box covers">
        <span class="stat-label">Active Covers</span>
        <strong class="stat-value">{{ totalGuests }} Guests</strong>
      </div>
    </div>

    <!-- Tables Grid -->
    <div class="tables-grid">
      <div
        v-for="table in filteredTables"
        :key="table.id"
        class="table-card"
        :class="[table.status, { 'selected': selectedTable?.id === table.id }]"
        @click="selectTable(table)"
      >
        <div class="table-top">
          <span class="table-number">Table {{ table.number }}</span>
          <span class="status-badge" :class="table.status">
            {{ formatStatus(table.status) }}
          </span>
        </div>

        <div class="table-details">
          <div class="detail-row">
            <PosIcon name="users" size="14" />
            <span>{{ table.guests }}/{{ table.capacity }} seats</span>
          </div>
          <div class="detail-row" v-if="table.server">
            <PosIcon name="user" size="14" />
            <span>Server: {{ table.server }}</span>
          </div>
          <div class="detail-row" v-if="table.seatedAt">
            <PosIcon name="clock" size="14" />
            <span>{{ table.seatedAt }}</span>
          </div>
        </div>

        <div class="table-amount" v-if="table.status !== 'available'">
          <span class="amount-label">Current Tab</span>
          <span class="amount-value">KES {{ (table.totalMinor / 100).toLocaleString() }}</span>
        </div>
        <div class="table-empty-hint" v-else>
          <span>Tap to seat guests</span>
        </div>
      </div>
    </div>

    <!-- Table Details / Operations Panel Modal or Drawer -->
    <div v-if="selectedTable" class="table-detail-drawer">
      <div class="drawer-header">
        <div class="drawer-title">
          <h4>Table {{ selectedTable.number }} &mdash; {{ selectedTable.sectionName }}</h4>
          <span class="status-badge" :class="selectedTable.status">{{ formatStatus(selectedTable.status) }}</span>
        </div>
        <button type="button" class="close-btn" @click="selectedTable = null">
          <PosIcon name="x" size="18" />
        </button>
      </div>

      <div class="drawer-body">
        <!-- If Available: Seat Form -->
        <div v-if="selectedTable.status === 'available'" class="seat-form">
          <h5>Seat New Party:</h5>
          <div class="form-row">
            <label>Number of Guests (Max {{ selectedTable.capacity }}):</label>
            <input type="number" v-model.number="seatForm.guests" :min="1" :max="selectedTable.capacity" class="form-input" />
          </div>
          <div class="form-row">
            <label>Assigned Server:</label>
            <select v-model="seatForm.server" class="form-input">
              <option value="Mary W.">Mary W.</option>
              <option value="Kevin O.">Kevin O.</option>
              <option value="Alice N.">Alice N.</option>
              <option value="Brian M.">Brian M.</option>
            </select>
          </div>
          <div class="form-row">
            <label>Guest / Booking Name (Optional):</label>
            <input type="text" v-model="seatForm.guestName" placeholder="e.g. Dr. Omondi (Reservation)" class="form-input" />
          </div>
          <button type="button" class="btn-primary" @click="seatTableAction">
            <PosIcon name="check" size="16" />
            Open Table &amp; Start Tab
          </button>
        </div>

        <!-- If Occupied or Billing: View Tab & Operations -->
        <div v-else class="tab-management">
          <div class="tab-info-strip">
            <div>
              <span class="lbl">Server:</span>
              <strong>{{ selectedTable.server }}</strong>
            </div>
            <div>
              <span class="lbl">Guests:</span>
              <strong>{{ selectedTable.guests }} covers</strong>
            </div>
            <div>
              <span class="lbl">Duration:</span>
              <strong>{{ selectedTable.seatedAt }}</strong>
            </div>
          </div>

          <div class="tab-items-list">
            <h5>Ordered Items:</h5>
            <div v-if="!selectedTable.items || !selectedTable.items.length" class="empty-items">
              No items ordered yet. Add items via modifiers or checkout.
            </div>
            <div v-for="(item, idx) in selectedTable.items" :key="idx" class="tab-item-row">
              <div class="item-meta">
                <span class="qty">{{ item.quantity }}x</span>
                <span class="name">{{ item.name }}</span>
                <span class="mods" v-if="item.modifiers">({{ item.modifiers }})</span>
              </div>
              <strong class="item-price">KES {{ ((item.priceMinor * item.quantity) / 100).toLocaleString() }}</strong>
            </div>
          </div>

          <div class="tab-totals">
            <div class="tot-row">
              <span>Subtotal:</span>
              <span>KES {{ (selectedTable.totalMinor / 100).toLocaleString() }}</span>
            </div>
            <div class="tot-row">
              <span>Catering Levy &amp; Service (10%):</span>
              <span>KES {{ Math.round(selectedTable.totalMinor * 0.1 / 100).toLocaleString() }}</span>
            </div>
            <div class="tot-row grand">
              <span>Total Bill:</span>
              <strong>KES {{ Math.round(selectedTable.totalMinor * 1.1 / 100).toLocaleString() }}</strong>
            </div>
          </div>

          <div class="drawer-actions">
            <button type="button" class="btn-primary" @click="transferToPosCart">
              <PosIcon name="shopping-cart" size="16" />
              Send to Checkout Till
            </button>
            <button
              v-if="selectedTable.status === 'occupied'"
              type="button"
              class="btn-secondary"
              @click="markAwaitingBill"
            >
              <PosIcon name="receipt" size="16" />
              Request Bill
            </button>
            <button type="button" class="btn-secondary" @click="openTransferModal = true">
              <PosIcon name="split" size="16" />
              Transfer Table
            </button>
            <button type="button" class="btn-danger-outline" @click="clearTableAction">
              <PosIcon name="trash" size="16" />
              Vacate &amp; Reset
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Table Transfer Modal -->
    <div v-if="openTransferModal" class="modal-backdrop" @click.self="openTransferModal = false">
      <div class="modal-card">
        <div class="modal-header">
          <h4>Transfer Table {{ selectedTable?.number }}</h4>
          <button class="close-btn" @click="openTransferModal = false"><PosIcon name="x" size="16" /></button>
        </div>
        <p class="modal-desc">Move the active dining tab and all orders to an available table.</p>
        <div class="form-row">
          <label>Select Target Table:</label>
          <select v-model="targetTableId" class="form-input">
            <option
              v-for="t in tables.filter(tbl => tbl.status === 'available' && tbl.id !== selectedTable?.id)"
              :key="t.id"
              :value="t.id"
            >
              Table {{ t.number }} &mdash; {{ t.sectionName }} ({{ t.capacity }} seats)
            </option>
          </select>
        </div>
        <div class="modal-actions">
          <button type="button" class="btn-secondary" @click="openTransferModal = false">Cancel</button>
          <button type="button" class="btn-primary" :disabled="!targetTableId" @click="executeTransfer">Confirm Transfer</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { PosIcon } from '@pos/ui'

const emit = defineEmits(['add-item', 'add-item-batch', 'notification'])

const activeSection = ref('all')
const selectedTable = ref(null)
const openTransferModal = ref(false)
const targetTableId = ref('')

const seatForm = ref({
  guests: 2,
  server: 'Mary W.',
  guestName: '',
})

const sections = [
  { id: 'all', name: 'All Areas' },
  { id: 'main', name: 'Main Dining Room' },
  { id: 'terrace', name: 'Garden Terrace' },
  { id: 'bar', name: 'Lounge & Bar' },
  { id: 'vip', name: 'VIP Suite' },
]

const tables = ref([
  {
    id: 't-1',
    number: '1',
    section: 'main',
    sectionName: 'Main Dining',
    capacity: 4,
    guests: 3,
    status: 'occupied',
    server: 'Mary W.',
    seatedAt: '42 mins ago',
    items: [
      { name: 'Prime T-Bone Steak 400g', priceMinor: 240000, quantity: 2, modifiers: 'Medium Rare, Mushroom Sauce' },
      { name: 'Kunde & Creamed Spinach', priceMinor: 35000, quantity: 2 },
      { name: 'Passion Fruit Mojito', priceMinor: 65000, quantity: 3, modifiers: 'No Sugar' },
    ],
    totalMinor: 745000,
  },
  {
    id: 't-2',
    number: '2',
    section: 'main',
    sectionName: 'Main Dining',
    capacity: 2,
    guests: 2,
    status: 'billing',
    server: 'Kevin O.',
    seatedAt: '1h 10m ago',
    items: [
      { name: 'Grilled Herb Half Chicken', priceMinor: 140000, quantity: 2, modifiers: 'Extra Peri-Peri' },
      { name: 'Chips Masala', priceMinor: 40000, quantity: 2 },
      { name: 'Tusker Cider 500ml', priceMinor: 45000, quantity: 2 },
    ],
    totalMinor: 450000,
  },
  {
    id: 't-3',
    number: '3',
    section: 'main',
    sectionName: 'Main Dining',
    capacity: 6,
    guests: 0,
    status: 'available',
    server: null,
    seatedAt: null,
    items: [],
    totalMinor: 0,
  },
  {
    id: 't-4',
    number: '4',
    section: 'main',
    sectionName: 'Main Dining',
    capacity: 4,
    guests: 0,
    status: 'available',
    server: null,
    seatedAt: null,
    items: [],
    totalMinor: 0,
  },
  {
    id: 't-5',
    number: '5',
    section: 'terrace',
    sectionName: 'Garden Terrace',
    capacity: 4,
    guests: 4,
    status: 'occupied',
    server: 'Mary W.',
    seatedAt: '18 mins ago',
    items: [
      { name: 'Whole Grilled Tilapia', priceMinor: 180000, quantity: 2, modifiers: 'Kachumbari & Ugali' },
      { name: 'Mineral Water 1L', priceMinor: 25000, quantity: 2 },
    ],
    totalMinor: 410000,
  },
  {
    id: 't-6',
    number: '6',
    section: 'terrace',
    sectionName: 'Garden Terrace',
    capacity: 8,
    guests: 0,
    status: 'reserved',
    server: 'Alice N.',
    seatedAt: 'Reserved 1:30 PM',
    items: [],
    totalMinor: 0,
  },
  {
    id: 't-7',
    number: 'Bar 1',
    section: 'bar',
    sectionName: 'Lounge & Bar',
    capacity: 2,
    guests: 1,
    status: 'occupied',
    server: 'Brian M.',
    seatedAt: '10 mins ago',
    items: [
      { name: 'Singleton 12Y Double', priceMinor: 95000, quantity: 1, modifiers: 'On the Rocks' },
      { name: 'Spiced Cashews Bowl', priceMinor: 30000, quantity: 1 },
    ],
    totalMinor: 125000,
  },
  {
    id: 't-8',
    number: 'Bar 2',
    section: 'bar',
    sectionName: 'Lounge & Bar',
    capacity: 2,
    guests: 0,
    status: 'available',
    server: null,
    seatedAt: null,
    items: [],
    totalMinor: 0,
  },
  {
    id: 't-9',
    number: 'VIP 1',
    section: 'vip',
    sectionName: 'VIP Suite',
    capacity: 10,
    guests: 8,
    status: 'occupied',
    server: 'Alice N.',
    seatedAt: '55 mins ago',
    items: [
      { name: 'Tomahawk Steak 1.2kg', priceMinor: 650000, quantity: 1, modifiers: 'Medium, Truffle Butter' },
      { name: 'Moët & Chandon Brut', priceMinor: 1800000, quantity: 1 },
      { name: 'Gourmet Mezze Platter', priceMinor: 220000, quantity: 2 },
    ],
    totalMinor: 2890000,
  },
])

const filteredTables = computed(() => {
  if (activeSection.value === 'all') return tables.value
  return tables.value.filter(t => t.section === activeSection.value)
})

const totalGuests = computed(() => {
  return tables.value.reduce((acc, t) => acc + (t.guests || 0), 0)
})

function countTablesBySection(sectionId) {
  if (sectionId === 'all') return tables.value.length
  return tables.value.filter(t => t.section === sectionId).length
}

function formatStatus(status) {
  switch (status) {
    case 'available': return 'Available'
    case 'occupied': return 'Occupied'
    case 'billing': return 'Bill Requested'
    case 'reserved': return 'Reserved'
    default: return status
  }
}

function selectTable(table) {
  selectedTable.value = table
  if (table.status === 'available') {
    seatForm.value.guests = Math.min(2, table.capacity)
  }
}

function seatTableAction() {
  if (!selectedTable.value) return
  selectedTable.value.status = 'occupied'
  selectedTable.value.guests = seatForm.value.guests
  selectedTable.value.server = seatForm.value.server
  selectedTable.value.seatedAt = 'Just seated'
  selectedTable.value.items = []
  selectedTable.value.totalMinor = 0

  emit('notification', {
    message: `Table ${selectedTable.value.number} opened for ${selectedTable.value.guests} guests (${selectedTable.value.server})`,
    type: 'success',
  })
}

function markAwaitingBill() {
  if (!selectedTable.value) return
  selectedTable.value.status = 'billing'
  emit('notification', {
    message: `Printed pro-forma receipt for Table ${selectedTable.value.number}. Status set to Bill Requested.`,
    type: 'info',
  })
}

function clearTableAction() {
  if (!selectedTable.value) return
  const num = selectedTable.value.number
  selectedTable.value.status = 'available'
  selectedTable.value.guests = 0
  selectedTable.value.server = null
  selectedTable.value.seatedAt = null
  selectedTable.value.items = []
  selectedTable.value.totalMinor = 0
  selectedTable.value = null

  emit('notification', {
    message: `Table ${num} has been cleared and reset to available.`,
    type: 'info',
  })
}

function executeTransfer() {
  if (!selectedTable.value || !targetTableId.value) return
  const target = tables.value.find(t => t.id === targetTableId.value)
  if (!target) return

  target.status = selectedTable.value.status
  target.guests = selectedTable.value.guests
  target.server = selectedTable.value.server
  target.seatedAt = selectedTable.value.seatedAt
  target.items = [...selectedTable.value.items]
  target.totalMinor = selectedTable.value.totalMinor

  const prevNum = selectedTable.value.number
  clearTableAction()
  openTransferModal.value = false
  selectedTable.value = target

  emit('notification', {
    message: `Successfully transferred Table ${prevNum} tab to Table ${target.number}.`,
    type: 'success',
  })
}

function transferToPosCart() {
  if (!selectedTable.value || !selectedTable.value.items || !selectedTable.value.items.length) {
    emit('notification', { message: 'Table has no items to checkout', type: 'error' })
    return
  }

  const itemsToCart = selectedTable.value.items.map((it, idx) => ({
    id: `table-${selectedTable.value.id}-${idx}-${Date.now()}`,
    name: `[T${selectedTable.value.number}] ${it.name}${it.modifiers ? ` (${it.modifiers})` : ''}`,
    sku: `DINE-T${selectedTable.value.number}`,
    price_minor: it.priceMinor,
    unit_price_minor: it.priceMinor,
    quantity: it.quantity,
    kind: 'service',
  }))

  // Save to pending queue and emit
  try {
    const existing = JSON.parse(localStorage.getItem('pos.retailCartPending') || '[]')
    localStorage.setItem('pos.retailCartPending', JSON.stringify([...existing, ...itemsToCart]))
  } catch (e) {}

  window.dispatchEvent(new CustomEvent('pos:add-item-batch', { detail: itemsToCart }))

  emit('notification', {
    message: `Transferred ${itemsToCart.length} orders from Table ${selectedTable.value.number} directly to POS Till Cart`,
    type: 'success',
  })
}
</script>

<style scoped>
.table-service-manager {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.manager-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  flex-wrap: wrap;
  gap: 12px;
}

.manager-header h3 {
  margin: 0;
  font-size: 18px;
  color: #1e293b;
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-icon {
  color: #ea580c;
}

.subtitle {
  margin: 4px 0 0;
  font-size: 13px;
  color: #64748b;
}

.section-filter {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.filter-pill {
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

.filter-pill:hover {
  background: #f1f5f9;
}

.filter-pill.active {
  background: #ea580c;
  color: white;
  border-color: #ea580c;
}

.occupancy-stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(130px, 1fr));
  gap: 10px;
}

.stat-box {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.stat-label {
  font-size: 11px;
  color: #64748b;
  text-transform: uppercase;
  font-weight: 700;
}

.stat-value {
  font-size: 20px;
  color: #0f172a;
}

.stat-box.occupied .stat-value { color: #dc2626; }
.stat-box.billing .stat-value { color: #d97706; }
.stat-box.available .stat-value { color: #16a34a; }
.stat-box.covers .stat-value { color: #ea580c; }

.tables-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(190px, 1fr));
  gap: 14px;
}

.table-card {
  background: white;
  border: 2px solid #e2e8f0;
  border-radius: 12px;
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  cursor: pointer;
  transition: transform 0.15s ease, box-shadow 0.15s ease, border-color 0.15s ease;
}

.table-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
}

.table-card.available {
  border-color: #bbf7d0;
  background: #f0fdf4;
}

.table-card.occupied {
  border-color: #fecaca;
  background: #fef2f2;
}

.table-card.billing {
  border-color: #fde68a;
  background: #fffbeb;
}

.table-card.reserved {
  border-color: #cbd5e1;
  background: #f8fafc;
}

.table-card.selected {
  outline: 2px solid #ea580c;
  box-shadow: 0 0 0 4px rgba(234, 88, 12, 0.15);
}

.table-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.table-number {
  font-weight: 700;
  font-size: 16px;
  color: #1e293b;
}

.status-badge {
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  padding: 3px 8px;
  border-radius: 6px;
}

.status-badge.available { background: #dcfce7; color: #15803d; }
.status-badge.occupied { background: #fee2e2; color: #b91c1c; }
.status-badge.billing { background: #fef3c7; color: #b45309; }
.status-badge.reserved { background: #e2e8f0; color: #475569; }

.table-details {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 12px;
  color: #475569;
}

.detail-row {
  display: flex;
  align-items: center;
  gap: 6px;
}

.table-amount {
  border-top: 1px dashed #cbd5e1;
  padding-top: 8px;
  display: flex;
  justify-content: space-between;
  align-items: baseline;
}

.amount-label {
  font-size: 11px;
  color: #64748b;
}

.amount-value {
  font-size: 13px;
  font-weight: 700;
  color: #0f172a;
}

.table-empty-hint {
  border-top: 1px dashed #cbd5e1;
  padding-top: 6px;
  font-size: 11px;
  color: #15803d;
  font-style: italic;
}

/* Detail Drawer */
.table-detail-drawer {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 18px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.08);
}

.drawer-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  border-bottom: 1px solid #f1f5f9;
  padding-bottom: 10px;
}

.drawer-title {
  display: flex;
  align-items: center;
  gap: 10px;
}

.drawer-title h4 {
  margin: 0;
  font-size: 16px;
  color: #0f172a;
}

.close-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: #94a3b8;
  padding: 4px;
  border-radius: 4px;
}

.close-btn:hover { color: #334155; background: #f1f5f9; }

.form-row {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-bottom: 12px;
}

.form-row label {
  font-size: 12px;
  font-weight: 600;
  color: #475569;
}

.form-input {
  padding: 8px 10px;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 13px;
}

.tab-info-strip {
  display: flex;
  gap: 20px;
  background: #f8fafc;
  padding: 10px 14px;
  border-radius: 8px;
  margin-bottom: 14px;
  font-size: 12px;
}

.tab-info-strip .lbl {
  color: #64748b;
  margin-right: 4px;
}

.tab-items-list h5 {
  margin: 0 0 8px;
  font-size: 13px;
  color: #334155;
}

.tab-item-row {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  padding: 6px 0;
  border-bottom: 1px solid #f1f5f9;
  font-size: 13px;
}

.item-meta {
  display: flex;
  gap: 8px;
  align-items: center;
}

.item-meta .qty {
  font-weight: 700;
  color: #ea580c;
}

.item-meta .mods {
  font-size: 11px;
  color: #64748b;
  font-style: italic;
}

.tab-totals {
  margin-top: 14px;
  padding-top: 10px;
  border-top: 2px solid #e2e8f0;
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 12px;
  color: #64748b;
}

.tot-row {
  display: flex;
  justify-content: space-between;
}

.tot-row.grand {
  font-size: 14px;
  color: #0f172a;
  padding-top: 6px;
  border-top: 1px dashed #cbd5e1;
}

.tot-row.grand strong {
  font-size: 16px;
  color: #ea580c;
}

.drawer-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 16px;
}

.btn-primary, .btn-secondary, .btn-danger-outline {
  padding: 8px 14px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  transition: all 0.15s ease;
}

.btn-primary {
  background: #ea580c;
  color: white;
  border: 1px solid #ea580c;
}

.btn-primary:hover {
  background: #c2410c;
}

.btn-secondary {
  background: white;
  border: 1px solid #cbd5e1;
  color: #334155;
}

.btn-secondary:hover {
  background: #f1f5f9;
}

.btn-danger-outline {
  background: white;
  border: 1px solid #fca5a5;
  color: #dc2626;
  margin-left: auto;
}

.btn-danger-outline:hover {
  background: #fee2e2;
}

/* Modal */
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 999;
}

.modal-card {
  background: white;
  border-radius: 12px;
  padding: 20px;
  width: 90%;
  max-width: 440px;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.modal-desc {
  font-size: 13px;
  color: #64748b;
  margin: 0 0 16px;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 18px;
}
</style>
