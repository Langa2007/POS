<template>
  <div class="batch-fefo-manager">
    <div class="manager-header">
      <div>
        <h3>🔬 Batch Tracking & FEFO Dispensing Engine</h3>
        <p class="subtitle">First-Expired, First-Out picking, batch traceability, and quarantine enforcement</p>
      </div>
      <button class="btn-primary" @click="showAddBatchModal = true">
        ➕ Receive New Batch
      </button>
    </div>

    <!-- FEFO Concept Banner -->
    <div class="fefo-banner">
      <div class="fefo-icon">💡</div>
      <div class="fefo-text">
        <strong>FEFO Rule Enforced:</strong> The platform automatically pre-selects the batch closest to expiration date to prevent shelf expiration and regulatory stock write-offs. Expired batches are automatically locked.
      </div>
    </div>

    <!-- Drug Batch Filter & Search -->
    <div class="filter-bar">
      <input
        type="text"
        v-model="searchDrug"
        placeholder="Search medicine or batch number (e.g. Amoxicillin, BTC-2026)..."
        class="search-field"
      />
      <div class="filter-pills">
        <button
          type="button"
          class="pill"
          :class="{ 'active': filterStatus === 'all' }"
          @click="filterStatus = 'all'"
        >
          All Batches ({{ allBatches.length }})
        </button>
        <button
          type="button"
          class="pill near-exp"
          :class="{ 'active': filterStatus === 'near_expiry' }"
          @click="filterStatus = 'near_expiry'"
        >
          ⚠️ Near Expiry &lt;60d ({{ nearExpiryCount }})
        </button>
        <button
          type="button"
          class="pill expired"
          :class="{ 'active': filterStatus === 'expired' }"
          @click="filterStatus = 'expired'"
        >
          🛑 Expired ({{ expiredCount }})
        </button>
      </div>
    </div>

    <!-- Batch Table -->
    <div class="batch-table-container">
      <table class="batch-table">
        <thead>
          <tr>
            <th>Medicine</th>
            <th>Batch No</th>
            <th>Mfg Date</th>
            <th>Expiry Date</th>
            <th>Days Left</th>
            <th>Stock</th>
            <th>FEFO Priority</th>
            <th>Status / Action</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="batch in filteredBatches"
            :key="batch.id"
            :class="batch.statusClass"
          >
            <td>
              <strong>{{ batch.drugName }}</strong>
              <small>{{ batch.strength }} · {{ batch.form }}</small>
            </td>
            <td><code>{{ batch.batchNo }}</code></td>
            <td>{{ batch.mfgDate }}</td>
            <td><strong :class="batch.statusClass">{{ batch.expDate }}</strong></td>
            <td>
              <span v-if="batch.daysLeft > 0">{{ batch.daysLeft }} days</span>
              <span v-else class="text-danger">EXPIRED ({{ Math.abs(batch.daysLeft) }}d ago)</span>
            </td>
            <td><strong>{{ batch.stockUnits }}</strong> {{ batch.form }}s</td>
            <td>
              <span v-if="batch.isRecommendedFefo" class="fefo-badge">
                ⭐ FEFO Pick #1
              </span>
              <span v-else-if="batch.daysLeft > 0" class="fefo-seq">
                Queue #{{ batch.fefoSequence }}
              </span>
              <span v-else class="quarantined-badge">
                🛑 Quarantined
              </span>
            </td>
            <td>
              <div class="action-cell">
                <button
                  v-if="batch.daysLeft > 0"
                  class="btn-pick"
                  :class="{ 'primary': batch.isRecommendedFefo }"
                  @click="pickBatch(batch)"
                >
                  Pick Batch
                </button>
                <button
                  v-else
                  class="btn-quarantine"
                  @click="quarantineBatch(batch)"
                >
                  Write Off
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Receive Batch Modal -->
    <div v-if="showAddBatchModal" class="modal-backdrop">
      <div class="modal-dialog">
        <div class="modal-header">
          <h4>Receive New Medicine Batch</h4>
          <button class="close-btn" @click="showAddBatchModal = false">✕</button>
        </div>
        <form @submit.prevent="saveNewBatch" class="modal-form">
          <label>Medication:
            <select v-model="newBatch.drugName" required>
              <option value="Amoxicillin Capsules 500mg">Amoxicillin Capsules 500mg</option>
              <option value="Augmentin Tablets 625mg">Augmentin Tablets 625mg</option>
              <option value="Panadol Extra Tablets 500mg/65mg">Panadol Extra Tablets 500mg/65mg</option>
              <option value="Codeine Phosphate 30mg">Codeine Phosphate 30mg</option>
              <option value="Metformin HCl 500mg">Metformin HCl 500mg</option>
            </select>
          </label>
          <div class="form-row">
            <label>Batch Number:
              <input type="text" v-model="newBatch.batchNo" required placeholder="e.g. BTC-2026-99" />
            </label>
            <label>Quantity Received:
              <input type="number" v-model.number="newBatch.stockUnits" min="1" required />
            </label>
          </div>
          <div class="form-row">
            <label>Manufacture Date:
              <input type="date" v-model="newBatch.mfgDate" required />
            </label>
            <label>Expiry Date:
              <input type="date" v-model="newBatch.expDate" required />
            </label>
          </div>
          <div class="modal-actions">
            <button type="button" class="btn-secondary" @click="showAddBatchModal = false">Cancel</button>
            <button type="submit" class="btn-primary">Add Batch to Inventory</button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const emit = defineEmits(['notification'])

