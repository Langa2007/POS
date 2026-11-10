<template>
  <div class="cut-to-length-manager">
    <div class="ctl-header">
      <div>
        <h3>
          <PosIcon name="scissors" size="20" class="header-icon" />
          Cut-to-Length Station &amp; Offcut Tracking
        </h3>
        <p class="subtitle">Sell fractional lengths of pipes, cables, steel rebars, and timber with automated remnant tracking</p>
      </div>
    </div>

    <div class="ctl-layout">
      <!-- Cut Config Form -->
      <div class="cutting-form-card">
        <h4>Material to Cut:</h4>
        <div class="form-row">
          <label>Select Material Stock:</label>
          <select v-model="selectedMaterialId" class="form-select" @change="onMaterialChange">
            <option v-for="item in cutMaterials" :key="item.id" :value="item.id">
              {{ item.name }} &mdash; Full Length: {{ item.standardLengthM }}m (Stock: {{ item.stockPieces }} pcs)
            </option>
          </select>
        </div>

        <div class="material-stats-strip">
          <div class="m-stat">
            <span class="m-lbl">Full Standard Length:</span>
            <strong>{{ currentMaterial.standardLengthM }} Metres</strong>
          </div>
          <div class="m-stat">
            <span class="m-lbl">Full Piece Price:</span>
            <strong>KES {{ (currentMaterial.fullPriceMinor / 100).toLocaleString() }}</strong>
          </div>
          <div class="m-stat">
            <span class="m-lbl">Rate per Metre:</span>
            <strong>KES {{ (ratePerMetreMinor / 100).toFixed(2) }}</strong>
          </div>
        </div>

        <div class="form-grid">
          <div class="form-row">
            <label>Length Required (Metres):</label>
            <div class="input-with-suffix">
              <input
                type="number"
                v-model.number="cutLengthM"
                :min="0.1"
                :max="currentMaterial.standardLengthM"
                step="0.05"
                class="form-input"
              />
              <span class="suffix">Metres</span>
            </div>
            <small class="hint">Max cut from single piece is {{ currentMaterial.standardLengthM }}m</small>
          </div>

          <div class="form-row">
            <label>Number of Cuts / Pieces:</label>
            <input
              type="number"
              v-model.number="cutQuantity"
              :min="1"
              step="1"
              class="form-input"
            />
          </div>
        </div>

        <div class="cutting-fee-row">
          <label class="check-label">
            <input type="checkbox" v-model="includeCutFee" />
            <span>Include Workshop Precision Cutting Fee (KES 50.00 / cut)</span>
          </label>
        </div>

        <!-- Remnant & Cost Calculation Banner -->
        <div class="cut-calculation-results">
          <div class="res-box">
            <span class="res-lbl">Customer Charged:</span>
            <strong class="res-val highlight">KES {{ (totalCutPriceMinor / 100).toLocaleString() }}</strong>
            <small>{{ cutQuantity }} x {{ cutLengthM }}m @ KES {{ (ratePerMetreMinor / 100).toFixed(2) }}/m</small>
          </div>

          <div class="res-box remnant">
            <span class="res-lbl">Remaining Remnant / Offcut:</span>
            <strong class="res-val">{{ remnantLengthM.toFixed(2) }} Metres</strong>
            <small>Automatically logged to Offcut Bin #{{ currentMaterial.binCode }}</small>
          </div>
        </div>

        <button type="button" class="btn-primary" :disabled="cutLengthM <= 0 || cutLengthM > currentMaterial.standardLengthM" @click="addCutOrderToCart">
          <PosIcon name="plus" size="16" />
          Add Cut Order to POS Cart
        </button>
      </div>

      <!-- Offcuts Bin & Remnants Inventory -->
      <div class="offcuts-bin-card">
        <div class="bin-top">
          <h4>
            <PosIcon name="package" size="16" />
            Active Offcuts &amp; Remnants Bin
          </h4>
          <span class="bin-badge">{{ offcuts.length }} Remnants Available</span>
        </div>
        <p class="bin-desc">Discounted pre-cut offcuts available for immediate purchase without cutting new stock:</p>

        <div class="offcuts-list">
          <div
            v-for="rem in offcuts"
            :key="rem.id"
            class="offcut-item"
          >
            <div class="offcut-info">
              <span class="offcut-name">{{ rem.name }}</span>
              <span class="offcut-dims">{{ rem.lengthM }}m remnant &bull; Bin {{ rem.bin }}</span>
            </div>
            <div class="offcut-action">
              <strong class="offcut-price">KES {{ (rem.priceMinor / 100).toLocaleString() }}</strong>
              <button type="button" class="btn-buy-offcut" @click="buyOffcut(rem)">
                <PosIcon name="shopping-cart" size="14" />
                Buy Offcut
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { PosIcon } from '@pos/ui'

