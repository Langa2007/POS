<template>
  <aside class="retail-checkout-sidebar">
    <div class="sidebar-header">
      <span class="header-icon">🛒</span>
      <div>
        <h4>Retail Till Options</h4>
        <small>Scale, Loyalty & Promotions</small>
      </div>
    </div>

    <!-- 1. Scale & Weighed Item Quick Action -->
    <section class="sidebar-block">
      <div class="block-title">
        <span>⚖️ Scale & Weighed Item</span>
        <button class="btn-micro" @click="showScaleModal = true">Open Scale</button>
      </div>
      <div class="quick-weigh-box">
        <select v-model="selectedProduce" class="micro-select">
          <option v-for="p in produceOptions" :key="p.id" :value="p">
            {{ p.icon }} {{ p.name }} (KES {{ (p.pricePerKgMinor / 100).toFixed(0) }}/kg)
          </option>
        </select>
        <div class="weight-row">
          <input
            type="number"
            step="0.05"
            min="0.05"
            v-model.number="quickWeightKg"
            class="weight-input"
          />
          <span class="unit">kg = KES {{ Math.round((selectedProduce?.pricePerKgMinor / 100) * quickWeightKg) }}</span>
          <button type="button" class="btn-add-weighed" @click="addWeighedLine">
            ➕ Add
          </button>
        </div>
      </div>
    </section>

    <!-- 2. Customer Loyalty Member Lookup -->
    <section class="sidebar-block">
      <div class="block-title">
        <span>⭐ Customer Loyalty</span>
        <span v-if="attachedCustomer" class="attached-pill">{{ attachedCustomer.name }}</span>
      </div>
      <div v-if="!attachedCustomer" class="loyalty-search-box">
        <input
          type="tel"
          v-model="loyaltyPhone"
          placeholder="Phone (e.g. 0712345678)"
          class="micro-input"
          @keydown.enter.prevent="searchLoyalty"
        />
        <button type="button" class="btn-micro-action" @click="searchLoyalty">Find</button>
      </div>
      <div v-else class="loyalty-attached-box">
        <div class="loyalty-balance">
          <span>Pts: <strong>{{ attachedCustomer.points }}</strong></span>
          <span>= KES {{ (attachedCustomer.points / 10).toFixed(0) }}</span>
        </div>
        <div class="loyalty-actions-row">
          <button type="button" class="btn-redeem-micro" @click="redeemLoyaltyPoints">
            Redeem Points
          </button>
          <button type="button" class="btn-clear-micro" @click="attachedCustomer = null">✕</button>
        </div>
      </div>
    </section>

    <!-- 3. Store Promo Code -->
    <section class="sidebar-block">
      <div class="block-title">
        <span>🏷️ Promo Code / Coupon</span>
      </div>
      <div class="coupon-box">
        <input
          type="text"
          v-model="promoCode"
          placeholder="e.g. SAVE10, WKND200"
          class="micro-input uppercase"
          @keydown.enter.prevent="applyPromo"
        />
        <button type="button" class="btn-micro-action" @click="applyPromo">Apply</button>
      </div>
      <div v-if="appliedPromoNotice" class="applied-promo-tag">
        ✅ {{ appliedPromoNotice }}
      </div>
    </section>

    <!-- Scale Modal Dialog -->
    <div v-if="showScaleModal" class="scale-modal-backdrop">
      <div class="scale-modal-dialog">
        <div class="modal-top">
          <h5>Digital Produce Scale</h5>
          <button class="close-x" @click="showScaleModal = false">✕</button>
        </div>
        <div class="modal-scale-reading">
          <div class="big-weight">{{ quickWeightKg.toFixed(3) }} <small>kg</small></div>
          <div class="scale-chips">
            <button @click="quickWeightKg = 0.5">500g</button>
            <button @click="quickWeightKg = 1.0">1.0kg</button>
            <button @click="quickWeightKg = 1.5">1.5kg</button>
            <button @click="quickWeightKg = 2.25">2.25kg</button>
          </div>
        </div>
        <div class="modal-produce-select">
          <label>Select Item:</label>
          <select v-model="selectedProduce" class="full-select">
            <option v-for="p in produceOptions" :key="p.id" :value="p">
              {{ p.icon }} {{ p.name }} · KES {{ (p.pricePerKgMinor / 100).toFixed(0) }}/kg
            </option>
          </select>
        </div>
        <div class="modal-total-preview">
          Total: <strong>KES {{ Math.round((selectedProduce?.pricePerKgMinor / 100) * quickWeightKg) }}</strong>
        </div>
        <div class="modal-buttons">
          <button class="btn-modal-cancel" @click="showScaleModal = false">Cancel</button>
          <button class="btn-modal-add" @click="addWeighedLine(); showScaleModal = false">
            ➕ Add to Cart
          </button>
        </div>
      </div>
    </div>
  </aside>
