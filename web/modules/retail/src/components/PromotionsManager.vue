<template>
  <div class="promotions-manager">
    <div class="manager-header">
      <div>
        <h3>🏷️ Retail Promotions & Discount Rules</h3>
        <p class="subtitle">Multi-buy bundles, percentage campaigns, and promotional voucher codes</p>
      </div>
      <button class="btn-create" @click="showCreateModal = true">
        ➕ New Campaign
      </button>
    </div>

    <!-- Voucher / Promo Code Validation Tester -->
    <div class="coupon-test-card">
      <div class="coupon-form">
        <label for="promoCodeInput">Apply Customer Promo Code / Voucher:</label>
        <div class="coupon-input-group">
          <input
            id="promoCodeInput"
            type="text"
            v-model="inputCoupon"
            placeholder="e.g. SAVE10, SUPER20, WKND200"
            class="coupon-input"
            @keydown.enter.prevent="testCoupon"
          />
          <button type="button" class="btn-primary" @click="testCoupon">
            Validate & Apply
          </button>
        </div>
      </div>
      <div v-if="couponResult" class="coupon-result" :class="{ 'valid': couponResult.valid, 'invalid': !couponResult.valid }">
        <span class="icon">{{ couponResult.valid ? '✅' : '❌' }}</span>
        <span>{{ couponResult.message }}</span>
        <button v-if="couponResult.valid" class="btn-apply-cart" @click="applyCouponToCart">
          Apply {{ couponResult.discountLabel }} to POS Sale
        </button>
      </div>
    </div>

    <!-- Active Promotions Table -->
    <div class="promotions-list">
      <h4>Active Store Promotions:</h4>
      <div class="promo-cards-grid">
        <div
          v-for="promo in activePromotions"
          :key="promo.id"
          class="promo-card"
          :class="promo.type"
        >
          <div class="promo-badge">{{ promo.badge }}</div>
          <h5 class="promo-title">{{ promo.name }}</h5>
          <p class="promo-desc">{{ promo.description }}</p>

          <div class="promo-rule-details">
            <div class="detail-row">
              <span class="label">Discount:</span>
              <span class="val highlight">{{ promo.discountText }}</span>
            </div>
            <div class="detail-row">
              <span class="label">Min Spend:</span>
              <span class="val">KES {{ promo.minSpend }}</span>
            </div>
            <div class="detail-row">
              <span class="label">Valid Until:</span>
              <span class="val">{{ promo.validUntil }}</span>
            </div>
          </div>

          <div class="promo-card-footer">
            <span class="promo-code-tag">Code: <code>{{ promo.code }}</code></span>
            <button class="btn-apply-rule" @click="applyPromoRule(promo)">
              Activate Rule
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- New Campaign Modal -->
    <div v-if="showCreateModal" class="modal-backdrop">
      <div class="modal-dialog">
        <div class="modal-header">
          <h4>Create Retail Campaign</h4>
          <button class="close-btn" @click="showCreateModal = false">✕</button>
        </div>
        <form @submit.prevent="createCampaign" class="modal-form">
          <label>Campaign Name:
            <input type="text" v-model="newPromo.name" required placeholder="e.g. End of Month Meat Special" />
          </label>
          <label>Promo Code:
            <input type="text" v-model="newPromo.code" required placeholder="e.g. MEAT20" />
          </label>
          <div class="form-row">
            <label>Promotion Type:
              <select v-model="newPromo.type">
                <option value="percent">Percentage Off</option>
                <option value="fixed">Fixed Amount Off</option>
                <option value="bogo">Buy 2 Get 1 Free</option>
                <option value="bundle">Multi-buy Bundle</option>
              </select>
            </label>
            <label>Discount Value:
              <input type="number" v-model.number="newPromo.value" min="1" required />
            </label>
          </div>
          <div class="form-row">
            <label>Minimum Cart Spend (KES):
              <input type="number" v-model.number="newPromo.minSpend" min="0" />
            </label>
            <label>Expiry Date:
              <input type="date" v-model="newPromo.validUntil" required />
            </label>
          </div>
          <label>Description:
            <textarea v-model="newPromo.description" rows="2" placeholder="Brief summary of promotion"></textarea>
          </label>
          <div class="modal-actions">
            <button type="button" class="btn-secondary" @click="showCreateModal = false">Cancel</button>
            <button type="submit" class="btn-primary">Save & Enable</button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const emit = defineEmits(['apply-discount', 'notification'])