const emit = defineEmits(['add-item', 'notification'])

const cutMaterials = [
  {
    id: 'cm-1',
    name: 'PVC Drainage Pipe 110mm Class B',
    sku: 'PVC-110-6M',
    standardLengthM: 6.0,
    fullPriceMinor: 160000, // KES 1,600
    stockPieces: 45,
    binCode: 'BIN-PLUMB-04',
  },
  {
    id: 'cm-2',
    name: 'SWA Armoured Underground Cable 3-Core 4mm',
    sku: 'CAB-SWA-4MM',
    standardLengthM: 100.0,
    fullPriceMinor: 3500000, // KES 35,000 (350/m)
    stockPieces: 3,
    binCode: 'BIN-ELEC-01',
  },
  {
    id: 'cm-3',
    name: 'Mahogany Hardwood Beam 4" x 2"',
    sku: 'TIM-MAH-4X2',
    standardLengthM: 5.4,
    fullPriceMinor: 220000, // KES 2,200 (407.40/m)
    stockPieces: 28,
    binCode: 'BIN-TIM-08',
  },
  {
    id: 'cm-4',
    name: 'High-Yield Ribbed Rebar Steel T12',
    sku: 'STL-REB-T12',
    standardLengthM: 12.0,
    fullPriceMinor: 140000, // KES 1,400 (116.67/m)
    stockPieces: 120,
    binCode: 'BIN-STEEL-02',
  },
]

const selectedMaterialId = ref(cutMaterials[0].id)
const cutLengthM = ref(3.8)
const cutQuantity = ref(1)
const includeCutFee = ref(true)

const currentMaterial = computed(() => {
  return cutMaterials.find(m => m.id === selectedMaterialId.value) || cutMaterials[0]
})

const ratePerMetreMinor = computed(() => {
  return Math.round(currentMaterial.value.fullPriceMinor / currentMaterial.value.standardLengthM)
})

const remnantLengthM = computed(() => {
  const rem = currentMaterial.value.standardLengthM - cutLengthM.value
  return Math.max(0, rem)
})

const totalCutPriceMinor = computed(() => {
  const materialCost = Math.round(ratePerMetreMinor.value * cutLengthM.value * cutQuantity.value)
  const cuttingFee = includeCutFee.value ? 5000 * cutQuantity.value : 0 // KES 50 fee
  return materialCost + cuttingFee
})

const offcuts = ref([
  { id: 'off-1', name: 'PVC Drainage Pipe 110mm', lengthM: 2.2, priceMinor: 48000, bin: 'PLUMB-04' },
  { id: 'off-2', name: 'SWA Armoured Cable 4mm', lengthM: 14.5, priceMinor: 420000, bin: 'ELEC-01' },
  { id: 'off-3', name: 'Mahogany Timber 4" x 2"', lengthM: 1.6, priceMinor: 55000, bin: 'TIM-08' },
  { id: 'off-4', name: 'Ribbed Rebar Steel T12', lengthM: 4.8, priceMinor: 45000, bin: 'STEEL-02' },
])

function onMaterialChange() {
  cutLengthM.value = Number((currentMaterial.value.standardLengthM * 0.6).toFixed(2))
}

function addCutOrderToCart() {
  const cartItem = {
    id: `cut-${currentMaterial.value.id}-${Date.now()}`,
    name: `${currentMaterial.value.name} [Cut: ${cutQuantity.value}x ${cutLengthM.value}m]`,
    sku: `CUT-${currentMaterial.value.sku}`,
    price_minor: totalCutPriceMinor.value,
    unit_price_minor: totalCutPriceMinor.value,
    quantity: 1,
    cut_length_m: cutLengthM.value,
    remnant_m: remnantLengthM.value,
    kind: 'variable',
  }

  // Also push remnant to offcut bin
  if (remnantLengthM.value > 0.5) {
    offcuts.value.unshift({
      id: `off-${Date.now()}`,
      name: currentMaterial.value.name,
      lengthM: Number(remnantLengthM.value.toFixed(2)),
      priceMinor: Math.round(ratePerMetreMinor.value * remnantLengthM.value * 0.85), // 15% discount for offcut
      bin: currentMaterial.value.binCode.replace('BIN-', ''),
    })
  }

  try {
    const existing = JSON.parse(localStorage.getItem('pos.retailCartPending') || '[]')
    localStorage.setItem('pos.retailCartPending', JSON.stringify([...existing, cartItem]))
  } catch (e) {}

  window.dispatchEvent(new CustomEvent('pos:add-item', { detail: cartItem }))
  emit('add-item', cartItem)

  emit('notification', {
    message: `Added precision cut (${cutLengthM.value}m · KES ${(totalCutPriceMinor.value / 100).toLocaleString()}) to cart. Remnant ${remnantLengthM.value.toFixed(2)}m logged to bin.`,
    type: 'success',
  })
}

