<template>
  <div class="multi-uom-converter">
    <div class="converter-header">
      <div>
        <h3>
          <PosIcon name="layers" size="20" class="header-icon" />
          Multi-Unit Packaging &amp; Bulk Conversions
        </h3>
        <p class="subtitle">Convert building supplies between bags, pallets, bundles, and running metres with bulk pricing</p>
      </div>
    </div>

    <div class="converter-layout">
      <!-- Item Selection List -->
      <div class="materials-list-col">
        <h4>Select Material / Hardware Supply:</h4>
        <div class="materials-list">
          <div
            v-for="mat in materials"
            :key="mat.id"
            class="material-card"
            :class="{ 'active': selectedMaterial?.id === mat.id }"
            @click="selectMaterial(mat)"
          >
            <div class="mat-name">{{ mat.name }}</div>
            <div class="mat-sku">{{ mat.sku }} &bull; {{ mat.category }}</div>
            <div class="mat-base-unit">Base Unit: <strong>{{ mat.baseUnit }}</strong></div>
          </div>
        </div>
      </div>

      <!-- Unit Conversion & Pricing Box -->
      <div class="conversion-engine-col" v-if="selectedMaterial">
        <div class="engine-card">
          <div class="engine-card-top">
            <div>
              <h4>{{ selectedMaterial.name }}</h4>
              <span class="category-pill">{{ selectedMaterial.category }}</span>
            </div>
            <span class="stock-pill">In Stock: {{ selectedMaterial.stockBase }} {{ selectedMaterial.baseUnit }}s</span>
          </div>

          <!-- Unit Selector & Quantity Input -->
          <div class="uom-selection-group">
            <label class="lbl">Select Ordering Unit of Measure:</label>
            <div class="unit-pills">
              <button
                v-for="uom in selectedMaterial.uoms"
                :key="uom.name"
                type="button"
                class="uom-pill"
                :class="{ 'active': activeUom.name === uom.name }"
                @click="setUom(uom)"
              >
                <span class="uom-name">{{ uom.name }}</span>
                <span class="uom-ratio">= {{ uom.factor }} {{ selectedMaterial.baseUnit }}s</span>
              </button>
            </div>
          </div>

          <div class="quantity-input-row">
            <div class="qty-field">
              <label>Enter Quantity (in {{ activeUom.name }}s):</label>
              <div class="input-wrap">
                <input
                  type="number"
                  v-model.number="orderQty"
                  :min="1"
                  step="1"
                  class="qty-input"
                />
                <span class="unit-suffix">{{ activeUom.name }}s</span>
              </div>
            </div>

            <div class="equivalent-preview">
              <span class="eq-lbl">Equivalent Total:</span>
              <strong class="eq-val">{{ totalBaseUnits.toLocaleString() }} {{ selectedMaterial.baseUnit }}s</strong>
              <small class="bulk-save" v-if="savingsMinor > 0">Bulk Savings: KES {{ (savingsMinor / 100).toLocaleString() }}</small>
            </div>
          </div>

          <!-- Price & Packaging Breakdown Table -->
          <div class="pricing-breakdown-table">
            <div class="t-row t-head">
              <span>Unit Tier</span>
              <span>Conversion Factor</span>
              <span>Unit Price</span>
              <span>Effective / {{ selectedMaterial.baseUnit }}</span>
            </div>
            <div
              v-for="uom in selectedMaterial.uoms"
              :key="uom.name"
              class="t-row"
              :class="{ 'highlight': activeUom.name === uom.name }"
            >
              <span class="uom-col">{{ uom.name }}</span>
              <span>1 {{ uom.name }} = {{ uom.factor }} {{ selectedMaterial.baseUnit }}s</span>
              <span class="price-col">KES {{ (uom.priceMinor / 100).toLocaleString() }}</span>
              <span class="eff-col">KES {{ ((uom.priceMinor / uom.factor) / 100).toFixed(2) }}</span>
            </div>
          </div>

          <!-- Add to Till Action -->
          <div class="action-footer">
            <div class="total-cost-box">
              <span class="total-lbl">Total Cost ({{ orderQty }} {{ activeUom.name }}s):</span>
              <strong class="total-val">KES {{ (calculatedTotalMinor / 100).toLocaleString() }}</strong>
            </div>

            <button type="button" class="btn-add-cart" @click="addToCartAction">
              <PosIcon name="plus" size="16" />
              Add to Active POS Cart
            </button>
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

