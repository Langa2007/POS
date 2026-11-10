<template>
  <div class="retail-view">
    <!-- Top Header -->
    <header class="view-header">
      <div class="header-main">
        <span class="shop-badge">🛒 RETAIL SUPERMARKET SUITE</span>
        <h1>Retail Counter & Back Office</h1>
        <p class="view-description">
          Integrated scale weighing, customer loyalty rewards, promotional campaigns, perishable tracking, and shelf labeling.
        </p>
      </div>

      <!-- Quick stats overview bar -->
      <div class="quick-kpis">
        <div class="kpi-card">
          <span class="kpi-label">Active Promos</span>
          <span class="kpi-val">4 Active</span>
        </div>
        <div class="kpi-card">
          <span class="kpi-label">Loyalty Club</span>
          <span class="kpi-val">1,248 Members</span>
        </div>
        <div class="kpi-card">
          <span class="kpi-label">Expiring Soon</span>
          <span class="kpi-val alert">3 Batches</span>
        </div>
      </div>
    </header>

    <!-- Notification Toast if any -->
    <div v-if="notification" class="inline-notice" :class="notification.type">
      <span>{{ notification.message }}</span>
      <button class="notice-close" @click="notification = null">✕</button>
    </div>

    <!-- Navigation Tabs -->
    <nav class="retail-tabs">
      <button
        type="button"
        class="tab-btn"
        :class="{ 'active': activeTab === 'weighed' }"
        @click="activeTab = 'weighed'"
      >
        ⚖️ Weighed Goods & Scale
      </button>
      <button
        type="button"
        class="tab-btn"
        :class="{ 'active': activeTab === 'promotions' }"
        @click="activeTab = 'promotions'"
      >
        🏷️ Promotions & Coupons
      </button>
      <button
        type="button"
        class="tab-btn"
        :class="{ 'active': activeTab === 'loyalty' }"
        @click="activeTab = 'loyalty'"
      >
        ⭐ Loyalty & Customers
      </button>
      <button
        type="button"
        class="tab-btn"
        :class="{ 'active': activeTab === 'perishables' }"
        @click="activeTab = 'perishables'"
      >
        ⏳ Perishables & Markdowns
      </button>
      <button
        type="button"
        class="tab-btn"
        :class="{ 'active': activeTab === 'shelflabels' }"
        @click="activeTab = 'shelflabels'"
      >
        🖨️ Shelf Price Labels
      </button>
    </nav>

    <!-- Tab Panels -->
    <main class="tab-content-panel">
      <!-- 1. Weighed Items & Scale -->
      <section v-if="activeTab === 'weighed'">
        <WeighedItemsManager
          @add-item="handleAddItem"
          @notification="showNotice"
        />
      </section>

      <!-- 2. Promotions & Coupons -->
      <section v-if="activeTab === 'promotions'">
        <PromotionsManager
          @apply-discount="handleApplyDiscount"
          @notification="showNotice"
        />
      </section>

      <!-- 3. Loyalty & Customers -->
      <section v-if="activeTab === 'loyalty'">
        <LoyaltyManager
          @set-customer="handleSetCustomer"
          @apply-discount="handleApplyDiscount"
          @notification="showNotice"
        />
      </section>

      <!-- 4. Perishables & Markdowns -->
      <section v-if="activeTab === 'perishables'">
        <PerishablesTracker
          @add-item="handleAddItem"
          @notification="showNotice"
        />
      </section>

      <!-- 5. Shelf Labels -->
      <section v-if="activeTab === 'shelflabels'">
        <ShelfLabelGenerator
          @notification="showNotice"
        />
      </section>
    </main>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import WeighedItemsManager from '../components/WeighedItemsManager.vue'
import PromotionsManager from '../components/PromotionsManager.vue'
import LoyaltyManager from '../components/LoyaltyManager.vue'
import PerishablesTracker from '../components/PerishablesTracker.vue'
import ShelfLabelGenerator from '../components/ShelfLabelGenerator.vue'