const inputCoupon = ref('')
const couponResult = ref(null)
const showCreateModal = ref(false)

const activePromotions = ref([
  {
    id: 'p-1',
    code: 'SAVE10',
    name: 'General 10% Store Discount',
    description: '10% off entire order for store anniversary celebration.',
    type: 'percent',
    badge: '10% OFF',
    discountText: '10% off total',
    discountPercent: 10,
    minSpend: 500,
    validUntil: '2026-12-31',
  },
  {
    id: 'p-2',
    code: 'WKND200',
    name: 'Weekend Super Shopper',
    description: 'KES 200 instant cash discount on grocery carts above KES 1,500.',
    type: 'fixed',
    badge: 'KES 200 OFF',
    discountText: 'KES 200 voucher',
    discountMinor: 20000,
    minSpend: 1500,
    validUntil: '2026-11-30',
  },
  {
    id: 'p-3',
    code: 'B2G1SODA',
    name: 'Sodas Buy 2 Get 1 Free',
    description: 'Purchase 2 sodas or juices and receive the 3rd bottle free.',
    type: 'bogo',
    badge: 'BUY 2 GET 1',
    discountText: '1 free item on trio',
    minSpend: 180,
    validUntil: '2026-11-25',
  },
  {
    id: 'p-4',
    code: 'BAKERY15',
    name: 'Daily Fresh Bakery Promo',
    description: '15% discount on all bread, cakes, and pastries after 5 PM.',
    type: 'percent',
    badge: '15% OFF',
    discountText: '15% off bakery',
    discountPercent: 15,
    minSpend: 200,
    validUntil: '2026-12-15',
  },
])

const newPromo = ref({
  name: '',
  code: '',
  type: 'percent',
  value: 10,
  minSpend: 500,
  validUntil: new Date(Date.now() + 30 * 86400000).toISOString().split('T')[0],
  description: '',
})

function testCoupon() {
  const code = inputCoupon.value.trim().toUpperCase()
  if (!code) return
  const match = activePromotions.value.find((p) => p.code === code)
  if (match) {
    couponResult.value = {
      valid: true,
      promo: match,
      message: `Coupon "${match.code}" is valid! (${match.name})`,
      discountLabel: match.discountText,
    }
  } else {
    couponResult.value = {
      valid: false,
      message: `Promo code "${code}" is invalid or expired. Check capitalization.`,
    }
  }
}

function applyCouponToCart() {
  if (!couponResult.value?.valid) return
  const p = couponResult.value.promo
  emit('apply-discount', {
    code: p.code,
    description: p.name,
    percent: p.discountPercent || null,
    amountMinor: p.discountMinor || (p.discountPercent ? null : p.value * 100),
  })
  emit('notification', {
    message: `Applied promotion "${p.name}" (${p.discountText}) to active cart`,
    type: 'success',
  })
}

function applyPromoRule(promo) {
  inputCoupon.value = promo.code
  testCoupon()
  applyCouponToCart()
}

function createCampaign() {
  const promo = {
    id: `p-${Date.now()}`,
    code: newPromo.value.code.toUpperCase(),
    name: newPromo.value.name,
    description: newPromo.value.description,
    type: newPromo.value.type,
    badge: newPromo.value.type === 'percent' ? `${newPromo.value.value}% OFF` : `KES ${newPromo.value.value} OFF`,
    discountText: newPromo.value.type === 'percent' ? `${newPromo.value.value}% off` : `KES ${newPromo.value.value} voucher`,
    discountPercent: newPromo.value.type === 'percent' ? newPromo.value.value : null,
    discountMinor: newPromo.value.type === 'fixed' ? newPromo.value.value * 100 : null,
    minSpend: newPromo.value.minSpend,
    validUntil: newPromo.value.validUntil,
  }
  activePromotions.value.unshift(promo)
  showCreateModal.value = false
  emit('notification', { message: `Created campaign "${promo.name}" with code ${promo.code}`, type: 'success' })
}
</script>

