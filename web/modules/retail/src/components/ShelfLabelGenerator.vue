<template>
  <div class="shelf-label-generator">
    <div class="manager-header">
      <div>
        <h3>🖨️ Shelf Price Labels & Barcode Tag Printer</h3>
        <p class="subtitle">Print high-visibility shelf edge talkers, barcode tags, and promotional stickers</p>
      </div>
      <div class="print-actions">
        <button class="btn-primary" @click="printAllLabels">
          🖨️ Print Shelf Labels ({{ labelItems.length }})
        </button>
      </div>
    </div>

    <!-- Template & Layout Controls -->
    <div class="controls-panel">
      <div class="control-group">
        <label>Label Template Format:</label>
        <select v-model="selectedFormat">
          <option value="standard">Standard Shelf Talker (80mm × 50mm)</option>
          <option value="compact">Compact Shelf Tag (58mm × 40mm)</option>
          <option value="promo">Promotional Yellow Endcap (Large)</option>
        </select>
      </div>
      <div class="control-group">
        <label>Store Brand Name:</label>
        <input type="text" v-model="storeName" placeholder="e.g. QuickMart Supermarket" />
      </div>
      <div class="control-checkbox">
        <label>
          <input type="checkbox" v-model="includeBarcode" /> Include EAN-13 Scannable Barcode
        </label>
        <label>
          <input type="checkbox" v-model="includeUnitPricing" /> Include Unit of Measure (UoM)
        </label>
      </div>
    </div>

    <!-- Live Preview Carousel / Grid -->
    <div class="preview-section">
      <h4>Shelf Labels Preview:</h4>
      <div class="labels-grid" :class="selectedFormat">
        <div
          v-for="(item, index) in labelItems"
          :key="item.id"
          class="shelf-tag"
          :class="[selectedFormat, { 'promo-border': item.isPromo }]"
        >
          <div class="tag-header">
            <span class="store-brand">{{ storeName }}</span>
            <span v-if="item.isPromo" class="promo-badge-tag">SPECIAL OFFER</span>
            <span v-else class="dept-tag">{{ item.dept }}</span>
          </div>

          <div class="tag-body">
            <div class="tag-item-name">{{ item.name }}</div>
            <div class="tag-sku">SKU: {{ item.sku }}</div>

            <div class="tag-price-block">
              <span class="currency-sign">KES</span>
              <span class="price-major">{{ Math.floor(item.priceMinor / 100) }}</span>
              <span class="price-cents">.{{ (item.priceMinor % 100).toString().padStart(2, '0') }}</span>
              <span v-if="includeUnitPricing" class="uom-suffix">/ {{ item.uom }}</span>
            </div>
          </div>

          <div class="tag-footer" v-if="includeBarcode">
            <div class="barcode-visual">
              <div class="barcode-bars"></div>
              <span class="barcode-digits">{{ item.barcode }}</span>
            </div>
            <div class="tag-date">Prc: {{ item.dateUpdated }}</div>
          </div>

          <button class="remove-label-btn" title="Remove" @click="removeItem(index)">✕</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const emit = defineEmits(['notification'])

const storeName = ref('QuickMart Supermarket')
const selectedFormat = ref('standard')
const includeBarcode = ref(true)
const includeUnitPricing = ref(true)

const labelItems = ref([
  {
    id: 'l-1',
    name: 'Basmati Rice Premium 2kg',
    sku: 'RICE-BAS-2KG',
    barcode: '6161100223344',
    priceMinor: 48000,
    uom: '2kg pack',
    dept: 'Grocery / Grains',
    dateUpdated: '10/11/2026',
    isPromo: true,
  },
  {
    id: 'l-2',
    name: 'Pure Vegetable Cooking Oil 3L',
    sku: 'OIL-VEG-3L',
    barcode: '6161100998877',
    priceMinor: 75000,
    uom: '3L Can',
    dept: 'Edible Oils',
    dateUpdated: '10/11/2026',
    isPromo: false,
  },
  {
    id: 'l-3',
    name: 'Cavendish Sweet Bananas',
    sku: 'PLU-00101',
    barcode: '2100101010005',
    priceMinor: 12000,
    uom: 'kg',
    dept: 'Fresh Produce',
    dateUpdated: '10/11/2026',
    isPromo: false,
  },
  {
    id: 'l-4',
    name: 'Drinking Chocolate 500g',
    sku: 'CHOC-500G',
    barcode: '6161100554433',
    priceMinor: 32000,
    uom: '500g tin',
    dept: 'Beverages',
    dateUpdated: '10/11/2026',
    isPromo: true,
  },
])

