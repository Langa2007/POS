<template>
  <div class="generic-substitution-finder">
    <div class="manager-header">
      <div>
        <h3>💊 Generic Drug Substitution & Formulary Guide</h3>
        <p class="subtitle">Search by brand name or active molecule (INN) to recommend bioequivalent, cost-effective generic alternatives</p>
      </div>
    </div>

    <!-- Search Input -->
    <div class="search-card">
      <label for="drugSearch">Search Brand Name or Active Pharmaceutical Ingredient (API):</label>
      <div class="search-row">
        <input
          id="drugSearch"
          type="text"
          v-model="searchQuery"
          placeholder="e.g. Panadol, Augmentin, Ventolin, Lipitor, Zithromax, Paracetamol..."
          class="search-input"
          @keydown.enter.prevent="performSearch"
        />
        <button type="button" class="btn-primary" @click="performSearch">
          🔍 Lookup Substitutes
        </button>
      </div>
      <div class="quick-drugs">
        <span>Quick Lookups:</span>
        <button type="button" class="btn-tag" @click="quickSearch('Panadol')">Panadol (Paracetamol)</button>
        <button type="button" class="btn-tag" @click="quickSearch('Augmentin')">Augmentin (Co-Amoxiclav)</button>
        <button type="button" class="btn-tag" @click="quickSearch('Ventolin')">Ventolin (Salbutamol)</button>
        <button type="button" class="btn-tag" @click="quickSearch('Lipitor')">Lipitor (Atorvastatin)</button>
      </div>
    </div>

    <!-- Search Result Substitution Comparison -->
    <div v-if="activeMoleculeGroup" class="comparison-card">
      <div class="molecule-header">
        <div>
          <span class="molecule-badge">ACTIVE INGREDIENT (INN)</span>
          <h4>{{ activeMoleculeGroup.molecule }} ({{ activeMoleculeGroup.strength }})</h4>
          <p class="class-text">Therapeutic Class: {{ activeMoleculeGroup.therapeuticClass }}</p>
        </div>
        <div class="savings-banner" v-if="maxSavingsPercent > 0">
          <span class="pct">{{ maxSavingsPercent }}%</span>
          <span class="text">Max Patient Cost Savings With Generic</span>
        </div>
      </div>

      <!-- Equivalence Table -->
      <div class="equivalents-table-wrapper">
        <table class="equivalents-table">
          <thead>
            <tr>
              <th>Brand / Product Name</th>
              <th>Manufacturer</th>
              <th>Form</th>
              <th>Type</th>
              <th>Price (Pack)</th>
              <th>Unit Cost</th>
              <th>Stock Status</th>
              <th>Action</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="drug in activeMoleculeGroup.equivalents"
              :key="drug.id"
              :class="{ 'brand-highlight': drug.isOriginalBrand, 'best-value': drug.isBestValue }"
            >
              <td>
                <div class="drug-name-cell">
                  <strong>{{ drug.brandName }}</strong>
                  <span v-if="drug.isOriginalBrand" class="origin-badge">Original Innovator</span>
                  <span v-if="drug.isBestValue" class="value-badge">⭐ Best Patient Value</span>
                </div>
              </td>
              <td>{{ drug.manufacturer }}</td>
              <td>{{ drug.form }}</td>
              <td>
                <span class="type-pill" :class="drug.isOriginalBrand ? 'brand' : 'generic'">
                  {{ drug.isOriginalBrand ? 'Originator Brand' : 'Bioequivalent Generic' }}
                </span>
              </td>
              <td><strong>KES {{ (drug.packPriceMinor / 100).toFixed(2) }}</strong></td>
              <td>KES {{ (drug.unitPriceMinor / 100).toFixed(2) }} / tab</td>
              <td>
                <span class="stock-pill" :class="drug.stock > 0 ? 'in-stock' : 'out-stock'">
                  {{ drug.stock > 0 ? `${drug.stock} in stock` : 'Out of stock' }}
                </span>
              </td>
              <td>
                <button
                  type="button"
                  class="btn-select-sub"
                  :disabled="drug.stock <= 0"
                  @click="selectSubstitute(drug)"
                >
                  {{ drug.isOriginalBrand ? 'Select Brand' : 'Swap to Generic' }}
                </button>
              </td>
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

const searchQuery = ref('Panadol')
const activeMoleculeGroup = ref(null)