</template>

<script setup>
import { ref } from 'vue'

const props = defineProps({
  cart: { type: Array, default: () => [] },
})

const emit = defineEmits(['add-item', 'apply-discount', 'set-customer'])

const showScaleModal = ref(false)
const quickWeightKg = ref(1.0)
const loyaltyPhone = ref('')
const attachedCustomer = ref(null)
const promoCode = ref('')
const appliedPromoNotice = ref('')

const produceOptions = [
  { id: 'w-1', plu: '00101', name: 'Bananas', pricePerKgMinor: 12000, icon: '🍌' },
  { id: 'w-2', plu: '00102', name: 'Red Onions', pricePerKgMinor: 15000, icon: '🧅' },
  { id: 'w-3', plu: '00103', name: 'Irish Potatoes', pricePerKgMinor: 9000, icon: '🥔' },
  { id: 'w-4', plu: '00104', name: 'Fresh Tomatoes', pricePerKgMinor: 16000, icon: '🍅' },
  { id: 'w-5', plu: '00105', name: 'Prime Beef Steak', pricePerKgMinor: 68000, icon: '🥩' },
]

const selectedProduce = ref(produceOptions[0])

function addWeighedLine() {
  if (!selectedProduce.value || quickWeightKg.value <= 0) return
  const costMinor = Math.round(selectedProduce.value.pricePerKgMinor * quickWeightKg.value)
  const item = {
    id: `weighed-${selectedProduce.value.id}-${Date.now()}`,
    name: `${selectedProduce.value.name} (${quickWeightKg.value} kg)`,
    sku: `PLU-${selectedProduce.value.plu}`,
    price_minor: costMinor,
    unit_price_minor: selectedProduce.value.pricePerKgMinor,
    quantity: 1,
    weight_kg: quickWeightKg.value,
    kind: 'variable',
  }

  // Check if parent listens to add-item or we trigger local storage
  emit('add-item', item)

  // Also push to localStorage pending
  try {
    const list = JSON.parse(localStorage.getItem('pos.retailCartPending') || '[]')
    list.push(item)
    localStorage.setItem('pos.retailCartPending', JSON.stringify(list))
    // Trigger custom event for CheckoutView
    window.dispatchEvent(new CustomEvent('pos:add-item', { detail: item }))
  } catch (e) {
    console.error(e)
  }
}

function searchLoyalty() {
  const p = loyaltyPhone.value.trim()
  if (!p) return
  // Sample members
  const mockMembers = {
    '0712345678': { name: 'John Kamau', points: 3450, tier: 'Gold' },
    '0722001122': { name: 'Mary Wanjiku', points: 1200, tier: 'Silver' },
    '0799887766': { name: 'Peter Omondi', points: 450, tier: 'Bronze' },
  }
  const match = mockMembers[p] || { name: `Customer ${p.slice(-4)}`, points: 650, tier: 'Silver' }
  attachedCustomer.value = match
  emit('set-customer', match)
}

function redeemLoyaltyPoints() {
  if (!attachedCustomer.value) return
  const pts = Math.min(1000, attachedCustomer.value.points)
  const discountMinor = (pts / 10) * 100
  attachedCustomer.value.points -= pts
  const discount = {
    code: 'LOYALTY',
    description: `Loyalty Points (${pts} pts)`,
    amountMinor: discountMinor,
  }
  emit('apply-discount', discount)
  window.dispatchEvent(new CustomEvent('pos:apply-discount', { detail: discount }))
  appliedPromoNotice.value = `Redeemed ${pts} pts (KES ${discountMinor / 100})`
}