function buyOffcut(rem) {
  const cartItem = {
    id: `offcut-${rem.id}-${Date.now()}`,
    name: `[Offcut Bin ${rem.bin}] ${rem.name} (${rem.lengthM}m)`,
    sku: `OFF-${rem.id}`,
    price_minor: rem.priceMinor,
    unit_price_minor: rem.priceMinor,
    quantity: 1,
    kind: 'stocked',
    stock_on_hand: 1,
  }

  try {
    const existing = JSON.parse(localStorage.getItem('pos.retailCartPending') || '[]')
    localStorage.setItem('pos.retailCartPending', JSON.stringify([...existing, cartItem]))
  } catch (e) {}

  window.dispatchEvent(new CustomEvent('pos:add-item', { detail: cartItem }))
  emit('add-item', cartItem)

  offcuts.value = offcuts.value.filter(o => o.id !== rem.id)

  emit('notification', {
    message: `Added remnant offcut (${rem.lengthM}m · KES ${(rem.priceMinor / 100).toLocaleString()}) to cart`,
    type: 'success',
  })
}
</script>

<style scoped>
.cut-to-length-manager {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.ctl-header h3 {
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

.ctl-layout {
  display: grid;
  grid-template-columns: minmax(320px, 1.4fr) minmax(280px, 1fr);
  gap: 18px;
  align-items: start;
}

.cutting-form-card, .offcuts-bin-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.cutting-form-card h4, .bin-top h4 {
  margin: 0;
  font-size: 15px;
  color: #0f172a;
  display: flex;
  align-items: center;
  gap: 6px;
}

.form-row {
  display: flex;
  flex-direction: column;
  gap: 6px;
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

.material-stats-strip {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
  background: #f8fafc;
  padding: 10px 12px;
  border-radius: 8px;
}

.m-stat {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.m-lbl {
  font-size: 10px;
  color: #64748b;
  text-transform: uppercase;
}

.m-stat strong {
  font-size: 13px;
  color: #0f172a;
}

.form-grid {
  display: grid;
  grid-template-columns: 1.4fr 1fr;
  gap: 12px;
}

.input-with-suffix {
  display: flex;
  align-items: center;
  gap: 6px;
}

.input-with-suffix input {
  flex: 1;
}

.suffix {
  font-size: 12px;
  font-weight: 600;
  color: #64748b;
}

.hint {
  font-size: 11px;
  color: #94a3b8;
}

.cutting-fee-row {
  font-size: 12px;
  color: #334155;
}

.check-label {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
}

.cut-calculation-results {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  margin-top: 4px;
}

.res-box {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.res-box.remnant {
  background: #fefce8;
  border-color: #fef08a;
}

.res-lbl {
  font-size: 11px;
  color: #64748b;
  text-transform: uppercase;
  font-weight: 700;
}

.res-val {
  font-size: 16px;
  color: #0f172a;
}

.res-val.highlight {
  color: #0284c7;
}

.res-box small {
  font-size: 11px;
  color: #64748b;
}

.btn-primary {
  background: #0284c7;
  color: white;
  border: none;
  padding: 10px 16px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

.btn-primary:hover {
  background: #0369a1;
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Offcuts */
.bin-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.bin-badge {
  background: #e0f2fe;
  color: #0369a1;
  font-size: 11px;
  font-weight: 700;
  padding: 2px 8px;
  border-radius: 10px;
}

.bin-desc {
  font-size: 12px;
  color: #64748b;
  margin: 0;
}

.offcuts-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.offcut-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  background: #fbfcfd;
}

.offcut-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.offcut-name {
  font-size: 13px;
  font-weight: 600;
  color: #1e293b;
}

.offcut-dims {
  font-size: 11px;
  color: #0284c7;
}

.offcut-action {
  display: flex;
  align-items: center;
  gap: 10px;
}

.offcut-price {
  font-size: 13px;
  color: #0f172a;
}

.btn-buy-offcut {
  padding: 6px 10px;
  border: 1px solid #cbd5e1;
  background: white;
  border-radius: 6px;
  font-size: 11px;
  font-weight: 600;
  color: #334155;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.btn-buy-offcut:hover {
  background: #f1f5f9;
}

@media (max-width: 768px) {
  .ctl-layout {
    grid-template-columns: 1fr;
  }
}
</style>