const formularyDirectory = [
  {
    molecule: 'Paracetamol',
    strength: '500mg',
    therapeuticClass: 'Analgesic & Antipyretic',
    searchKeywords: ['panadol', 'paracetamol', 'calpol', 'para', 'fever', 'headache'],
    equivalents: [
      { id: 'sub-p1', brandName: 'Panadol Extra', manufacturer: 'GSK / Haleon', form: 'Tablet', packPriceMinor: 25000, unitPriceMinor: 2500, stock: 450, isOriginalBrand: true, isBestValue: false },
      { id: 'sub-p2', brandName: 'Calpol 500mg', manufacturer: 'Johnson & Johnson', form: 'Tablet', packPriceMinor: 22000, unitPriceMinor: 2200, stock: 180, isOriginalBrand: false, isBestValue: false },
      { id: 'sub-p3', brandName: 'Para-Denk 500mg', manufacturer: 'Denk Pharma (Germany)', form: 'Tablet', packPriceMinor: 14000, unitPriceMinor: 1400, stock: 320, isOriginalBrand: false, isBestValue: false },
      { id: 'sub-p4', brandName: 'Paracetamol Generics Local', manufacturer: 'Dawa Limited (Kenya)', form: 'Tablet', packPriceMinor: 7000, unitPriceMinor: 700, stock: 1200, isOriginalBrand: false, isBestValue: true },
    ],
  },
  {
    molecule: 'Amoxicillin + Clavulanic Acid (Co-Amoxiclav)',
    strength: '625mg',
    therapeuticClass: 'Broad-Spectrum Antibacterial (Penicillin + Beta-Lactamase Inhibitor)',
    searchKeywords: ['augmentin', 'amoxicillin', 'clavulanic', 'co-amoxiclav', 'clavamox', 'curam'],
    equivalents: [
      { id: 'sub-a1', brandName: 'Augmentin 625mg', manufacturer: 'GSK', form: 'Tablet', packPriceMinor: 240000, unitPriceMinor: 17140, stock: 65, isOriginalBrand: true, isBestValue: false },
      { id: 'sub-a2', brandName: 'Curam 625mg', manufacturer: 'Sandoz / Novartis', form: 'Tablet', packPriceMinor: 165000, unitPriceMinor: 11780, stock: 110, isOriginalBrand: false, isBestValue: false },
      { id: 'sub-a3', brandName: 'Clavamox 625mg', manufacturer: 'Square / Biodeal', form: 'Tablet', packPriceMinor: 95000, unitPriceMinor: 6780, stock: 240, isOriginalBrand: false, isBestValue: true },
    ],
  },
  {
    molecule: 'Salbutamol Sulfate',
    strength: '100mcg/dose',
    therapeuticClass: 'Short-Acting Beta-2 Adagonist (Bronchodilator)',
    searchKeywords: ['ventolin', 'salbutamol', 'asthma', 'inhaler', 'asmasol'],
    equivalents: [
      { id: 'sub-v1', brandName: 'Ventolin Evohaler', manufacturer: 'GSK', form: 'Inhaler (200 doses)', packPriceMinor: 65000, unitPriceMinor: 65000, stock: 28, isOriginalBrand: true, isBestValue: false },
      { id: 'sub-v2', brandName: 'Asmasol Inhaler', manufacturer: 'Beximco Pharma', form: 'Inhaler (200 doses)', packPriceMinor: 42000, unitPriceMinor: 42000, stock: 45, isOriginalBrand: false, isBestValue: true },
    ],
  },
  {
    molecule: 'Atorvastatin Calcium',
    strength: '20mg',
    therapeuticClass: 'HMG-CoA Reductase Inhibitor (Lipid-Lowering Statin)',
    searchKeywords: ['lipitor', 'atorvastatin', 'cholesterol', 'atorva', 'lipicure'],
    equivalents: [
      { id: 'sub-l1', brandName: 'Lipitor 20mg', manufacturer: 'Pfizer', form: 'Tablet (30s)', packPriceMinor: 380000, unitPriceMinor: 12660, stock: 18, isOriginalBrand: true, isBestValue: false },
      { id: 'sub-l2', brandName: 'Atorva 20mg', manufacturer: 'Zydus Cadila', form: 'Tablet (30s)', packPriceMinor: 145000, unitPriceMinor: 4830, stock: 75, isOriginalBrand: false, isBestValue: false },
      { id: 'sub-l3', brandName: 'Lipicure 20mg', manufacturer: 'Cipla', form: 'Tablet (30s)', packPriceMinor: 110000, unitPriceMinor: 3660, stock: 90, isOriginalBrand: false, isBestValue: true },
    ],
  },
]

// Initialize with Panadol
activeMoleculeGroup.value = formularyDirectory[0]

const maxSavingsPercent = computed(() => {
  if (!activeMoleculeGroup.value?.equivalents?.length) return 0
  const brand = activeMoleculeGroup.value.equivalents.find((e) => e.isOriginalBrand)
  const bestValue = activeMoleculeGroup.value.equivalents.find((e) => e.isBestValue)
  if (brand && bestValue) {
    const diff = brand.packPriceMinor - bestValue.packPriceMinor
    return Math.round((diff / brand.packPriceMinor) * 100)
  }
  return 0
})

