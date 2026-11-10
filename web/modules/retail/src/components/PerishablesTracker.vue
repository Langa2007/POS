<template>
  <div class="perishables-tracker">
    <div class="manager-header">
      <div>
        <h3>⏳ Perishables & Clearance Markdown Tracker</h3>
        <p class="subtitle">Monitor shelf-life, prevent shrinkage, and apply quick clearance discounts</p>
      </div>
      <div class="summary-stats">
        <span class="stat-pill critical">🚨 {{ criticalCount }} Critical (≤ 48h)</span>
        <span class="stat-pill warning">⚠️ {{ warningCount }} Warning (≤ 5d)</span>
        <span class="stat-pill good">✅ {{ safeCount }} Safe</span>
      </div>
    </div>

    <!-- Perishable Items Table -->
    <div class="table-container">
      <table class="perishables-table">
        <thead>
          <tr>
            <th>Product Name</th>
            <th>Batch / Lot</th>
            <th>Category</th>
            <th>Expiry Date</th>
            <th>Status</th>
            <th>Stock</th>
            <th>Current Price</th>
            <th>Quick Clearance Markdown</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in perishableStock" :key="item.id" :class="item.status">
            <td>
              <div class="product-title">
                <span class="icon">{{ item.icon }}</span>
                <div>
                  <strong>{{ item.name }}</strong>
                  <small>SKU: {{ item.sku }}</small>
                </div>
              </div>
            </td>
            <td><code>{{ item.batch }}</code></td>
            <td>{{ item.category }}</td>
            <td>
              <span class="expiry-date">{{ item.expiryDate }}</span>
              <small class="days-remaining">({{ item.daysLeft }} days left)</small>
            </td>
            <td>
              <span class="status-badge" :class="item.status">
                {{ item.statusText }}
              </span>
            </td>
            <td><strong>{{ item.stock }}</strong> units</td>
            <td>
              <span v-if="item.discountPercent" class="strike-price">KES {{ (item.originalPriceMinor / 100).toFixed(2) }}</span>
              <span class="active-price">KES {{ (item.priceMinor / 100).toFixed(2) }}</span>
            </td>
            <td>
              <div class="markdown-actions">
                <button
                  type="button"
                  class="btn-markdown"
                  :class="{ 'active': item.discountPercent === 20 }"
                  @click="applyMarkdown(item, 20)"
                >
                  -20%
                </button>
                <button
                  type="button"
                  class="btn-markdown"
                  :class="{ 'active': item.discountPercent === 35 }"
                  @click="applyMarkdown(item, 35)"
                >
                  -35%
                </button>
                <button
                  type="button"
                  class="btn-markdown danger"
                  :class="{ 'active': item.discountPercent === 50 }"
                  @click="applyMarkdown(item, 50)"
                >
                  -50%
                </button>
                <button
                  type="button"
                  class="btn-push-cart"
                  @click="addMarkdownToCart(item)"
                >
                  ➕ Sell
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Clearance Sticker Preview Modal/Drawer -->
    <div v-if="selectedMarkdownSticker" class="sticker-preview-box">
      <div class="sticker-header">
        <span>🖨️ Clearance Markdown Sticker Preview:</span>
        <button class="close-btn" @click="selectedMarkdownSticker = null">✕</button>
      </div>
      <div class="sticker-tag">
        <div class="sticker-banner">🚨 CLEARANCE REDUCED TO CLEAR 🚨</div>
        <div class="sticker-name">{{ selectedMarkdownSticker.name }}</div>
        <div class="sticker-prices">
          <span class="was">WAS: KES {{ (selectedMarkdownSticker.originalPriceMinor / 100).toFixed(2) }}</span>
          <span class="now">NOW: KES {{ (selectedMarkdownSticker.priceMinor / 100).toFixed(2) }}</span>
        </div>
        <div class="sticker-meta">
          <span>Batch: {{ selectedMarkdownSticker.batch }}</span>
          <span>Best Before: {{ selectedMarkdownSticker.expiryDate }}</span>
        </div>
      </div>
      <button class="btn-print-sticker" @click="printSticker">
        🖨️ Print Clearance Shelf Barcode
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const emit = defineEmits(['add-item', 'notification'])

