<template>
  <div class="weighed-items-manager">
    <div class="manager-header">
      <div>
        <h3>⚖️ Weighed Items & Scale Station</h3>
        <p class="subtitle">Sell fresh produce, bulk dry goods, and butchery items by weight</p>
      </div>
      <div class="scale-status-badge" :class="{ 'connected': scaleConnected }">
        <span class="dot"></span>
        {{ scaleConnected ? 'Digital Scale Connected (COM3 / 9600 baud)' : 'Manual Tare / Simulated Mode' }}
      </div>
    </div>

    <!-- Active Scale Reading Panel -->
    <div class="scale-reading-panel">
      <div class="digital-display">
        <div class="display-label">DIGITAL SCALE NET WEIGHT</div>
        <div class="display-value">
          <span class="number">{{ currentWeightKg.toFixed(3) }}</span>
          <span class="unit">kg</span>
        </div>
        <div class="display-sub">
          <span>Gross: {{ (currentWeightKg + tareKg).toFixed(3) }} kg</span>
          <span>Tare: {{ tareKg.toFixed(3) }} kg</span>
          <span>Zero: Calibrated</span>
        </div>
      </div>

      <div class="scale-controls">
        <label class="control-label">Quick Simulated Weight:</label>
        <div class="preset-weights">
          <button type="button" class="btn-chip" @click="setWeight(0.250)">250 g</button>
          <button type="button" class="btn-chip" @click="setWeight(0.500)">500 g</button>
          <button type="button" class="btn-chip" @click="setWeight(1.000)">1.00 kg</button>
          <button type="button" class="btn-chip" @click="setWeight(1.450)">1.45 kg</button>
          <button type="button" class="btn-chip" @click="setWeight(2.250)">2.25 kg</button>
        </div>

        <div class="manual-input-row">
          <label for="manualWeightInput">Manual Weight (kg):</label>
          <div class="input-with-actions">
            <input
              id="manualWeightInput"
              type="number"
              step="0.005"
              min="0.005"
              max="50"
              v-model.number="currentWeightKg"
              class="weight-input"
            />
            <button type="button" class="btn-secondary" @click="tareScale">Tare Zero</button>
            <button type="button" class="btn-secondary" @click="resetWeight">Clear</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Weighed Catalog Grid -->
    <div class="catalog-section">
      <h4>Select Produce / Weighed Item:</h4>
      <div class="produce-grid">
        <div
          v-for="item in weighedCatalog"
          :key="item.id"
          class="produce-card"
          :class="{ 'selected': selectedItem?.id === item.id }"
          @click="selectItem(item)"
        >
          <div class="produce-icon">{{ item.icon }}</div>
          <div class="produce-info">
            <div class="produce-name">{{ item.name }}</div>
            <div class="produce-sku">PLU #{{ item.plu }} · {{ item.category }}</div>
            <div class="produce-price">KES {{ (item.pricePerKgMinor / 100).toFixed(2) }} / kg</div>
          </div>
          <div class="produce-cost-preview" v-if="currentWeightKg > 0">
            <span class="preview-total">KES {{ calculateCost(item.pricePerKgMinor).toFixed(2) }}</span>
            <span class="preview-weight">for {{ currentWeightKg.toFixed(3) }} kg</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Action Bar / Barcode simulation -->
    <div class="action-footer" v-if="selectedItem">
      <div class="calculation-summary">
        <strong>{{ selectedItem.name }}</strong>:
        {{ currentWeightKg.toFixed(3) }} kg × KES {{ (selectedItem.pricePerKgMinor / 100).toFixed(2) }}/kg =
        <span class="highlight-total">KES {{ calculateCost(selectedItem.pricePerKgMinor).toFixed(2) }}</span>
      </div>

      <div class="footer-buttons">
        <button type="button" class="btn-primary" :disabled="currentWeightKg <= 0" @click="addToCartAction">
          ➕ Add to POS Cart
        </button>
        <button type="button" class="btn-secondary" @click="generateWeightBarcode">
          🏷️ Generate Scale Barcode
        </button>
      </div>
    </div>

    <!-- Generated Barcode Display Modal/Box -->
    <div v-if="generatedBarcode" class="barcode-preview-box">
      <div class="barcode-header">
        <span>Standard EAN-13 Weight-Embedded Barcode:</span>
        <button class="close-btn" @click="generatedBarcode = null">✕</button>
      </div>
      <div class="barcode-number">{{ generatedBarcode.barcode }}</div>
      <div class="barcode-breakdown">
        <span>Prefix: <code>{{ generatedBarcode.prefix }}</code></span>
        <span>PLU Code: <code>{{ generatedBarcode.plu }}</code></span>
        <span>Weight: <code>{{ generatedBarcode.weightGrams }} g</code> ({{ generatedBarcode.weightKg }} kg)</span>
        <span>Price: <code>KES {{ generatedBarcode.cost }}</code></span>
      </div>
      <small>Format: <code>21 + [5-digit PLU] + [5-digit grams] + [check]</code>. Barcode scanner will auto-weigh on checkout.</small>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const emit = defineEmits(['add-item', 'notification'])