function performSearch() {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return
  const match = formularyDirectory.find((item) =>
    item.searchKeywords.some((k) => k.includes(q) || q.includes(k)) ||
    item.molecule.toLowerCase().includes(q)
  )
  if (match) {
    activeMoleculeGroup.value = match
    emit('notification', { message: `Loaded generics for ${match.molecule}`, type: 'info' })
  } else {
    emit('notification', { message: `No exact formulary group found for "${searchQuery.value}". Showing Paracetamol.`, type: 'warning' })
    activeMoleculeGroup.value = formularyDirectory[0]
  }
}

function quickSearch(name) {
  searchQuery.value = name
  performSearch()
}

function selectSubstitute(drug) {
  const cartItem = {
    id: `sub-${drug.id}-${Date.now()}`,
    name: `${drug.brandName} (${activeMoleculeGroup.value.strength})`,
    sku: `SUB-${drug.id.toUpperCase()}`,
    price_minor: drug.packPriceMinor,
    unit_price_minor: drug.packPriceMinor,
    quantity: 1,
    kind: 'stocked',
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
    message: `Selected ${drug.brandName} (KES ${(drug.packPriceMinor / 100).toFixed(2)}) for POS cart!`,
    type: 'success',
  })
}
</script>

<style scoped>
.generic-substitution-finder {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
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
.search-card {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}
.search-card label {
  font-weight: 600;
  font-size: 0.9rem;
  color: #334155;
}
.search-row {
  display: flex;
  gap: 0.5rem;
  max-width: 580px;
}
.search-input {
  flex: 1;
  padding: 0.6rem 0.85rem;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  font-size: 0.95rem;
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
.quick-drugs {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.8rem;
  color: #64748b;
  flex-wrap: wrap;
}
.btn-tag {
  background: #e2e8f0;
  border: none;
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.75rem;
  font-weight: 600;
}
.comparison-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
.molecule-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  flex-wrap: wrap;
  gap: 1rem;
  border-bottom: 1px solid #e2e8f0;
  padding-bottom: 0.85rem;
}
.molecule-badge {
  font-size: 0.7rem;
  font-weight: 800;
  background: #e0f2fe;
  color: #0369a1;
  padding: 0.15rem 0.45rem;
  border-radius: 4px;
}
.molecule-header h4 {
  margin: 0.35rem 0 0.2rem;
  font-size: 1.35rem;
  color: #0f172a;
}
.class-text {
  margin: 0;
  font-size: 0.85rem;
  color: #64748b;
}
.savings-banner {
  background: #dcfce7;
  border: 1px solid #86efac;
  padding: 0.5rem 0.85rem;
  border-radius: 8px;
  display: flex;
  align-items: center;
  gap: 0.65rem;
}
.savings-banner .pct {
  font-size: 1.65rem;
  font-weight: 900;
  color: #166534;
}
.savings-banner .text {
  font-size: 0.8rem;
  font-weight: 700;
  color: #166534;
  line-height: 1.2;
}
.equivalents-table-wrapper {
  overflow-x: auto;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
}
.equivalents-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
  font-size: 0.85rem;
}
.equivalents-table th, .equivalents-table td {
  padding: 0.75rem 1rem;
  border-bottom: 1px solid #f1f5f9;
}
.equivalents-table th {
  background: #f8fafc;
  font-weight: 700;
  color: #475569;
}
.brand-highlight {
  background: #fafafa;
}
.best-value {
  background: #f0fdf4;
}
.drug-name-cell {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}
.origin-badge {
  font-size: 0.7rem;
  font-weight: 700;
  color: #6366f1;
}
.value-badge {
  font-size: 0.7rem;
  font-weight: 800;
  color: #166534;
}
.type-pill {
  font-size: 0.75rem;
  font-weight: 700;
  padding: 0.15rem 0.45rem;
  border-radius: 4px;
}
.type-pill.brand {
  background: #ede9fe;
  color: #5b21b6;
}
.type-pill.generic {
  background: #dcfce7;
  color: #166534;
}
.stock-pill {
  font-size: 0.75rem;
  font-weight: 600;
}
.stock-pill.in-stock { color: #166534; }
.stock-pill.out-stock { color: #dc2626; }
.btn-select-sub {
  background: #0284c7;
  color: white;
  border: none;
  padding: 0.35rem 0.75rem;
  border-radius: 6px;
  font-weight: 700;
  font-size: 0.75rem;
  cursor: pointer;
  white-space: nowrap;
}
.btn-select-sub:hover:not(:disabled) {
  background: #0369a1;
}
.btn-select-sub:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}
</style>