function removeItem(index) {
  labelItems.value.splice(index, 1)
}

function printAllLabels() {
  window.print()
  emit('notification', { message: `Printing ${labelItems.value.length} shelf labels...`, type: 'info' })
}
</script>

<style scoped>
.shelf-label-generator {
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
  padding: 0.6rem 1.25rem;
  border-radius: 8px;
  font-weight: 700;
  cursor: pointer;
}
.controls-panel {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 1.25rem;
  background: #f8fafc;
  padding: 1rem 1.25rem;
  border-radius: 10px;
  border: 1px solid #e2e8f0;
}
.control-group {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  font-size: 0.85rem;
  font-weight: 600;
}
.control-group select, .control-group input {
  padding: 0.45rem 0.75rem;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  background: white;
}
.control-checkbox {
  display: flex;
  gap: 1rem;
  font-size: 0.85rem;
  font-weight: 600;
  color: #334155;
}
.control-checkbox label {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  cursor: pointer;
}
.preview-section h4 {
  margin: 0 0 0.85rem;
  font-size: 1rem;
}
.labels-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 1rem;
}
.shelf-tag {
  background: white;
  border: 2px solid #0f172a;
  border-radius: 8px;
  padding: 0.85rem 1rem;
  display: flex;
  flex-direction: column;
  position: relative;
  box-shadow: 0 4px 10px rgba(0,0,0,0.06);
}
.shelf-tag.promo {
  background: #fefce8;
  border: 3px solid #eab308;
}
.shelf-tag.promo-border {
  border-color: #eab308;
}
.tag-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.75rem;
  font-weight: 800;
  border-bottom: 1px solid #cbd5e1;
  padding-bottom: 0.35rem;
}
.store-brand {
  text-transform: uppercase;
  color: #0f172a;
}
.dept-tag {
  color: #64748b;
}
.promo-badge-tag {
  background: #ef4444;
  color: white;
  padding: 0.1rem 0.4rem;
  border-radius: 4px;
}
.tag-body {
  margin: 0.5rem 0;
}
.tag-item-name {
  font-size: 1.05rem;
  font-weight: 800;
  color: #0f172a;
  line-height: 1.2;
}
.tag-sku {
  font-size: 0.75rem;
  color: #64748b;
  margin-top: 0.15rem;
}
.tag-price-block {
  display: flex;
  align-items: baseline;
  gap: 0.2rem;
  margin-top: 0.4rem;
}
.currency-sign {
  font-size: 1rem;
  font-weight: 800;
  color: #0f172a;
}
.price-major {
  font-size: 2.25rem;
  font-weight: 900;
  color: #0f172a;
  line-height: 1;
}
.price-cents {
  font-size: 1.15rem;
  font-weight: 800;
  color: #0f172a;
}
.uom-suffix {
  font-size: 0.85rem;
  font-weight: 600;
  color: #64748b;
  margin-left: 0.35rem;
}
.tag-footer {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  border-top: 1px dashed #cbd5e1;
  padding-top: 0.35rem;
}
.barcode-visual {
  display: flex;
  flex-direction: column;
}
.barcode-bars {
  height: 20px;
  width: 140px;
  background: repeating-linear-gradient(
    90deg,
    #000,
    #000 2px,
    #fff 2px,
    #fff 4px,
    #000 4px,
    #000 5px,
    #fff 5px,
    #fff 7px
  );
}
.barcode-digits {
  font-family: monospace;
  font-size: 0.75rem;
  letter-spacing: 0.1em;
  color: #0f172a;
}
.tag-date {
  font-size: 0.7rem;
  color: #94a3b8;
}
.remove-label-btn {
  position: absolute;
  top: 4px;
  right: 4px;
  background: none;
  border: none;
  color: #94a3b8;
  cursor: pointer;
  font-size: 0.85rem;
}
.remove-label-btn:hover {
  color: #ef4444;
}

@media print {
  body * {
    visibility: hidden;
  }
  .shelf-label-generator, .shelf-label-generator * {
    visibility: visible;
  }
  .controls-panel, .manager-header, .remove-label-btn {
    display: none !important;
  }
  .shelf-label-generator {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
  }
}
</style>