const materials = [
  {
    id: 'mat-1',
    sku: 'CEM-BAM-325',
    name: 'Bamburi Portland Cement 32.5R',
    category: 'Masonry & Cement',
    baseUnit: 'Bag',
    stockBase: 480,
    uoms: [
      { name: 'Single Bag (50kg)', factor: 1, priceMinor: 78000 },
      { name: 'Pallet (20 Bags / 1-Tonne)', factor: 20, priceMinor: 1480000 }, // KES 14,800 vs 15,600
      { name: 'Lorry Load (40 Bags / 2-Tonne)', factor: 40, priceMinor: 2850000 }, // KES 28,500 vs 31,200
    ],
  },
  {
    id: 'mat-2',
    sku: 'TIM-CYP-2X2',
    name: 'Treated Cypress Timber 2" x 2"',
    category: 'Timber & Boards',
    baseUnit: 'Metre',
    stockBase: 1200,
    uoms: [
      { name: 'Running Metre', factor: 1, priceMinor: 9500 },
      { name: 'Standard Piece (12 Metres)', factor: 12, priceMinor: 108000 }, // KES 1,080 vs 1,140
      { name: 'Trade Bundle (10 Pieces / 120m)', factor: 120, priceMinor: 1020000 }, // KES 10,200 vs 11,400
    ],
  },
  {
    id: 'mat-3',
    sku: 'SCR-GYP-25',
    name: 'Black Phosphated Gypsum Screws 25mm',
    category: 'Fixings & Fasteners',
    baseUnit: 'Piece',
    stockBase: 25000,
    uoms: [
      { name: 'Single Screw', factor: 1, priceMinor: 200 },
      { name: 'Box (100 Pieces)', factor: 100, priceMinor: 18000 }, // KES 180 vs 200
      { name: 'Trade Bucket (1,000 Pieces)', factor: 1000, priceMinor: 150000 }, // KES 1,500 vs 2,000
    ],
  },
  {
    id: 'mat-4',
    sku: 'PNT-CRW-VIN',
    name: 'Crown Vinyl Matt Emulsion (White)',
    category: 'Paints & Coatings',
    baseUnit: 'Litre',
    stockBase: 350,
    uoms: [
      { name: '1-Litre Tin', factor: 1, priceMinor: 85000 },
      { name: '4-Litre Gallon', factor: 4, priceMinor: 280000 }, // KES 2,800 vs 3,400
      { name: '20-Litre Drum', factor: 20, priceMinor: 1250000 }, // KES 12,500 vs 17,000
    ],
  },
]

const selectedMaterial = ref(materials[0])
const activeUom = ref(materials[0].uoms[1]) // Default to Pallet
const orderQty = ref(2)

function selectMaterial(mat) {
  selectedMaterial.value = mat
  activeUom.value = mat.uoms[0]
  orderQty.value = 1
}

function setUom(uom) {
  activeUom.value = uom
}

const totalBaseUnits = computed(() => {
  if (!activeUom.value || !orderQty.value) return 0
  return activeUom.value.factor * orderQty.value
})

const calculatedTotalMinor = computed(() => {
  if (!activeUom.value || !orderQty.value) return 0
  return activeUom.value.priceMinor * orderQty.value
})

const savingsMinor = computed(() => {
  if (!selectedMaterial.value || !activeUom.value) return 0
  const basePrice = selectedMaterial.value.uoms[0].priceMinor
  const costAtBasePrice = basePrice * totalBaseUnits.value
  return Math.max(0, costAtBasePrice - calculatedTotalMinor.value)
})

function addToCartAction() {
  if (!selectedMaterial.value || !orderQty.value) return

  const cartItem = {
    id: `mat-${selectedMaterial.value.id}-${Date.now()}`,
    name: `${selectedMaterial.value.name} [${orderQty.value} x ${activeUom.value.name}]`,
    sku: `${selectedMaterial.value.sku}-${activeUom.value.name.slice(0, 3).toUpperCase()}`,
    price_minor: calculatedTotalMinor.value,
    unit_price_minor: activeUom.value.priceMinor,
    quantity: orderQty.value,
    uom: activeUom.value.name,
    base_quantity: totalBaseUnits.value,
    kind: 'stocked',
    stock_on_hand: Math.floor(selectedMaterial.value.stockBase / activeUom.value.factor),
  }

  // Pending queue
  try {
    const existing = JSON.parse(localStorage.getItem('pos.retailCartPending') || '[]')
    localStorage.setItem('pos.retailCartPending', JSON.stringify([...existing, cartItem]))
  } catch (e) {}

  window.dispatchEvent(new CustomEvent('pos:add-item', { detail: cartItem }))
  emit('add-item', cartItem)

  emit('notification', {
    message: `Added ${orderQty.value}x ${activeUom.value.name} of ${selectedMaterial.value.name} (KES ${(calculatedTotalMinor.value / 100).toLocaleString()}) to POS Cart`,
    type: 'success',
  })
}
</script>

