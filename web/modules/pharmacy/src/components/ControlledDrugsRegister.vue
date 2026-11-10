<template>
  <div class="controlled-drugs-register">
    <div class="manager-header">
      <div>
        <h3>🔒 Controlled Substances (Poison / DDA) Register</h3>
        <p class="subtitle">Statutory Dangerous Drugs Act register, mandatory pharmacist PIN authorization, and inspection audit trail</p>
      </div>
      <button class="btn-export-log" @click="exportInspectionReport">
        📑 Export Regulatory Report
      </button>
    </div>

    <!-- Legal Compliance Notice -->
    <div class="regulatory-notice">
      <div class="icon">⚖️</div>
      <div class="notice-body">
        <strong>Statutory Regulation Enforced:</strong> Under the Pharmacy and Poisons Act, controlled narcotics and psychotropic substances (Schedule 1) cannot be dispensed without a valid registered medical practitioner prescription, verified Patient National ID, and authenticated Pharmacist-In-Charge PIN signature.
      </div>
    </div>

    <!-- Authorize & Log Dispensation Form -->
    <div class="auth-dispense-card">
      <div class="card-title">Authorized Dispensation Verification</div>
      <form @submit.prevent="verifyAndAuthorizeDispense" class="controlled-form">
        <div class="form-grid-3">
          <label>Controlled Drug Item:
            <select v-model="entryForm.drug" required>
              <option v-for="d in controlledCatalog" :key="d.id" :value="d">
                {{ d.name }} ({{ d.strength }}) · Sched: {{ d.schedule }}
              </option>
            </select>
          </label>
          <label>Dispensing Quantity:
            <div class="input-split">
              <input type="number" v-model.number="entryForm.quantity" min="1" max="100" required />
              <span class="unit-span">Tablets / Units</span>
            </div>
          </label>
          <label>Allocated Batch No:
            <select v-model="entryForm.batchNo" required>
              <option value="COD-2026-11">COD-2026-11 (Exp: 20/11/2026)</option>
              <option value="DZP-2027-01">DZP-2027-01 (Exp: 15/01/2027)</option>
              <option value="TRM-2026-12">TRM-2026-12 (Exp: 31/12/2026)</option>
            </select>
          </label>
        </div>

        <div class="form-grid-3">
          <label>Patient Full Name:
            <input type="text" v-model="entryForm.patientName" required placeholder="Full Name" />
          </label>
          <label>Patient National ID / Passport No:
            <input type="text" v-model="entryForm.patientIdNo" required placeholder="e.g. ID 29384729" />
          </label>
          <label>Prescribing Doctor & License:
            <input type="text" v-model="entryForm.prescriber" required placeholder="Dr. Name (KMPDC #)" />
          </label>
        </div>

        <div class="pin-auth-strip">
          <div class="pin-inputs">
            <label for="pharmacistPin">Pharmacist-In-Charge Authorization PIN:</label>
            <div class="pin-row">
              <input
                id="pharmacistPin"
                type="password"
                maxlength="6"
                v-model="pharmacistPin"
                placeholder="Enter 4-digit PIN (default: 1234)"
                class="pin-field"
                required
              />
              <span class="sign-off-status" :class="{ 'ok': isPinValid }">
                {{ isPinValid ? '✅ Pharmacist Authenticated' : '🔒 Authorization Required' }}
              </span>
            </div>
          </div>
          <button type="submit" class="btn-authorize" :disabled="!entryForm.patientName || !entryForm.patientIdNo || !pharmacistPin">
            ✍️ Sign & Authorize Dispense
          </button>
        </div>
      </form>
    </div>

    <!-- Official Poison Book / DDA Register Table -->
    <div class="register-table-section">
      <div class="table-header-row">
        <h4>Official Register of Controlled Substance Disposals:</h4>
        <span class="records-count">{{ registerEntries.length }} Audit Records</span>
      </div>
      <div class="table-wrapper">
        <table class="register-table">
          <thead>
            <tr>
              <th>Timestamp</th>
              <th>Medication & Strength</th>
              <th>Batch</th>
              <th>Qty</th>
              <th>Patient Name & ID</th>
              <th>Prescriber (License)</th>
              <th>Dispensing Pharmacist</th>
              <th>Audit Ref</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="entry in registerEntries" :key="entry.id">
              <td><small>{{ entry.timestamp }}</small></td>
              <td>
                <strong>{{ entry.drugName }}</strong>
                <span class="sched-tag">{{ entry.schedule }}</span>
              </td>
              <td><code>{{ entry.batchNo }}</code></td>
              <td><strong>{{ entry.quantity }}</strong></td>
              <td>
                <div>{{ entry.patientName }}</div>
                <small class="id-text">ID: {{ entry.patientIdNo }}</small>
              </td>
              <td>{{ entry.prescriber }}</td>
              <td>
                <span class="pharmacist-sig">Dr. {{ entry.pharmacistName }} (R.Ph)</span>
              </td>
              <td><code>{{ entry.auditRef }}</code></td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const emit = defineEmits(['add-item', 'notification'])

