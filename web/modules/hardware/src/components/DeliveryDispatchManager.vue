<template>
  <div class="delivery-dispatch-manager">
    <div class="dispatch-header">
      <div>
        <h3>
          <PosIcon name="truck" size="20" class="header-icon" />
          Heavy Site Delivery &amp; Dispatch Manifest
        </h3>
        <p class="subtitle">Vehicle capacity planning, site delivery scheduling, payload weight checks, and gate passes</p>
      </div>

      <button type="button" class="btn-create-dispatch" @click="openCreateModal = true">
        <PosIcon name="plus" size="16" />
        Schedule Site Dispatch
      </button>
    </div>

    <div class="dispatch-layout">
      <!-- Active Dispatches List -->
      <div class="dispatches-list-col">
        <h4>Scheduled &amp; In-Transit Dispatches:</h4>
        <div class="dispatches-grid">
          <div
            v-for="d in dispatches"
            :key="d.id"
            class="dispatch-card"
            :class="d.status"
            @click="selectedDispatch = d"
          >
            <div class="disp-top">
              <span class="disp-num">{{ d.dispatchNo }}</span>
              <span class="status-pill" :class="d.status">{{ formatStatus(d.status) }}</span>
            </div>

            <div class="disp-site">
              <strong>{{ d.siteName }}</strong>
              <small>{{ d.address }}</small>
            </div>

            <div class="disp-vehicle">
              <PosIcon name="truck" size="14" />
              <span>{{ d.vehicleName }} (Reg: {{ d.vehicleReg }})</span>
            </div>

            <div class="payload-bar-wrap">
              <div class="p-labels">
                <span>Payload: {{ d.totalWeightKg }} kg</span>
                <span>Max: {{ d.maxPayloadKg }} kg</span>
              </div>
              <div class="p-track">
                <div
                  class="p-fill"
                  :style="{ width: `${Math.min(100, Math.round((d.totalWeightKg / d.maxPayloadKg) * 100))}%` }"
                ></div>
              </div>
            </div>

            <div class="disp-foot">
              <span>Driver: {{ d.driverName }}</span>
              <button type="button" class="btn-manifest" @click.stop="printManifest(d)">
                <PosIcon name="printer" size="13" />
                Gate Pass
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Live Payload Estimator -->
      <div class="estimator-col">
        <div class="estimator-card">
          <h4>
            <PosIcon name="scale" size="16" />
            Truck Payload &amp; Delivery Fee Estimator
          </h4>

          <div class="form-row">
            <label>Selected Vehicle Fleet:</label>
            <select v-model="calcVehicle" class="form-select">
              <option v-for="v in vehicles" :key="v.id" :value="v">
                {{ v.name }} (Max: {{ v.maxPayloadKg }} kg / {{ v.tonnage }}T)
              </option>
            </select>
          </div>

          <div class="form-row">
            <label>Destination Zone:</label>
            <select v-model="calcZone" class="form-select">
              <option v-for="z in zones" :key="z.id" :value="z">
                {{ z.name }} &mdash; KES {{ (z.feeMinor / 100).toLocaleString() }} ({{ z.distance }})
              </option>
            </select>
          </div>

          <div class="form-row">
            <label>Estimated Cargo Weight (kg):</label>
            <input type="number" v-model.number="calcWeightKg" :min="100" step="100" class="form-input" />
          </div>

          <!-- Weight Safety Gauge -->
          <div class="weight-check-box" :class="{ 'overload': calcWeightKg > calcVehicle.maxPayloadKg }">
            <div class="wc-header">
              <span class="wc-title">{{ calcWeightKg > calcVehicle.maxPayloadKg ? 'OVERLOAD WARNING' : 'Payload Status' }}</span>
              <span class="wc-ratio">{{ Math.round((calcWeightKg / calcVehicle.maxPayloadKg) * 100) }}% Capacity</span>
            </div>
            <p v-if="calcWeightKg > calcVehicle.maxPayloadKg" class="wc-alert">
              Selected cargo ({{ calcWeightKg }}kg) exceeds {{ calcVehicle.name }} safe limit ({{ calcVehicle.maxPayloadKg }}kg). Please upgrade to a larger truck.
            </p>
            <p v-else class="wc-safe">
              Safe load within axle limits. Vehicle can carry this shipment safely.
            </p>
          </div>

          <div class="calc-footer">
            <div class="fee-preview">
              <span class="f-lbl">Calculated Site Delivery Fee:</span>
              <strong class="f-val">KES {{ (calcZone.feeMinor / 100).toLocaleString() }}</strong>
            </div>
            <button type="button" class="btn-add-fee" :disabled="calcWeightKg > calcVehicle.maxPayloadKg" @click="addDeliveryFeeToCart">
              <PosIcon name="plus" size="16" />
              Add Delivery Fee to Cart
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { PosIcon } from '@pos/ui'