<style scoped>
.multi-uom-converter {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.converter-header h3 {
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

.converter-layout {
  display: grid;
  grid-template-columns: minmax(240px, 1fr) minmax(320px, 1.8fr);
  gap: 18px;
  align-items: start;
}

.materials-list-col h4 {
  margin: 0 0 10px;
  font-size: 14px;
  color: #475569;
}

.materials-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.material-card {
  background: white;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  padding: 12px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.material-card:hover {
  background: #f8fafc;
  border-color: #94a3b8;
}

.material-card.active {
  border-color: #0284c7;
  background: #f0f9ff;
  box-shadow: 0 0 0 2px rgba(2, 132, 199, 0.2);
}

.mat-name {
  font-weight: 700;
  font-size: 14px;
  color: #1e293b;
}

.mat-sku {
  font-size: 11px;
  color: #64748b;
  margin-top: 2px;
}

.mat-base-unit {
  font-size: 12px;
  color: #0369a1;
  margin-top: 6px;
}

/* Engine Card */
.engine-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.engine-card-top {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  border-bottom: 1px solid #f1f5f9;
  padding-bottom: 12px;
}

.engine-card-top h4 {
  margin: 0 0 4px;
  font-size: 16px;
  color: #0f172a;
}

.category-pill {
  font-size: 11px;
  color: #64748b;
}

.stock-pill {
  background: #dcfce7;
  color: #15803d;
  font-size: 11px;
  font-weight: 700;
  padding: 4px 8px;
  border-radius: 6px;
}

.uom-selection-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.lbl {
  font-size: 12px;
  font-weight: 700;
  color: #475569;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.unit-pills {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.uom-pill {
  padding: 8px 12px;
  border: 1px solid #cbd5e1;
  background: #f8fafc;
  border-radius: 8px;
  cursor: pointer;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  transition: all 0.15s ease;
}

.uom-pill:hover {
  background: #f1f5f9;
}

.uom-pill.active {
  background: #0284c7;
  border-color: #0284c7;
  color: white;
}

.uom-name {
  font-size: 12px;
  font-weight: 700;
}

.uom-ratio {
  font-size: 10px;
  opacity: 0.85;
}

.quantity-input-row {
  display: grid;
  grid-template-columns: 1.2fr 1fr;
  gap: 16px;
  background: #f8fafc;
  padding: 14px;
  border-radius: 8px;
}

.qty-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.qty-field label {
  font-size: 12px;
  font-weight: 600;
  color: #475569;
}

.input-wrap {
  display: flex;
  align-items: center;
  gap: 6px;
}

.qty-input {
  width: 100px;
  padding: 8px 10px;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 700;
  text-align: center;
}

.unit-suffix {
  font-size: 13px;
  color: #64748b;
  font-weight: 600;
}

.equivalent-preview {
  display: flex;
  flex-direction: column;
  justify-content: center;
  border-left: 1px solid #e2e8f0;
  padding-left: 16px;
}

.eq-lbl {
  font-size: 11px;
  color: #64748b;
}

.eq-val {
  font-size: 16px;
  color: #0f172a;
}

.bulk-save {
  font-size: 11px;
  color: #16a34a;
  font-weight: 700;
  margin-top: 2px;
}

/* Table */
.pricing-breakdown-table {
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  overflow: hidden;
  font-size: 12px;
}

.t-row {
  display: grid;
  grid-template-columns: 1.6fr 1.6fr 1fr 1fr;
  padding: 8px 12px;
  border-bottom: 1px solid #f1f5f9;
  align-items: center;
}

.t-head {
  background: #f1f5f9;
  font-weight: 700;
  color: #475569;
  text-transform: uppercase;
  font-size: 10px;
}

.t-row.highlight {
  background: #f0fdf4;
  font-weight: 600;
}

.price-col {
  font-weight: 700;
  color: #0f172a;
}

.eff-col {
  color: #16a34a;
  font-weight: 600;
}

.action-footer {
  margin-top: 10px;
  padding-top: 14px;
  border-top: 2px solid #e2e8f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.total-lbl {
  font-size: 11px;
  color: #64748b;
  display: block;
}

.total-val {
  font-size: 18px;
  color: #0f172a;
}

.btn-add-cart {
  background: #0284c7;
  color: white;
  border: none;
  padding: 10px 18px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  transition: background 0.15s ease;
}

.btn-add-cart:hover {
  background: #0369a1;
}

@media (max-width: 768px) {
  .converter-layout {
    grid-template-columns: 1fr;
  }
}
</style>
