<template>
  <aside class="pharmacy-checkout-sidebar">
    <div class="sidebar-header">
      <span class="header-icon">💊</span>
      <div>
        <h4>Pharmacy Dispensing</h4>
        <small>Prescriptions & FEFO Batches</small>
      </div>
    </div>

    <!-- 1. Quick Rx Intake -->
    <section class="sidebar-block">
      <div class="block-title">
        <span>📋 Clinical Prescription</span>
        <button class="btn-micro" @click="showRxModal = true">Intake Rx</button>
      </div>
      <div class="rx-status-box">
        <p v-if="dispensedCount > 0" class="rx-active-note">
          ✅ {{ dispensedCount }} Rx medication(s) in active cart
        </p>
        <p v-else class="rx-empty-note">
          No prescription attached to current sale. Click Intake Rx to add clinical medications.
        </p>
      </div>
    </section>

    <!-- 2. FEFO Batch Selector for Cart -->
    <section class="sidebar-block">
      <div class="block-title">
        <span>🔬 FEFO Batch Allocation</span>
      </div>
      <div class="batch-select-box">
        <div class="batch-info">
          <span>Active Batch: <strong>AMX-2025-01</strong></span>
          <span class="badge-fefo">⭐ FEFO 15d left</span>
        </div>
        <small class="fefo-desc">System auto-allocates earliest expiring batch on checkout.</small>
      </div>
    </section>

    <!-- 3. Controlled Substance Authorization -->
    <section class="sidebar-block">
      <div class="block-title">
        <span>🔒 Controlled Drug Sign-off</span>
      </div>
      <div class="dda-status-box">
        <span class="auth-pill">R.Ph Authorized (Dr. F. Waweru)</span>
        <small>Poison Book DDA Register active.</small>
      </div>
    </section>

    <!-- 4. Clinical Label Print -->
    <section class="sidebar-block">
      <div class="block-title">
        <span>🏷️ Dispensing Label</span>
      </div>
      <button type="button" class="btn-print-rx-label" @click="printCurrentLabel">
        🖨️ Print Rx Label for Sale
      </button>
    </section>

    <!-- Quick Rx Intake Modal -->
    <div v-if="showRxModal" class="rx-modal-backdrop">
      <div class="rx-modal-dialog">
        <div class="modal-top">
          <h5>Quick Prescription Intake</h5>
          <button class="close-x" @click="showRxModal = false">✕</button>
        </div>
        <div class="rx-modal-form">
          <label>Patient Name:
            <input type="text" v-model="quickRx.patientName" placeholder="Patient Full Name" />
          </label>
          <label>Prescribing Doctor:
            <input type="text" v-model="quickRx.doctorName" placeholder="Dr. Name (Hospital)" />
          </label>
          <label>Medication:
            <select v-model="quickRx.selectedDrug">
              <option v-for="d in drugList" :key="d.id" :value="d">
                {{ d.name }} ({{ d.strength }}) · KES {{ (d.priceMinor / 100).toFixed(0) }}
              </option>
            </select>
          </label>
          <div class="row-2">
            <label>Dosage Regimen:
              <input type="text" v-model="quickRx.directions" placeholder="e.g. 1 tab TDS pc" />
            </label>
            <label>Quantity:
              <input type="number" v-model.number="quickRx.quantity" min="1" max="100" />
            </label>
          </div>
        </div>
        <div class="modal-buttons">
          <button class="btn-modal-cancel" @click="showRxModal = false">Cancel</button>
          <button class="btn-modal-add" @click="dispenseQuickRx">
            ➕ Dispense to Cart
          </button>
        </div>
      </div>
    </div>
  </aside>
</template>

<script setup>
import { ref } from 'vue'

const emit = defineEmits(['add-item', 'notification'])

const showRxModal = ref(false)
const dispensedCount = ref(0)

const drugList = [
  { id: 'd-1', name: 'Amoxicillin Capsules', strength: '500mg', priceMinor: 2500 },
  { id: 'd-2', name: 'Augmentin Tablets', strength: '625mg', priceMinor: 12000 },
  { id: 'd-3', name: 'Panadol Extra Tablets', strength: '500mg', priceMinor: 1500 },
  { id: 'd-4', name: 'Codeine Phosphate', strength: '30mg', priceMinor: 4500 },
]

const quickRx = ref({
  patientName: 'Sarah Mwangi',
  doctorName: 'Dr. Arthur Muriuki',
  selectedDrug: drugList[0],
  directions: '1 cap TDS x 5 days (after meals)',
  quantity: 15,
})