const perishableStock = ref([
  {
    id: 'per-1',
    name: 'Fresh Whole Milk 1L',
    sku: 'MILK-1L',
    batch: 'LOT-2026-1108',
    category: 'Dairy',
    icon: '🥛',
    expiryDate: new Date(Date.now() + 1 * 86400000).toISOString().split('T')[0],
    daysLeft: 1,
    status: 'critical',
    statusText: 'Expires Tomorrow',
    stock: 14,
    originalPriceMinor: 11000,
    priceMinor: 11000,
    discountPercent: 0,
  },
  {
    id: 'per-2',
    name: 'Country Loaf White Bread',
    sku: 'BREAD-400G',
    batch: 'LOT-2026-1109',
    category: 'Bakery',
    icon: '🍞',
    expiryDate: new Date(Date.now() + 2 * 86400000).toISOString().split('T')[0],
    daysLeft: 2,
    status: 'critical',
    statusText: 'Expires in 48h',
    stock: 22,
    originalPriceMinor: 6500,
    priceMinor: 6500,
    discountPercent: 0,
  },
  {
    id: 'per-3',
    name: 'Natural Greek Yogurt 500g',
    sku: 'YOG-500G',
    batch: 'LOT-2026-1104',
    category: 'Dairy',
    icon: '🥣',
    expiryDate: new Date(Date.now() + 4 * 86400000).toISOString().split('T')[0],
    daysLeft: 4,
    status: 'warning',
    statusText: 'Expires in 4 days',
    stock: 8,
    originalPriceMinor: 19500,
    priceMinor: 19500,
    discountPercent: 0,
  },
  {
    id: 'per-4',
    name: 'Smoked Beef Sausages 500g',
    sku: 'SAUS-500G',
    batch: 'LOT-2026-1102',
    category: 'Meat/Chilled',
    icon: '🌭',
    expiryDate: new Date(Date.now() + 12 * 86400000).toISOString().split('T')[0],
    daysLeft: 12,
    status: 'good',
    statusText: 'Fresh (12d)',
    stock: 35,
    originalPriceMinor: 38000,
    priceMinor: 38000,
    discountPercent: 0,
  },
])

const selectedMarkdownSticker = ref(null)

const criticalCount = computed(() => perishableStock.value.filter((i) => i.status === 'critical').length)
const warningCount = computed(() => perishableStock.value.filter((i) => i.status === 'warning').length)
const safeCount = computed(() => perishableStock.value.filter((i) => i.status === 'good').length)

function applyMarkdown(item, percent) {
  item.discountPercent = percent
  item.priceMinor = Math.round(item.originalPriceMinor * (1 - percent / 100))
  selectedMarkdownSticker.value = item
  emit('notification', {
    message: `Marked down ${item.name} by ${percent}% (Now KES ${(item.priceMinor / 100).toFixed(2)})`,
    type: 'warning',
  })
}

function addMarkdownToCart(item) {
  emit('add-item', {
    id: `clearance-${item.id}-${Date.now()}`,
    name: `${item.name} [CLEARANCE ${item.discountPercent ? `-${item.discountPercent}%` : ''}]`,
    sku: item.sku,
    price_minor: item.priceMinor,
    unit_price_minor: item.priceMinor,
    quantity: 1,
    kind: 'stocked',
  })
  emit('notification', {
    message: `Added clearance ${item.name} (KES ${(item.priceMinor / 100).toFixed(2)}) to POS cart`,
    type: 'success',
  })
}

function printSticker() {
  window.print()
}
</script>