const scaleConnected = ref(true)
const currentWeightKg = ref(1.250)
const tareKg = ref(0.000)
const selectedItem = ref(null)
const generatedBarcode = ref(null)

const weighedCatalog = [
  { id: 'w-1', plu: '00101', name: 'Bananas (Cavendish)', category: 'Fruits', pricePerKgMinor: 12000, icon: '🍌' },
  { id: 'w-2', plu: '00102', name: 'Red Onions (Local)', category: 'Vegetables', pricePerKgMinor: 15000, icon: '🧅' },
  { id: 'w-3', plu: '00103', name: 'Irish Potatoes (Shangi)', category: 'Vegetables', pricePerKgMinor: 9000, icon: '🥔' },
  { id: 'w-4', plu: '00104', name: 'Fresh Salad Tomatoes', category: 'Vegetables', pricePerKgMinor: 16000, icon: '🍅' },
  { id: 'w-5', plu: '00105', name: 'Prime Beef Steak', category: 'Butchery', pricePerKgMinor: 68000, icon: '🥩' },
  { id: 'w-6', plu: '00106', name: 'Pink Lady Apples', category: 'Fruits', pricePerKgMinor: 32000, icon: '🍎' },
  { id: 'w-7', plu: '00107', name: 'Chicken Wings (Fresh)', category: 'Poultry', pricePerKgMinor: 52000, icon: '🍗' },
  { id: 'w-8', plu: '00108', name: 'Hass Avocados', category: 'Fruits', pricePerKgMinor: 20000, icon: '🥑' },
]

selectedItem.value = weighedCatalog[0]

function setWeight(kg) {
  currentWeightKg.value = kg
}

function tareScale() {
  tareKg.value = 0.050 // tare 50g container
  emit('notification', { message: 'Scale tared with 0.050kg container offset', type: 'info' })
}

function resetWeight() {
  currentWeightKg.value = 0
  tareKg.value = 0
}

function selectItem(item) {
  selectedItem.value = item
}

function calculateCost(pricePerKgMinor) {
  const priceKes = pricePerKgMinor / 100
  return Number((priceKes * currentWeightKg.value).toFixed(2))
}

function addToCartAction() {
  if (!selectedItem.value || currentWeightKg.value <= 0) return
  const totalMinor = Math.round(selectedItem.value.pricePerKgMinor * currentWeightKg.value)
  const cartItem = {
    id: `weighed-${selectedItem.value.id}-${Date.now()}`,
    name: `${selectedItem.value.name} (${currentWeightKg.value.toFixed(3)} kg)`,
    sku: `PLU-${selectedItem.value.plu}`,
    barcode: `21${selectedItem.value.plu}${Math.round(currentWeightKg.value * 1000).toString().padStart(5, '0')}`,
    price_minor: totalMinor,
    unit_price_minor: selectedItem.value.pricePerKgMinor,
    quantity: 1,
    weight_kg: currentWeightKg.value,
    kind: 'variable',
  }

  // Also store in global custom event or localStorage cart queue
  emit('add-item', cartItem)
  emit('notification', {
    message: `Added ${selectedItem.value.name} (${currentWeightKg.value} kg · KES ${(totalMinor / 100).toFixed(2)}) to sale`,
    type: 'success',
  })
}

function generateWeightBarcode() {
  if (!selectedItem.value) return
  const grams = Math.round(currentWeightKg.value * 1000)
  const prefix = '21'
  const plu = selectedItem.value.plu.padStart(5, '0')
  const gramsStr = grams.toString().padStart(5, '0')
  const raw12 = `${prefix}${plu}${gramsStr}`
  
  // Calculate checksum for EAN-13
  let sum = 0
  for (let i = 0; i < 12; i++) {
    const digit = parseInt(raw12[i], 10)
    sum += i % 2 === 0 ? digit : digit * 3
  }
  const check = (10 - (sum % 10)) % 10
  const barcode = `${raw12}${check}`

  generatedBarcode.value = {
    barcode,
    prefix,
    plu,
    weightGrams: grams,
    weightKg: currentWeightKg.value.toFixed(3),
    cost: calculateCost(selectedItem.value.pricePerKgMinor).toFixed(2),
  }
}
</script>