const searchDrug = ref('')
const filterStatus = ref('all')
const showAddBatchModal = ref(false)

const allBatches = ref([
  {
    id: 'b-1',
    drugName: 'Amoxicillin Capsules',
    strength: '500mg',
    form: 'Capsule',
    batchNo: 'AMX-2025-01',
    mfgDate: '2024-01-15',
    expDate: '2026-11-25',
    daysLeft: 15,
    stockUnits: 120,
    statusClass: 'near-expiry',
    isRecommendedFefo: true,
    fefoSequence: 1,
  },
  {
    id: 'b-2',
    drugName: 'Amoxicillin Capsules',
    strength: '500mg',
    form: 'Capsule',
    batchNo: 'AMX-2026-08',
    mfgDate: '2024-08-10',
    expDate: '2027-08-10',
    daysLeft: 638,
    stockUnits: 500,
    statusClass: 'good',
    isRecommendedFefo: false,
    fefoSequence: 2,
  },
  {
    id: 'b-3',
    drugName: 'Amoxicillin Capsules',
    strength: '500mg',
    form: 'Capsule',
    batchNo: 'AMX-2024-09',
    mfgDate: '2022-09-01',
    expDate: '2026-09-01',
    daysLeft: -70,
    stockUnits: 45,
    statusClass: 'expired',
    isRecommendedFefo: false,
    fefoSequence: 99,
  },
  {
    id: 'b-4',
    drugName: 'Augmentin Tablets',
    strength: '625mg',
    form: 'Tablet',
    batchNo: 'AUG-2026-02',
    mfgDate: '2024-02-14',
    expDate: '2026-12-10',
    daysLeft: 30,
    stockUnits: 80,
    statusClass: 'near-expiry',
    isRecommendedFefo: true,
    fefoSequence: 1,
  },
  {
    id: 'b-5',
    drugName: 'Augmentin Tablets',
    strength: '625mg',
    form: 'Tablet',
    batchNo: 'AUG-2027-04',
    mfgDate: '2024-10-01',
    expDate: '2027-04-01',
    daysLeft: 507,
    stockUnits: 300,
    statusClass: 'good',
    isRecommendedFefo: false,
    fefoSequence: 2,
  },
  {
    id: 'b-6',
    drugName: 'Codeine Phosphate',
    strength: '30mg',
    form: 'Tablet',
    batchNo: 'COD-2026-11',
    mfgDate: '2024-11-01',
    expDate: '2026-11-20',
    daysLeft: 10,
    stockUnits: 40,
    statusClass: 'near-expiry',
    isRecommendedFefo: true,
    fefoSequence: 1,
  },
])

const newBatch = ref({
  drugName: 'Amoxicillin Capsules 500mg',
  batchNo: '',
  stockUnits: 200,
  mfgDate: new Date().toISOString().split('T')[0],
  expDate: new Date(Date.now() + 365 * 86400000).toISOString().split('T')[0],
})

const nearExpiryCount = computed(() => allBatches.value.filter((b) => b.daysLeft > 0 && b.daysLeft <= 60).length)
const expiredCount = computed(() => allBatches.value.filter((b) => b.daysLeft <= 0).length)

const filteredBatches = computed(() => {
  return allBatches.value.filter((b) => {
    const q = searchDrug.value.toLowerCase()
    const matchesQuery = !q || b.drugName.toLowerCase().includes(q) || b.batchNo.toLowerCase().includes(q)
    if (!matchesQuery) return false

    if (filterStatus.value === 'near_expiry') return b.daysLeft > 0 && b.daysLeft <= 60
    if (filterStatus.value === 'expired') return b.daysLeft <= 0
    return true
  })
})

function pickBatch(batch) {
  emit('notification', {
    message: `Selected Batch ${batch.batchNo} for ${batch.drugName} (Expires: ${batch.expDate})`,
    type: 'success',
  })
}

function quarantineBatch(batch) {
  allBatches.value = allBatches.value.filter((b) => b.id !== batch.id)
  emit('notification', {
    message: `Batch ${batch.batchNo} (${batch.stockUnits} units) written off to destruction log.`,
    type: 'warning',
  })
}