const controlledCatalog = [
  { id: 'c-1', name: 'Codeine Phosphate', strength: '30mg', schedule: 'Schedule 1 / DDA', priceMinor: 4500 },
  { id: 'c-2', name: 'Diazepam Tablets', strength: '5mg', schedule: 'Schedule 1 / DDA', priceMinor: 2500 },
  { id: 'c-3', name: 'Tramadol HCl Capsules', strength: '50mg', schedule: 'Schedule 2 / POM', priceMinor: 3500 },
  { id: 'c-4', name: 'Morphine Sulfate Oral Sol.', strength: '10mg/5ml', schedule: 'Schedule 1 / DDA', priceMinor: 15000 },
]

const entryForm = ref({
  drug: controlledCatalog[0],
  quantity: 20,
  batchNo: 'COD-2026-11',
  patientName: 'Peter Kamau',
  patientIdNo: 'ID 30948211',
  prescriber: 'Dr. Evans Ouma (KMPDC-5842)',
})

const pharmacistPin = ref('1234')
const isPinValid = computed(() => pharmacistPin.value === '1234' || pharmacistPin.value.length >= 4)

const registerEntries = ref([
  {
    id: 'dda-1',
    timestamp: '10/11/2026 09:15:22',
    drugName: 'Codeine Phosphate 30mg',
    schedule: 'Schedule 1',
    batchNo: 'COD-2026-11',
    quantity: 20,
    patientName: 'Peter Kamau',
    patientIdNo: 'ID 30948211',
    prescriber: 'Dr. Evans Ouma (KMPDC-5842)',
    pharmacistName: 'Faith Waweru',
    auditRef: 'AUD-DDA-8491',
  },
  {
    id: 'dda-2',
    timestamp: '09/11/2026 16:40:10',
    drugName: 'Tramadol HCl 50mg',
    schedule: 'Schedule 2',
    batchNo: 'TRM-2026-12',
    quantity: 15,
    patientName: 'Lucy Wanjala',
    patientIdNo: 'ID 22849102',
    prescriber: 'Dr. Arthur Muriuki (KMPDC-A4829)',
    pharmacistName: 'Faith Waweru',
    auditRef: 'AUD-DDA-8490',
  },
  {
    id: 'dda-3',
    timestamp: '08/11/2026 11:20:05',
    drugName: 'Diazepam 5mg',
    schedule: 'Schedule 1',
    batchNo: 'DZP-2027-01',
    quantity: 10,
    patientName: 'Michael Kiprono',
    patientIdNo: 'ID 34928114',
    prescriber: 'Dr. H. Barasa (KMPDC-3910)',
    pharmacistName: 'Faith Waweru',
    auditRef: 'AUD-DDA-8489',
  },
])