const router = useRouter()
const activeTab = ref('weighed')
const notification = ref(null)

function showNotice(payload) {
  notification.value = payload
  setTimeout(() => {
    if (notification.value === payload) {
      notification.value = null
    }
  }, 5000)
}

function handleAddItem(item) {
  // Save to offline queue or local storage so when user navigates to checkout, item is in cart!
  try {
    const existingCart = JSON.parse(localStorage.getItem('pos.retailCartPending') || '[]')
    existingCart.push(item)
    localStorage.setItem('pos.retailCartPending', JSON.stringify(existingCart))
  } catch (e) {
    console.error(e)
  }
  showNotice({
    message: `Added "${item.name}" (KES ${(item.price_minor / 100).toFixed(2)}) to retail cart. Go to Checkout to complete sale.`,
    type: 'success',
  })
}

function handleApplyDiscount(discount) {
  try {
    localStorage.setItem('pos.activeDiscount', JSON.stringify(discount))
  } catch (e) {
    console.error(e)
  }
  showNotice({
    message: `Promo/Discount "${discount.description}" set for active checkout!`,
    type: 'success',
  })
}

function handleSetCustomer(customer) {
  try {
    localStorage.setItem('pos.activeCustomer', JSON.stringify(customer))
  } catch (e) {
    console.error(e)
  }
  showNotice({
    message: `Attached member ${customer.name} (${customer.loyaltyCard}) to current sale.`,
    type: 'info',
  })
}
</script>

<style scoped>
.retail-view {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}
.view-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  flex-wrap: wrap;
  gap: 1rem;
  border-bottom: 1px solid #e2e8f0;
  padding-bottom: 1.25rem;
}
.shop-badge {
  font-size: 0.75rem;
  letter-spacing: 0.08em;
  font-weight: 800;
  color: #0284c7;
  background: #e0f2fe;
  padding: 0.2rem 0.6rem;
  border-radius: 4px;
  display: inline-block;
  margin-bottom: 0.5rem;
}
.header-main h1 {
  margin: 0 0 0.4rem;
  font-size: 1.85rem;
  font-weight: 800;
  color: #0f172a;
}
.view-description {
  margin: 0;
  color: #64748b;
  font-size: 0.95rem;
  max-width: 650px;
  line-height: 1.5;
}
.quick-kpis {
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
}
.kpi-card {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  padding: 0.75rem 1rem;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  min-width: 120px;
}
.kpi-label {
  font-size: 0.75rem;
  color: #64748b;
  font-weight: 600;
}
.kpi-val {
  font-size: 1.15rem;
  font-weight: 800;
  color: #0f172a;
}
.kpi-val.alert {
  color: #dc2626;
}
.inline-notice {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 1.25rem;
  border-radius: 8px;
  font-weight: 600;
  font-size: 0.9rem;
}
.inline-notice.success { background: #dcfce7; color: #166534; border: 1px solid #86efac; }
.inline-notice.warning { background: #fef3c7; color: #92400e; border: 1px solid #fde68a; }
.inline-notice.info { background: #e0f2fe; color: #0369a1; border: 1px solid #bae6fd; }
.notice-close { background: none; border: none; font-size: 1.1rem; cursor: pointer; color: inherit; }
.retail-tabs {
  display: flex;
  gap: 0.5rem;
  border-bottom: 2px solid #e2e8f0;
  overflow-x: auto;
  padding-bottom: 2px;
}
.tab-btn {
  background: none;
  border: none;
  padding: 0.75rem 1.15rem;
  font-size: 0.95rem;
  font-weight: 700;
  color: #64748b;
  cursor: pointer;
  border-radius: 8px 8px 0 0;
  transition: all 0.15s ease;
  white-space: nowrap;
}
.tab-btn:hover {
  color: #0284c7;
  background: #f1f5f9;
}
.tab-btn.active {
  color: #0284c7;
  background: #f0f9ff;
  border-bottom: 3px solid #0284c7;
}
.tab-content-panel {
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 1.5rem;
}
</style>