function saveNewBatch() {
  const days = Math.round((new Date(newBatch.value.expDate).getTime() - Date.now()) / 86400000)
  const created = {
    id: `b-${Date.now()}`,
    drugName: newBatch.value.drugName,
    strength: '500mg',
    form: 'Unit',
    batchNo: newBatch.value.batchNo.toUpperCase(),
    mfgDate: newBatch.value.mfgDate,
    expDate: newBatch.value.expDate,
    daysLeft: days,
    stockUnits: newBatch.value.stockUnits,
    statusClass: days <= 0 ? 'expired' : days <= 60 ? 'near-expiry' : 'good',
    isRecommendedFefo: false,
    fefoSequence: 3,
  }
  allBatches.value.unshift(created)
  showAddBatchModal.value = false
  emit('notification', {
    message: `Received Batch ${created.batchNo} with ${created.stockUnits} units into pharmacy stock.`,
    type: 'success',
  })
}
</script>

<style scoped>
.batch-fefo-manager {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}
.manager-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.75rem;
}
.manager-header h3 {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 700;
}
.subtitle {
  margin: 0.25rem 0 0;
  color: #64748b;
  font-size: 0.9rem;
}
.btn-primary {
  background: #0284c7;
  color: white;
  border: none;
  padding: 0.55rem 1rem;
  border-radius: 8px;
  font-weight: 700;
  cursor: pointer;
}
.fefo-banner {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  background: #f0fdf4;
  border: 1px solid #bbf7d0;
  padding: 0.85rem 1.25rem;
  border-radius: 8px;
  font-size: 0.85rem;
  color: #166534;
}
.fefo-icon {
  font-size: 1.5rem;
}
.filter-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  flex-wrap: wrap;
}
.search-field {
  flex: 1;
  min-width: 260px;
  padding: 0.5rem 0.75rem;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 0.9rem;
}
.filter-pills {
  display: flex;
  gap: 0.5rem;
}
.pill {
  background: #f1f5f9;
  border: 1px solid #cbd5e1;
  padding: 0.35rem 0.75rem;
  border-radius: 999px;
  font-size: 0.75rem;
  font-weight: 600;
  cursor: pointer;
}
.pill.active {
  background: #0284c7;
  color: white;
  border-color: #0284c7;
}
.pill.near-exp.active {
  background: #f59e0b;
  border-color: #f59e0b;
}
.pill.expired.active {
  background: #dc2626;
  border-color: #dc2626;
}
.batch-table-container {
  overflow-x: auto;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  background: white;
}
.batch-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
  font-size: 0.85rem;
}
.batch-table th, .batch-table td {
  padding: 0.85rem 1rem;
  border-bottom: 1px solid #f1f5f9;
}
.batch-table th {
  background: #f8fafc;
  font-weight: 700;
  color: #475569;
}
.batch-table td small {
  display: block;
  color: #64748b;
}
.batch-table tr.near-expiry {
  background: #fffbeb;
}
.batch-table tr.expired {
  background: #fef2f2;
}
.text-danger {
  color: #dc2626;
  font-weight: bold;
}
.fefo-badge {
  background: #166534;
  color: white;
  padding: 0.2rem 0.55rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 700;
  white-space: nowrap;
}
.fefo-seq {
  color: #64748b;
  font-size: 0.8rem;
}
.quarantined-badge {
  background: #fee2e2;
  color: #991b1b;
  padding: 0.2rem 0.55rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 700;
}
.action-cell {
  display: flex;
  gap: 0.35rem;
}
.btn-pick {
  background: #f1f5f9;
  border: 1px solid #cbd5e1;
  padding: 0.35rem 0.75rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
  cursor: pointer;
}
.btn-pick.primary {
  background: #0284c7;
  color: white;
  border-color: #0284c7;
}
.btn-quarantine {
  background: #fee2e2;
  border: 1px solid #fca5a5;
  color: #991b1b;
  padding: 0.35rem 0.75rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 700;
  cursor: pointer;
}
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.5);
  display: grid;
  place-items: center;
  z-index: 1000;
}
.modal-dialog {
  background: white;
  width: min(520px, 94vw);
  border-radius: 12px;
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.close-btn {
  background: none;
  border: none;
  cursor: pointer;
}
.modal-form {
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
}
.modal-form label {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  font-size: 0.85rem;
  font-weight: 600;
  color: #334155;
}
.modal-form select, .modal-form input {
  padding: 0.5rem 0.75rem;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
}
.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem;
}
.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  margin-top: 0.5rem;
}
.btn-secondary {
  background: #f1f5f9;
  border: 1px solid #cbd5e1;
  padding: 0.5rem 1rem;
  border-radius: 6px;
  cursor: pointer;
}
</style>