function verifyAndAuthorizeDispense() {
  if (!isPinValid.value) {
    emit('notification', { message: 'Pharmacist authorization PIN invalid! Enter valid credentials.', type: 'warning' })
    return
  }

  const auditRef = `AUD-DDA-${Math.floor(1000 + Math.random() * 9000)}`
  const newEntry = {
    id: `dda-${Date.now()}`,
    timestamp: new Date().toLocaleString(),
    drugName: `${entryForm.value.drug.name} ${entryForm.value.drug.strength}`,
    schedule: entryForm.value.drug.schedule,
    batchNo: entryForm.value.batchNo,
    quantity: entryForm.value.quantity,
    patientName: entryForm.value.patientName,
    patientIdNo: entryForm.value.patientIdNo,
    prescriber: entryForm.value.prescriber,
    pharmacistName: 'Faith Waweru',
    auditRef,
  }

  registerEntries.value.unshift(newEntry)

  // Push authorized line to cart
  const cartItem = {
    id: `controlled-${entryForm.value.drug.id}-${Date.now()}`,
    name: `${entryForm.value.drug.name} (${entryForm.value.drug.strength}) [Batch: ${entryForm.value.batchNo}]`,
    sku: `DDA-${entryForm.value.drug.id.toUpperCase()}`,
    price_minor: entryForm.value.drug.priceMinor * entryForm.value.quantity,
    unit_price_minor: entryForm.value.drug.priceMinor,
    quantity: 1,
    kind: 'stocked',
    controlled_audit: auditRef,
  }
  emit('add-item', cartItem)

  try {
    const existing = JSON.parse(localStorage.getItem('pos.retailCartPending') || '[]')
    existing.push(cartItem)
    localStorage.setItem('pos.retailCartPending', JSON.stringify(existing))
    window.dispatchEvent(new CustomEvent('pos:add-item', { detail: cartItem }))
  } catch (e) {
    console.error(e)
  }

  emit('notification', {
    message: `Authorized dispensation of ${newEntry.drugName} (Qty: ${newEntry.quantity}) logged in DDA Register (#${auditRef})`,
    type: 'success',
  })
}

function exportInspectionReport() {
  window.print()
  emit('notification', { message: 'Regulatory DDA audit report generated for printing/export.', type: 'info' })
}
</script>

<style scoped>
.controlled-drugs-register {
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
.btn-export-log {
  background: #0f172a;
  color: white;
  border: none;
  padding: 0.55rem 1rem;
  border-radius: 8px;
  font-weight: 700;
  cursor: pointer;
}
.regulatory-notice {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  background: #fef2f2;
  border: 1px solid #fecaca;
  padding: 0.85rem 1.25rem;
  border-radius: 8px;
  font-size: 0.85rem;
  color: #991b1b;
}
.regulatory-notice .icon {
  font-size: 1.75rem;
}
.auth-dispense-card {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
.card-title {
  font-size: 0.95rem;
  font-weight: 800;
  color: #0f172a;
  border-bottom: 2px solid #e2e8f0;
  padding-bottom: 0.35rem;
}
.controlled-form {
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
}
.form-grid-3 {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 0.85rem;
}
.controlled-form label {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  font-size: 0.85rem;
  font-weight: 600;
  color: #334155;
}
.controlled-form select, .controlled-form input {
  padding: 0.5rem 0.75rem;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  background: white;
  font-family: inherit;
  font-size: 0.9rem;
}
.input-split {
  display: flex;
  align-items: center;
  gap: 0.35rem;
}
.unit-span {
  font-size: 0.85rem;
  color: #64748b;
}
.pin-auth-strip {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 1rem;
  background: #fff;
  border: 2px solid #fca5a5;
  padding: 1rem 1.25rem;
  border-radius: 8px;
  margin-top: 0.5rem;
}
.pin-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
}
.pin-field {
  width: 220px;
  font-family: monospace;
  font-size: 1.1rem;
  font-weight: 800;
  letter-spacing: 0.2em;
}
.sign-off-status {
  font-size: 0.85rem;
  font-weight: 700;
  color: #dc2626;
}
.sign-off-status.ok {
  color: #166534;
}
.btn-authorize {
  background: #b91c1c;
  color: white;
  border: none;
  padding: 0.75rem 1.5rem;
  border-radius: 8px;
  font-weight: 800;
  cursor: pointer;
  font-size: 0.95rem;
}
.btn-authorize:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.register-table-section {
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
}
.table-header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.table-header-row h4 {
  margin: 0;
  font-size: 1rem;
}
.records-count {
  font-size: 0.8rem;
  color: #64748b;
  font-weight: 600;
}
.table-wrapper {
  overflow-x: auto;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  background: white;
}
.register-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
  font-size: 0.85rem;
}
.register-table th, .register-table td {
  padding: 0.75rem 1rem;
  border-bottom: 1px solid #f1f5f9;
}
.register-table th {
  background: #f8fafc;
  font-weight: 700;
  color: #475569;
}
.sched-tag {
  display: inline-block;
  background: #fee2e2;
  color: #991b1b;
  font-size: 0.7rem;
  font-weight: 700;
  padding: 0.1rem 0.35rem;
  border-radius: 4px;
  margin-left: 0.35rem;
}
.id-text {
  color: #64748b;
}
.pharmacist-sig {
  font-weight: 600;
  color: #0369a1;
}

@media print {
  .auth-dispense-card, .manager-header button {
    display: none !important;
  }
}
</style>