<style scoped>
.perishables-tracker {
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
.summary-stats {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}
.stat-pill {
  font-size: 0.8rem;
  font-weight: 700;
  padding: 0.35rem 0.75rem;
  border-radius: 999px;
}
.stat-pill.critical { background: #fee2e2; color: #991b1b; }
.stat-pill.warning { background: #fef3c7; color: #92400e; }
.stat-pill.good { background: #dcfce7; color: #166534; }
.table-container {
  overflow-x: auto;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  background: white;
}
.perishables-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
  font-size: 0.85rem;
}
.perishables-table th, .perishables-table td {
  padding: 0.85rem 1rem;
  border-bottom: 1px solid #f1f5f9;
}
.perishables-table th {
  background: #f8fafc;
  font-weight: 700;
  color: #475569;
}
.product-title {
  display: flex;
  align-items: center;
  gap: 0.65rem;
}
.product-title .icon {
  font-size: 1.5rem;
}
.product-title small {
  display: block;
  color: #64748b;
  font-size: 0.75rem;
}
.days-remaining {
  display: block;
  font-size: 0.75rem;
  color: #64748b;
}
.status-badge {
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 700;
}
.status-badge.critical { background: #fee2e2; color: #991b1b; }
.status-badge.warning { background: #fef3c7; color: #92400e; }
.status-badge.good { background: #dcfce7; color: #166534; }
.strike-price {
  text-decoration: line-through;
  color: #94a3b8;
  font-size: 0.75rem;
  display: block;
}
.active-price {
  font-weight: 800;
  color: #0f172a;
}
.markdown-actions {
  display: flex;
  gap: 0.35rem;
  align-items: center;
}
.btn-markdown {
  background: #f1f5f9;
  border: 1px solid #cbd5e1;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 700;
  cursor: pointer;
}
.btn-markdown:hover {
  background: #e2e8f0;
}
.btn-markdown.active {
  background: #f59e0b;
  color: #fff;
  border-color: #f59e0b;
}
.btn-markdown.danger.active {
  background: #ef4444;
  color: #fff;
  border-color: #ef4444;
}
.btn-push-cart {
  background: #0284c7;
  color: white;
  border: none;
  padding: 0.3rem 0.65rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 700;
  cursor: pointer;
  margin-left: 0.35rem;
}
.btn-push-cart:hover {
  background: #0369a1;
}
.sticker-preview-box {
  background: #fff;
  border: 2px dashed #f59e0b;
  border-radius: 12px;
  padding: 1rem 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  max-width: 480px;
}
.sticker-header {
  display: flex;
  justify-content: space-between;
  font-weight: 700;
}
.close-btn {
  background: none;
  border: none;
  cursor: pointer;
}
.sticker-tag {
  background: #fffbeb;
  border: 2px solid #f59e0b;
  padding: 1rem;
  border-radius: 8px;
  text-align: center;
}
.sticker-banner {
  background: #ef4444;
  color: #fff;
  font-weight: 900;
  font-size: 0.8rem;
  letter-spacing: 0.05em;
  padding: 0.2rem;
  margin-bottom: 0.5rem;
}
.sticker-name {
  font-size: 1.15rem;
  font-weight: 800;
  color: #0f172a;
}
.sticker-prices {
  display: flex;
  justify-content: center;
  align-items: baseline;
  gap: 1rem;
  margin: 0.5rem 0;
}
.sticker-prices .was {
  text-decoration: line-through;
  color: #64748b;
  font-size: 0.9rem;
}
.sticker-prices .now {
  font-size: 1.75rem;
  font-weight: 900;
  color: #b91c1c;
}
.sticker-meta {
  display: flex;
  justify-content: space-between;
  font-size: 0.75rem;
  color: #64748b;
  border-top: 1px dashed #cbd5e1;
  padding-top: 0.5rem;
}
.btn-print-sticker {
  background: #0f172a;
  color: #fff;
  border: none;
  padding: 0.6rem;
  border-radius: 6px;
  font-weight: 700;
  cursor: pointer;
}
</style>