const emit = defineEmits(['add-item', 'notification'])

const openCreateModal = ref(false)
const selectedDispatch = ref(null)

const vehicles = [
  { id: 'v-1', name: 'Isuzu D-Max 1-Tonne Pickup', tonnage: 1, maxPayloadKg: 1000 },
  { id: 'v-2', name: 'Mitsubishi Canter 3-Tonne Truck', tonnage: 3, maxPayloadKg: 3000 },
  { id: 'v-3', name: 'Mercedes Actros 10-Tonne Tipper', tonnage: 10, maxPayloadKg: 10000 },
]

const zones = [
  { id: 'z-1', name: 'Zone 1: Local Municipality (Under 10 km)', feeMinor: 150000, distance: '< 10 km' },
  { id: 'z-2', name: 'Zone 2: County Suburbs (10 - 30 km)', feeMinor: 350000, distance: '10 - 30 km' },
  { id: 'z-3', name: 'Zone 3: Remote Upcountry Sites (30+ km)', feeMinor: 750000, distance: '30+ km' },
]

const calcVehicle = ref(vehicles[1])
const calcZone = ref(zones[0])
const calcWeightKg = ref(2200)

const dispatches = ref([
  {
    id: 'dsp-1',
    dispatchNo: 'DSP-2026-091',
    siteName: 'Kilimani Heights Apartment Site',
    address: 'Plot 41B, Argwings Kodhek Rd',
    vehicleName: 'Mitsubishi Canter 3-Tonne',
    vehicleReg: 'KBZ 894M',
    driverName: 'Peter Kamau',
    totalWeightKg: 2450,
    maxPayloadKg: 3000,
    status: 'in-transit',
  },
  {
    id: 'dsp-2',
    dispatchNo: 'DSP-2026-092',
    siteName: 'Riverside Villa Construction',
    address: 'Gate 4, Riverside Drive',
    vehicleName: 'Isuzu D-Max 1-Tonne',
    vehicleReg: 'KDA 120P',
    driverName: 'Daniel Otiato',
    totalWeightKg: 850,
    maxPayloadKg: 1000,
    status: 'scheduled',
  },
  {
    id: 'dsp-3',
    dispatchNo: 'DSP-2026-093',
    siteName: 'Coast Commercial Warehouse',
    address: 'Mombasa Road Km 18',
    vehicleName: 'Mercedes Actros 10-Tonne',
    vehicleReg: 'KCR 419G',
    driverName: 'Hassan Ali',
    totalWeightKg: 9600,
    maxPayloadKg: 10000,
    status: 'loading',
  },
])

function formatStatus(status) {
  switch (status) {
    case 'scheduled': return 'Scheduled'
    case 'loading': return 'Loading at Yard'
    case 'in-transit': return 'On The Road'
    case 'delivered': return 'Delivered'
    default: return status
  }
}

function printManifest(d) {
  emit('notification', {
    message: `Delivery Gate Pass & Weighbridge Slip printed for ${d.dispatchNo} (${d.vehicleReg})`,
    type: 'info',
  })
}

function addDeliveryFeeToCart() {
  const cartItem = {
    id: `del-fee-${Date.now()}`,
    name: `Site Delivery Service [${calcZone.value.name}]`,
    sku: `SRV-DEL-${calcZone.value.id}`,
    price_minor: calcZone.value.feeMinor,
    unit_price_minor: calcZone.value.feeMinor,
    quantity: 1,
    kind: 'service',
  }

  try {
    const existing = JSON.parse(localStorage.getItem('pos.retailCartPending') || '[]')
    localStorage.setItem('pos.retailCartPending', JSON.stringify([...existing, cartItem]))
  } catch (e) {}

  window.dispatchEvent(new CustomEvent('pos:add-item', { detail: cartItem }))
  emit('add-item', cartItem)

  emit('notification', {
    message: `Added Site Delivery Fee (KES ${(calcZone.value.feeMinor / 100).toLocaleString()}) to cart`,
    type: 'success',
  })
}
</script>