function dispenseQuickRx() {
  if (!quickRx.value.selectedDrug) return
  const drug = quickRx.value.selectedDrug
  const totalMinor = drug.priceMinor * quickRx.value.quantity

  const cartItem = {
    id: `rx-${drug.id}-${Date.now()}`,
    name: `${drug.name} (${drug.strength}) [Rx: ${quickRx.value.patientName}]`,
    sku: `RX-${drug.id.toUpperCase()}`,
    price_minor: totalMinor,
    unit_price_minor: drug.priceMinor,
    quantity: 1,
    kind: 'stocked',
    notes: quickRx.value.directions,
  }

  emit('add-item', cartItem)
  try {
    const list = JSON.parse(localStorage.getItem('pos.retailCartPending') || '[]')
    list.push(cartItem)
    localStorage.setItem('pos.retailCartPending', JSON.stringify(list))
    window.dispatchEvent(new CustomEvent('pos:add-item', { detail: cartItem }))
  } catch (e) {
    console.error(e)
  }

  dispensedCount.value++
  showRxModal.value = false
}

function printCurrentLabel() {
  window.print()
}
</script>

<style scoped>
.pharmacy-checkout-sidebar {
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
  background: #f0fdf4;
  border: 1px solid #bbf7d0;
  border-radius: 12px;
  padding: 1rem;
}
.sidebar-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid #bbf7d0;
}
.header-icon {
  font-size: 1.35rem;
}
.sidebar-header h4 {
  margin: 0;
  font-size: 0.95rem;
  font-weight: 800;
  color: #166534;
}
.sidebar-header small {
  color: #15803d;
  font-size: 0.75rem;
}
.sidebar-block {
  background: white;
  border: 1px solid #bbf7d0;
  border-radius: 8px;
  padding: 0.75rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
.block-title {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.8rem;
  font-weight: 700;
  color: #166534;
}
.btn-micro {
  background: #dcfce7;
  color: #166534;
  border: 1px solid #86efac;
  padding: 0.2rem 0.45rem;
  border-radius: 4px;
  font-size: 0.7rem;
  font-weight: 700;
  cursor: pointer;
}
.rx-status-box {
  font-size: 0.75rem;
}
.rx-active-note {
  color: #166534;
  font-weight: 700;
  margin: 0;
}
.rx-empty-note {
  color: #64748b;
  margin: 0;
}
.batch-select-box {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}
.batch-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.8rem;
}
.badge-fefo {
  background: #166534;
  color: white;
  font-size: 0.7rem;
  font-weight: 800;
  padding: 0.15rem 0.4rem;
  border-radius: 4px;
}
.fefo-desc {
  font-size: 0.7rem;
  color: #64748b;
}
.dda-status-box {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}
.auth-pill {
  font-size: 0.75rem;
  font-weight: 700;
  color: #0369a1;
  background: #e0f2fe;
  padding: 0.15rem 0.4rem;
  border-radius: 4px;
}
.btn-print-rx-label {
  background: #166534;
  color: white;
  border: none;
  padding: 0.45rem 0.75rem;
  border-radius: 6px;
  font-weight: 700;
  font-size: 0.8rem;
  cursor: pointer;
}
.rx-modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.5);
  display: grid;
  place-items: center;
  z-index: 1100;
}
.rx-modal-dialog {
  background: white;
  width: min(420px, 94vw);
  border-radius: 12px;
  padding: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
.modal-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.modal-top h5 {
  margin: 0;
  font-size: 1.1rem;
}
.close-x {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 1.1rem;
}
.rx-modal-form {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}
.rx-modal-form label {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  font-size: 0.8rem;
  font-weight: 600;
  color: #334155;
}
.rx-modal-form input, .rx-modal-form select {
  padding: 0.45rem 0.65rem;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 0.85rem;
}
.row-2 {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: 0.5rem;
}
.modal-buttons {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}
.btn-modal-cancel {
  background: #f1f5f9;
  border: 1px solid #cbd5e1;
  padding: 0.5rem 1rem;
  border-radius: 6px;
  font-weight: 600;
  cursor: pointer;
}
.btn-modal-add {
  background: #166534;
  color: white;
  border: none;
  padding: 0.5rem 1rem;
  border-radius: 6px;
  font-weight: 700;
  cursor: pointer;
}
</style>