<style scoped>
.weighed-items-manager {
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
  color: var(--color-muted, #64748b);
  font-size: 0.9rem;
}
.scale-status-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.8rem;
  font-weight: 600;
  padding: 0.35rem 0.75rem;
  border-radius: 999px;
  background: #f1f5f9;
  color: #475569;
}
.scale-status-badge.connected {
  background: #dcfce7;
  color: #166534;
}
.scale-status-badge .dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: currentColor;
}
.scale-reading-panel {
  display: grid;
  grid-template-columns: minmax(240px, 1fr) 2fr;
  gap: 1.25rem;
  background: #0f172a;
  color: #f8fafc;
  padding: 1.25rem;
  border-radius: 12px;
}
.digital-display {
  display: flex;
  flex-direction: column;
  justify-content: center;
  border-right: 1px solid #334155;
  padding-right: 1.25rem;
}
.display-label {
  font-size: 0.75rem;
  letter-spacing: 0.08em;
  color: #94a3b8;
  font-weight: 600;
}
.display-value {
  font-family: monospace;
  font-size: 2.75rem;
  font-weight: 800;
  color: #38bdf8;
  line-height: 1.1;
  margin: 0.25rem 0;
}
.display-value .unit {
  font-size: 1.25rem;
  color: #cbd5e1;
  margin-left: 0.35rem;
}
.display-sub {
  display: flex;
  gap: 0.75rem;
  font-size: 0.75rem;
  color: #94a3b8;
}
.scale-controls {
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 0.75rem;
}
.control-label {
  font-size: 0.8rem;
  color: #cbd5e1;
}
.preset-weights {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}
.btn-chip {
  background: #1e293b;
  border: 1px solid #334155;
  color: #e2e8f0;
  padding: 0.35rem 0.75rem;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 600;
  font-size: 0.85rem;
  transition: all 0.15s;
}
.btn-chip:hover {
  background: #38bdf8;
  color: #0f172a;
  border-color: #38bdf8;
}
.manual-input-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
}
.input-with-actions {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}
.weight-input {
  width: 110px;
  padding: 0.4rem 0.6rem;
  border-radius: 6px;
  border: 1px solid #475569;
  background: #1e293b;
  color: #fff;
  font-weight: bold;
}
.catalog-section h4 {
  margin: 0 0 0.75rem;
  font-size: 1rem;
}
.produce-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(190px, 1fr));
  gap: 0.85rem;
}
.produce-card {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  padding: 0.85rem;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.15s ease;
}
.produce-card:hover {
  border-color: #0284c7;
  transform: translateY(-1px);
  box-shadow: 0 4px 10px rgba(0,0,0,0.05);
}
.produce-card.selected {
  border: 2px solid #0284c7;
  background: #f0f9ff;
}
.produce-icon {
  font-size: 1.75rem;
}
.produce-name {
  font-weight: 700;
  font-size: 0.95rem;
  color: #1e293b;
}
.produce-sku {
  font-size: 0.75rem;
  color: #64748b;
}
.produce-price {
  font-size: 0.85rem;
  font-weight: 700;
  color: #0369a1;
}
.produce-cost-preview {
  margin-top: 0.35rem;
  padding-top: 0.35rem;
  border-top: 1px dashed #cbd5e1;
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.preview-total {
  font-weight: 800;
  color: #0f172a;
}
.preview-weight {
  font-size: 0.75rem;
  color: #64748b;
}
.action-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 1rem;
  padding: 1rem 1.25rem;
  background: #f8fafc;
  border: 1px solid #cbd5e1;
  border-radius: 10px;
}
.calculation-summary {
  font-size: 1rem;
  color: #1e293b;
}
.highlight-total {
  color: #0284c7;
  font-weight: 800;
  font-size: 1.15rem;
}
.footer-buttons {
  display: flex;
  gap: 0.5rem;
}
.btn-primary {
  background: #0284c7;
  color: white;
  border: none;
  padding: 0.6rem 1.25rem;
  border-radius: 8px;
  font-weight: 700;
  cursor: pointer;
  transition: background 0.15s;
}
.btn-primary:hover:not(:disabled) {
  background: #0369a1;
}
.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.btn-secondary {
  background: #fff;
  border: 1px solid #cbd5e1;
  color: #334155;
  padding: 0.5rem 0.85rem;
  border-radius: 6px;
  font-weight: 600;
  cursor: pointer;
}
.btn-secondary:hover {
  background: #f1f5f9;
}
.barcode-preview-box {
  background: #fff;
  border: 2px dashed #0284c7;
  border-radius: 10px;
  padding: 1rem 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
.barcode-header {
  display: flex;
  justify-content: space-between;
  font-weight: 700;
  color: #1e293b;
}
.close-btn {
  background: none;
  border: none;
  cursor: pointer;
  font-weight: bold;
}
.barcode-number {
  font-family: monospace;
  font-size: 1.75rem;
  font-weight: 900;
  letter-spacing: 0.15em;
  color: #0f172a;
}
.barcode-breakdown {
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
  font-size: 0.85rem;
}
.barcode-breakdown code {
  background: #f1f5f9;
  padding: 0.1rem 0.35rem;
  border-radius: 4px;
}
@media (max-width: 768px) {
  .scale-reading-panel {
    grid-template-columns: 1fr;
  }
  .digital-display {
    border-right: none;
    border-bottom: 1px solid #334155;
    padding-bottom: 1rem;
    padding-right: 0;
  }
}
</style>