<style scoped>
.promotions-manager {
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
.btn-create {
  background: #10b981;
  color: white;
  border: none;
  padding: 0.55rem 1rem;
  border-radius: 8px;
  font-weight: 700;
  cursor: pointer;
}
.btn-create:hover {
  background: #059669;
}
.coupon-test-card {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}
.coupon-form label {
  font-weight: 600;
  font-size: 0.9rem;
  color: #334155;
  display: block;
  margin-bottom: 0.35rem;
}
.coupon-input-group {
  display: flex;
  gap: 0.5rem;
  max-width: 480px;
}
.coupon-input {
  flex: 1;
  padding: 0.6rem 0.85rem;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  font-weight: 700;
  text-transform: uppercase;
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
.coupon-result {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  border-radius: 8px;
  font-size: 0.9rem;
  font-weight: 600;
}
.coupon-result.valid {
  background: #dcfce7;
  color: #166534;
  border: 1px solid #86efac;
}
.coupon-result.invalid {
  background: #fee2e2;
  color: #991b1b;
  border: 1px solid #fca5a5;
}
.btn-apply-cart {
  margin-left: auto;
  background: #166534;
  color: white;
  border: none;
  padding: 0.4rem 0.85rem;
  border-radius: 6px;
  font-weight: 700;
  cursor: pointer;
}
.promotions-list h4 {
  margin: 0 0 0.85rem;
  font-size: 1rem;
}
.promo-cards-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 1rem;
}
.promo-card {
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 1.15rem;
  display: flex;
  flex-direction: column;
  position: relative;
  transition: all 0.15s ease;
}
.promo-card:hover {
  border-color: #0284c7;
  box-shadow: 0 4px 12px rgba(0,0,0,0.06);
}
.promo-badge {
  align-self: flex-start;
  font-size: 0.75rem;
  font-weight: 800;
  padding: 0.2rem 0.5rem;
  border-radius: 6px;
  background: #e0f2fe;
  color: #0369a1;
  margin-bottom: 0.5rem;
}
.promo-title {
  margin: 0 0 0.35rem;
  font-size: 1.05rem;
  font-weight: 700;
  color: #0f172a;
}
.promo-desc {
  margin: 0 0 0.85rem;
  font-size: 0.85rem;
  color: #64748b;
  line-height: 1.4;
  flex-grow: 1;
}
.promo-rule-details {
  border-top: 1px dashed #e2e8f0;
  padding: 0.65rem 0;
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
  font-size: 0.8rem;
}
.detail-row {
  display: flex;
  justify-content: space-between;
}
.detail-row .label {
  color: #64748b;
}
.detail-row .val {
  font-weight: 600;
  color: #1e293b;
}
.detail-row .highlight {
  color: #0284c7;
  font-weight: 700;
}
.promo-card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 0.5rem;
  padding-top: 0.65rem;
  border-top: 1px solid #f1f5f9;
}
.promo-code-tag code {
  font-weight: 800;
  color: #0f172a;
  background: #f1f5f9;
  padding: 0.15rem 0.4rem;
  border-radius: 4px;
}
.btn-apply-rule {
  background: #f1f5f9;
  border: 1px solid #cbd5e1;
  color: #1e293b;
  padding: 0.35rem 0.75rem;
  border-radius: 6px;
  font-weight: 700;
  cursor: pointer;
  font-size: 0.8rem;
}
.btn-apply-rule:hover {
  background: #0284c7;
  color: #fff;
  border-color: #0284c7;
}
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.5);
  display: grid;
  place-items: center;
  z-index: 1000;
}
.modal-dialog {
  background: white;
  width: min(520px, 94vw);
  border-radius: 12px;
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.modal-header h4 {
  margin: 0;
  font-size: 1.2rem;
}
.close-btn {
  background: none;
  border: none;
  font-size: 1.25rem;
  cursor: pointer;
}
.modal-form {
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
}
.modal-form label {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  font-size: 0.85rem;
  font-weight: 600;
  color: #334155;
}
.modal-form input, .modal-form select, .modal-form textarea {
  padding: 0.5rem 0.75rem;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-family: inherit;
}
.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem;
}
.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  margin-top: 0.5rem;
}
.btn-secondary {
  background: #f1f5f9;
  border: 1px solid #cbd5e1;
  padding: 0.5rem 1rem;
  border-radius: 6px;
  font-weight: 600;
  cursor: pointer;
}
</style>