function applyPromo() {
  const code = promoCode.value.trim().toUpperCase()
  if (!code) return
  let discount = null
  if (code === 'SAVE10') {
    discount = { code: 'SAVE10', description: '10% Store Discount', percent: 10 }
  } else if (code === 'WKND200') {
    discount = { code: 'WKND200', description: 'KES 200 Voucher', amountMinor: 20000 }
  } else {
    discount = { code, description: `Promo ${code}`, percent: 5 }
  }
  emit('apply-discount', discount)
  window.dispatchEvent(new CustomEvent('pos:apply-discount', { detail: discount }))
  appliedPromoNotice.value = `Applied ${discount.description}`
  promoCode.value = ''
}
</script>

<style scoped>
.retail-checkout-sidebar {
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 1rem;
}
.sidebar-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid #e2e8f0;
}
.header-icon {
  font-size: 1.35rem;
}
.sidebar-header h4 {
  margin: 0;
  font-size: 0.95rem;
  font-weight: 800;
  color: #0f172a;
}
.sidebar-header small {
  color: #64748b;
  font-size: 0.75rem;
}
.sidebar-block {
  background: white;
  border: 1px solid #e2e8f0;
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
  color: #334155;
}
.btn-micro {
  background: #e0f2fe;
  color: #0369a1;
  border: none;
  padding: 0.2rem 0.45rem;
  border-radius: 4px;
  font-size: 0.7rem;
  font-weight: 700;
  cursor: pointer;
}
.quick-weigh-box {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}
.micro-select {
  width: 100%;
  padding: 0.35rem 0.5rem;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 0.8rem;
}
.weight-row {
  display: flex;
  align-items: center;
  gap: 0.35rem;
}
.weight-input {
  width: 65px;
  padding: 0.35rem 0.4rem;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 0.85rem;
  font-weight: bold;
}
.unit {
  font-size: 0.75rem;
  color: #64748b;
  flex-grow: 1;
}
.btn-add-weighed {
  background: #0284c7;
  color: white;
  border: none;
  padding: 0.35rem 0.65rem;
  border-radius: 6px;
  font-weight: 700;
  font-size: 0.75rem;
  cursor: pointer;
}
.loyalty-search-box, .coupon-box {
  display: flex;
  gap: 0.35rem;
}
.micro-input {
  flex: 1;
  padding: 0.35rem 0.5rem;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 0.8rem;
}
.micro-input.uppercase {
  text-transform: uppercase;
}
.btn-micro-action {
  background: #0f172a;
  color: white;
  border: none;
  padding: 0.35rem 0.65rem;
  border-radius: 6px;
  font-weight: 700;
  font-size: 0.75rem;
  cursor: pointer;
}
.attached-pill {
  font-size: 0.7rem;
  background: #dcfce7;
  color: #166534;
  padding: 0.15rem 0.4rem;
  border-radius: 4px;
}
.loyalty-attached-box {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.8rem;
}
.loyalty-actions-row {
  display: flex;
  gap: 0.35rem;
}
.btn-redeem-micro {
  background: #166534;
  color: white;
  border: none;
  padding: 0.25rem 0.55rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 700;
  cursor: pointer;
}
.btn-clear-micro {
  background: none;
  border: 1px solid #cbd5e1;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.7rem;
}
.applied-promo-tag {
  font-size: 0.75rem;
  color: #166534;
  background: #dcfce7;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-weight: 600;
}
.scale-modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.5);
  display: grid;
  place-items: center;
  z-index: 1100;
}
.scale-modal-dialog {
  background: white;
  width: min(380px, 92vw);
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
.modal-scale-reading {
  background: #0f172a;
  color: #38bdf8;
  padding: 1rem;
  border-radius: 8px;
  text-align: center;
}
.big-weight {
  font-family: monospace;
  font-size: 2.5rem;
  font-weight: 800;
}
.big-weight small {
  font-size: 1rem;
  color: #94a3b8;
}
.scale-chips {
  display: flex;
  justify-content: center;
  gap: 0.35rem;
  margin-top: 0.5rem;
}
.scale-chips button {
  background: #1e293b;
  border: 1px solid #334155;
  color: #e2e8f0;
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  font-size: 0.75rem;
  cursor: pointer;
}
.modal-produce-select {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  font-size: 0.85rem;
  font-weight: 600;
}
.full-select {
  padding: 0.5rem;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
}
.modal-total-preview {
  font-size: 1.1rem;
  text-align: right;
  color: #0f172a;
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
  background: #0284c7;
  color: white;
  border: none;
  padding: 0.5rem 1rem;
  border-radius: 6px;
  font-weight: 700;
  cursor: pointer;
}
</style>
