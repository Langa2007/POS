<template>
  <div class="prescription-station">
    <div class="manager-header">
      <div>
        <h3>📋 Prescription Verification & Dispensing Station</h3>
        <p class="subtitle">Clinical prescription intake, dosage calculation, and direct dispensing to POS cart</p>
      </div>
      <button class="btn-new-rx" @click="resetForm">
        ➕ New Prescription Intake
      </button>
    </div>

    <!-- Main Intake Form -->
    <div class="rx-intake-card">
      <div class="form-section-title">1. Patient & Prescriber Details</div>
      <div class="form-grid-3">
        <label>Patient Full Name:
          <input type="text" v-model="rxForm.patientName" placeholder="e.g. Sarah Mwangi" required />
        </label>
        <label>Patient Age / Gender:
          <div class="input-split">
            <input type="number" v-model.number="rxForm.patientAge" placeholder="Age" min="0" max="120" />
            <select v-model="rxForm.patientGender">
              <option value="Female">Female</option>
              <option value="Male">Male</option>
              <option value="Child">Child / Infant</option>
            </select>
          </div>
        </label>
        <label>Patient ID / File / Insurance No:
          <input type="text" v-model="rxForm.patientId" placeholder="e.g. SHA-99201 or ID 28472911" />
        </label>
      </div>

      <div class="form-grid-3">
        <label>Prescribing Doctor:
          <input type="text" v-model="rxForm.doctorName" placeholder="e.g. Dr. Arthur Muriuki" required />
        </label>
        <label>Hospital / Clinic:
          <input type="text" v-model="rxForm.clinicName" placeholder="e.g. Kenyatta National Hospital" />
        </label>
        <label>Doctor KMPDC Reg No:
          <input type="text" v-model="rxForm.doctorRegNo" placeholder="e.g. KMPDC-A4829" required />
        </label>
      </div>

      <!-- Prescribed Medications Section -->
      <div class="form-section-title">2. Prescribed Medications</div>
      <div class="medications-builder">
        <div v-for="(med, index) in rxForm.medications" :key="index" class="med-line-card">
          <div class="med-line-top">
            <span class="med-badge">Rx Item #{{ index + 1 }}</span>
            <button v-if="rxForm.medications.length > 1" type="button" class="btn-remove-med" @click="removeMed(index)">✕ Remove</button>
          </div>

          <div class="form-grid-4">
            <label class="col-span-2">Medication:
              <select v-model="med.item" @change="onMedSelect(med)">
                <option v-for="drug in drugDirectory" :key="drug.id" :value="drug">
                  {{ drug.name }} ({{ drug.strength }}) · {{ drug.form }} · Stock: {{ drug.stock }}
                </option>
              </select>
            </label>
            <label>Dosage Regimen:
              <select v-model="med.frequency">
                <option value="OD - Once daily">OD (Once daily)</option>
                <option value="BD - Twice daily (every 12h)">BD (Twice daily)</option>
                <option value="TDS - Three times daily (every 8h)">TDS (3x daily)</option>
                <option value="QID - Four times daily (every 6h)">QID (4x daily)</option>
                <option value="PRN - As needed for pain/symptoms">PRN (As needed)</option>
              </select>
            </label>
            <label>Duration:
              <div class="input-split">
                <input type="number" v-model.number="med.days" min="1" max="90" @input="calculateQty(med)" />
                <span class="unit-span">days</span>
              </div>
            </label>
          </div>

          <div class="form-grid-3 mt-2">
            <label>Clinical Directions for Patient:
              <input type="text" v-model="med.directions" placeholder="e.g. Take 1 tablet after meals with plenty of water" />
            </label>
            <label>Quantity to Dispense:
              <div class="input-split">
                <input type="number" v-model.number="med.dispenseQty" min="1" />
                <span class="unit-span">{{ med.form }}s</span>
              </div>
            </label>
            <label>Subtotal:
              <div class="price-calc-box">
                KES <strong>{{ ((med.item?.unitPriceMinor * med.dispenseQty) / 100).toFixed(2) }}</strong>
              </div>
            </label>
          </div>
        </div>

        <button type="button" class="btn-add-med-line" @click="addMedLine">
          ➕ Add Another Medication
        </button>
      </div>

      <!-- Action & Dispense to Cart -->
      <div class="intake-footer">
        <div class="total-summary">
          <span>Total Medications: <strong>{{ rxForm.medications.length }}</strong></span>
          <span>Estimated Total: <strong>KES {{ totalRxCostKes }}</strong></span>
        </div>
        <div class="footer-btns">
          <button type="button" class="btn-secondary" @click="saveDraft">Save Rx Record</button>
          <button type="button" class="btn-dispense" @click="dispenseToPosCart">
            💊 Dispense Directly to POS Cart
          </button>
        </div>
      </div>
    </div>

    <!-- Active Prescription Queue -->
    <div class="queue-section">
      <h4>Recent Prescriptions & Dispensing Log:</h4>
      <div class="rx-queue-grid">
        <div v-for="record in rxHistory" :key="record.id" class="rx-card">
          <div class="rx-card-header">
            <strong>Rx #{{ record.rxNumber }}</strong>
            <span class="rx-status-pill" :class="record.status">{{ record.status }}</span>
          </div>
          <div class="rx-patient">
            👤 {{ record.patientName }} ({{ record.patientAge }}y) · Dr. {{ record.doctorName }}
          </div>
          <div class="rx-meds-summary">
            <ul>
              <li v-for="(m, i) in record.meds" :key="i">
                {{ m.name }} · Qty: {{ m.qty }} ({{ m.directions }})
              </li>
            </ul>
          </div>
          <div class="rx-card-footer">
            <small>Dispensed: {{ record.date }}</small>
            <span class="rx-price">KES {{ record.totalKes }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const emit = defineEmits(['add-item', 'notification'])

const drugDirectory = [
  { id: 'd-1', name: 'Amoxicillin Capsules', strength: '500mg', form: 'Capsule', unitPriceMinor: 2500, stock: 450, generic: 'Amoxicillin', isControlled: false },
  { id: 'd-2', name: 'Augmentin Tablets', strength: '625mg', form: 'Tablet', unitPriceMinor: 12000, stock: 120, generic: 'Amoxicillin + Clavulanate', isControlled: false },
  { id: 'd-3', name: 'Panadol Extra Tablets', strength: '500mg/65mg', form: 'Tablet', unitPriceMinor: 1500, stock: 800, generic: 'Paracetamol + Caffeine', isControlled: false },
  { id: 'd-4', name: 'Azithromycin Tablets', strength: '500mg', form: 'Tablet', unitPriceMinor: 18000, stock: 65, generic: 'Azithromycin', isControlled: false },
  { id: 'd-5', name: 'Codeine Phosphate Tablets', strength: '30mg', form: 'Tablet', unitPriceMinor: 4500, stock: 40, generic: 'Codeine', isControlled: true },
  { id: 'd-6', name: 'Ventolin Inhaler', strength: '100mcg/dose', form: 'Inhaler', unitPriceMinor: 65000, stock: 28, generic: 'Salbutamol', isControlled: false },
  { id: 'd-7', name: 'Metformin HCl Tablets', strength: '500mg', form: 'Tablet', unitPriceMinor: 1200, stock: 600, generic: 'Metformin', isControlled: false },
]

const rxForm = ref({
  patientName: 'Sarah Mwangi',
  patientAge: 34,
  patientGender: 'Female',
  patientId: 'SHA-984210',
  doctorName: 'Dr. Arthur Muriuki',
  clinicName: 'Nairobi West Hospital',
  doctorRegNo: 'KMPDC-A4829',
  medications: [
    {
      item: drugDirectory[0],
      frequency: 'TDS - Three times daily (every 8h)',
      days: 5,
      dispenseQty: 15,
      directions: 'Take 1 capsule 3 times daily after meals for 5 days. Complete full course.',
    },
    {
      item: drugDirectory[2],
      frequency: 'PRN - As needed for pain/symptoms',
      days: 3,
      dispenseQty: 10,
      directions: 'Take 1-2 tablets every 6 hours as needed for fever or pain. Do not exceed 8 tabs/day.',
    },
  ],
})

const rxHistory = ref([
  {
    id: 'rx-rec-1',
    rxNumber: 'RX-2026-0891',
    patientName: 'David Kiprono',
    patientAge: 42,
    doctorName: 'Dr. K. Mutua',
    status: 'Dispensed',
    date: '10/11/2026 09:30',
    totalKes: '850.00',
    meds: [
      { name: 'Amoxicillin 500mg', qty: 21, directions: '1 cap TDS x 7 days' },
      { name: 'Panadol Extra', qty: 12, directions: '2 tabs PRN' },
    ],
  },
  {
    id: 'rx-rec-2',
    rxNumber: 'RX-2026-0890',
    patientName: 'Grace Achieng',
    patientAge: 28,
    doctorName: 'Dr. E. Omondi',
    status: 'Dispensed',
    date: '10/11/2026 08:45',
    totalKes: '2,400.00',
    meds: [
      { name: 'Augmentin 625mg', qty: 14, directions: '1 tab BD x 7 days' },
    ],
  },
])

const totalRxCostKes = computed(() => {
  let totalMinor = 0
  for (const med of rxForm.value.medications) {
    if (med.item && med.dispenseQty) {
      totalMinor += med.item.unitPriceMinor * med.dispenseQty
    }
  }
  return (totalMinor / 100).toFixed(2)
})

function onMedSelect(med) {
  calculateQty(med)
}

function calculateQty(med) {
  if (!med.days) return
  let dosesPerDay = 1
  if (med.frequency.startsWith('BD')) dosesPerDay = 2
  else if (med.frequency.startsWith('TDS')) dosesPerDay = 3
  else if (med.frequency.startsWith('QID')) dosesPerDay = 4
  med.dispenseQty = dosesPerDay * med.days
}

function addMedLine() {
  rxForm.value.medications.push({
    item: drugDirectory[1],
    frequency: 'BD - Twice daily (every 12h)',
    days: 7,
    dispenseQty: 14,
    directions: 'Take 1 tablet every 12 hours after meals.',
  })
}

function removeMed(index) {
  rxForm.value.medications.splice(index, 1)
}

function resetForm() {
  rxForm.value = {
    patientName: '',
    patientAge: '',
    patientGender: 'Female',
    patientId: '',
    doctorName: '',
    clinicName: '',
    doctorRegNo: '',
    medications: [
      {
        item: drugDirectory[0],
        frequency: 'TDS - Three times daily (every 8h)',
        days: 5,
        dispenseQty: 15,
        directions: 'Take after meals',
      },
    ],
  }
}

function saveDraft() {
  const rxNum = `RX-2026-${Math.floor(1000 + Math.random() * 9000)}`
  rxHistory.value.unshift({
    id: `rx-${Date.now()}`,
    rxNumber: rxNum,
    patientName: rxForm.value.patientName || 'Anonymous Patient',
    patientAge: rxForm.value.patientAge || 0,
    doctorName: rxForm.value.doctorName || 'Dr. On Duty',
    status: 'Logged',
    date: new Date().toLocaleString(),
    totalKes: totalRxCostKes.value,
    meds: rxForm.value.medications.map((m) => ({
      name: m.item.name,
      qty: m.dispenseQty,
      directions: m.directions,
    })),
  })
  emit('notification', { message: `Saved Prescription record ${rxNum}`, type: 'success' })
}

function dispenseToPosCart() {
  if (!rxForm.value.patientName || !rxForm.value.doctorName) {
    emit('notification', { message: 'Patient Name and Doctor Name are required before dispensing.', type: 'warning' })
    return
  }

  saveDraft()

  // Convert each med line into a cart item and push to cart
  const pendingItems = []
  for (const med of rxForm.value.medications) {
    const item = {
      id: `rx-med-${med.item.id}-${Date.now()}-${Math.random()}`,
      name: `${med.item.name} (${med.item.strength}) - Qty: ${med.dispenseQty}`,
      sku: `RX-${med.item.id.toUpperCase()}`,
      price_minor: med.item.unitPriceMinor * med.dispenseQty,
      unit_price_minor: med.item.unitPriceMinor,
      quantity: 1,
      kind: 'stocked',
      clinical_notes: `Patient: ${rxForm.value.patientName} · Directions: ${med.directions}`,
    }
    pendingItems.push(item)
    emit('add-item', item)
  }

  try {
    const existing = JSON.parse(localStorage.getItem('pos.retailCartPending') || '[]')
    existing.push(...pendingItems)
    localStorage.setItem('pos.retailCartPending', JSON.stringify(existing))
    window.dispatchEvent(new CustomEvent('pos:add-item-batch', { detail: pendingItems }))
  } catch (e) {
    console.error(e)
  }

  emit('notification', {
    message: `Dispensed ${rxForm.value.medications.length} prescribed medications into POS cart for ${rxForm.value.patientName}!`,
    type: 'success',
  })
}
</script>

<style scoped>
.prescription-station {
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
.btn-new-rx {
  background: #0284c7;
  color: white;
  border: none;
  padding: 0.55rem 1rem;
  border-radius: 8px;
  font-weight: 700;
  cursor: pointer;
}
.rx-intake-card {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
.form-section-title {
  font-size: 0.95rem;
  font-weight: 800;
  color: #0f172a;
  border-bottom: 2px solid #e2e8f0;
  padding-bottom: 0.35rem;
}
.form-grid-3 {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 0.85rem;
}
.form-grid-4 {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 0.85rem;
}
.col-span-2 {
  grid-column: span 2;
}
.rx-intake-card label {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  font-size: 0.85rem;
  font-weight: 600;
  color: #334155;
}
.rx-intake-card input, .rx-intake-card select {
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
.input-split input {
  flex: 1;
}
.unit-span {
  font-size: 0.85rem;
  color: #64748b;
  font-weight: 600;
}
.price-calc-box {
  padding: 0.5rem 0.75rem;
  background: #e0f2fe;
  border: 1px solid #bae6fd;
  border-radius: 6px;
  color: #0369a1;
  font-size: 0.95rem;
}
.medications-builder {
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
}
.med-line-card {
  background: white;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  padding: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
.med-line-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.med-badge {
  font-size: 0.75rem;
  font-weight: 800;
  background: #0284c7;
  color: white;
  padding: 0.15rem 0.5rem;
  border-radius: 4px;
}
.btn-remove-med {
  background: none;
  border: 1px solid #fca5a5;
  color: #b91c1c;
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: bold;
  cursor: pointer;
}
.btn-add-med-line {
  align-self: flex-start;
  background: #f1f5f9;
  border: 1px dashed #0284c7;
  color: #0284c7;
  padding: 0.5rem 1rem;
  border-radius: 6px;
  font-weight: 700;
  cursor: pointer;
}
.btn-add-med-line:hover {
  background: #e0f2fe;
}
.intake-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 1rem;
  padding-top: 1rem;
  border-top: 1px solid #cbd5e1;
}
.total-summary {
  display: flex;
  gap: 1.5rem;
  font-size: 1rem;
  color: #0f172a;
}
.footer-btns {
  display: flex;
  gap: 0.75rem;
}
.btn-secondary {
  background: white;
  border: 1px solid #cbd5e1;
  padding: 0.6rem 1rem;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
}
.btn-dispense {
  background: #166534;
  color: white;
  border: none;
  padding: 0.6rem 1.25rem;
  border-radius: 8px;
  font-weight: 700;
  cursor: pointer;
}
.queue-section h4 {
  margin: 0 0 0.85rem;
  font-size: 1rem;
}
.rx-queue-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 1rem;
}
.rx-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
.rx-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.9rem;
}
.rx-status-pill {
  font-size: 0.7rem;
  font-weight: 700;
  padding: 0.15rem 0.45rem;
  border-radius: 4px;
  background: #dcfce7;
  color: #166534;
}
.rx-patient {
  font-size: 0.85rem;
  color: #475569;
  font-weight: 600;
}
.rx-meds-summary ul {
  margin: 0;
  padding-left: 1.25rem;
  font-size: 0.8rem;
  color: #334155;
}
.rx-card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-top: 1px dashed #e2e8f0;
  padding-top: 0.5rem;
  margin-top: 0.25rem;
}
.rx-price {
  font-weight: 800;
  color: #0284c7;
}
@media (max-width: 768px) {
  .col-span-2 {
    grid-column: span 1;
  }
}
</style>