<style scoped>
.delivery-dispatch-manager {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.dispatch-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  flex-wrap: wrap;
  gap: 12px;
}

.dispatch-header h3 {
  margin: 0;
  font-size: 18px;
  color: #1e293b;
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-icon {
  color: #0284c7;
}

.subtitle {
  margin: 4px 0 0;
  font-size: 13px;
  color: #64748b;
}

.btn-create-dispatch {
  background: #0284c7;
  color: white;
  border: none;
  padding: 8px 14px;
  border-radius: 8px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.btn-create-dispatch:hover {
  background: #0369a1;
}

.dispatch-layout {
  display: grid;
  grid-template-columns: minmax(320px, 1.4fr) minmax(280px, 1fr);
  gap: 18px;
  align-items: start;
}

.dispatches-list-col h4, .estimator-card h4 {
  margin: 0 0 12px;
  font-size: 14px;
  color: #475569;
  display: flex;
  align-items: center;
  gap: 6px;
}

.dispatches-grid {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.dispatch-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.dispatch-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
}

.disp-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.disp-num {
  font-weight: 700;
  font-size: 14px;
  color: #0f172a;
}

.status-pill {
  font-size: 10px;
  font-weight: 700;
  padding: 2px 8px;
  border-radius: 6px;
  text-transform: uppercase;
}

.status-pill.scheduled { background: #e0f2fe; color: #0369a1; }
.status-pill.loading { background: #fef3c7; color: #b45309; }
.status-pill.in-transit { background: #dcfce7; color: #15803d; }

.disp-site {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.disp-site strong {
  font-size: 13px;
  color: #1e293b;
}

.disp-site small {
  font-size: 11px;
  color: #64748b;
}

.disp-vehicle {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #475569;
}

.payload-bar-wrap {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-top: 4px;
}

.p-labels {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: #64748b;
}

.p-track {
  height: 6px;
  background: #e2e8f0;
  border-radius: 3px;
  overflow: hidden;
}

.p-fill {
  height: 100%;
  background: #0284c7;
  border-radius: 3px;
}

.disp-foot {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-top: 1px solid #f1f5f9;
  padding-top: 8px;
  font-size: 11px;
  color: #64748b;
}

.btn-manifest {
  padding: 4px 8px;
  border: 1px solid #cbd5e1;
  background: white;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  color: #334155;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.btn-manifest:hover { background: #f1f5f9; }

/* Estimator */
.estimator-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 18px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.form-row {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.form-row label {
  font-size: 12px;
  font-weight: 600;
  color: #475569;
}

.form-select, .form-input {
  padding: 8px 10px;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 13px;
}

.weight-check-box {
  background: #f0fdf4;
  border: 1px solid #bbf7d0;
  border-radius: 8px;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.weight-check-box.overload {
  background: #fef2f2;
  border-color: #fca5a5;
}

.wc-header {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
}

.weight-check-box.overload .wc-title { color: #dc2626; }
.weight-check-box:not(.overload) .wc-title { color: #166534; }

.wc-safe {
  font-size: 12px;
  color: #166534;
  margin: 0;
}

.wc-alert {
  font-size: 12px;
  color: #b91c1c;
  margin: 0;
}

.calc-footer {
  margin-top: 6px;
  padding-top: 12px;
  border-top: 2px solid #e2e8f0;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.fee-preview {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
}

.f-lbl {
  font-size: 11px;
  color: #64748b;
}

.f-val {
  font-size: 18px;
  color: #0f172a;
}

.btn-add-fee {
  width: 100%;
  padding: 10px;
  border-radius: 8px;
  background: #0284c7;
  color: white;
  border: none;
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

.btn-add-fee:hover { background: #0369a1; }
.btn-add-fee:disabled { opacity: 0.5; cursor: not-allowed; }

@media (max-width: 768px) {
  .dispatch-layout {
    grid-template-columns: 1fr;
  }
}
</style>
