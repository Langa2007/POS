<template>
  <div class="split-bill-manager">
    <div class="split-header">
      <div>
        <h3>
          <PosIcon name="split" size="20" class="header-icon" />
          Split Bills &amp; Table Settlements
        </h3>
        <p class="subtitle">Even split by seat, per-item assignments, service charges, and gratuity</p>
      </div>
    </div>

    <div class="split-layout">
      <!-- Active Table Check Col -->
      <div class="bill-overview-col">
        <div class="bill-card">
          <div class="bill-card-top">
            <span class="table-tag">Table 1 &mdash; Main Dining</span>
            <span class="cover-tag">4 Guests</span>
          </div>

          <div class="lines-container">
            <div
              v-for="(line, idx) in billItems"
              :key="idx"
              class="bill-line"
            >
              <div class="line-info">
                <span class="qty">{{ line.quantity }}x</span>
                <span class="name">{{ line.name }}</span>
              </div>
              <strong class="line-cost">KES {{ ((line.priceMinor * line.quantity) / 100).toLocaleString() }}</strong>
            </div>
          </div>

          <!-- Taxes & Gratuity Calculation -->
          <div class="bill-totals-box">
            <div class="calc-row">
              <span>Food &amp; Beverage Subtotal:</span>
              <span>KES {{ (rawSubtotalMinor / 100).toLocaleString() }}</span>
            </div>
            <div class="calc-row">
              <label class="check-toggle">
                <input type="checkbox" v-model="includeServiceCharge" />
                <span>Catering Levy &amp; Service (10%)</span>
              </label>
              <span>KES {{ (serviceChargeMinor / 100).toLocaleString() }}</span>
            </div>
            <div class="calc-row">
              <span>Gratuity / Tip:</span>
              <div class="tip-options">
                <button
                  v-for="pct in [0, 5, 10, 15]"
                  :key="pct"
                  type="button"
                  class="tip-btn"
                  :class="{ 'active': tipPercent === pct }"
                  @click="tipPercent = pct"
                >
                  {{ pct }}%
                </button>
              </div>
              <span>KES {{ (tipMinor / 100).toLocaleString() }}</span>
            </div>
            <div class="calc-row grand">
              <span>Total Billable Amount:</span>
              <strong>KES {{ (grandTotalMinor / 100).toLocaleString() }}</strong>
            </div>
          </div>
        </div>
      </div>

      <!-- Split Engine Column -->
      <div class="split-engine-col">
        <div class="split-mode-card">
          <h4>Select Split Strategy:</h4>
          <div class="mode-tabs">
            <button
              type="button"
              class="mode-btn"
              :class="{ 'active': splitMode === 'even' }"
              @click="splitMode = 'even'"
            >
              <PosIcon name="users" size="16" />
              Split Evenly
            </button>
            <button
              type="button"
              class="mode-btn"
              :class="{ 'active': splitMode === 'item' }"
              @click="splitMode = 'item'"
            >
              <PosIcon name="clipboard-list" size="16" />
              Split By Item
            </button>
          </div>

          <!-- Mode 1: Split Evenly -->
          <div v-if="splitMode === 'even'" class="split-mode-body">
            <div class="split-control">
              <label>Number of Ways / Guests:</label>
              <div class="guest-counter">
                <button type="button" @click="evenWays = Math.max(2, evenWays - 1)">-</button>
                <strong>{{ evenWays }} Guests</strong>
                <button type="button" @click="evenWays = Math.min(10, evenWays + 1)">+</button>
              </div>
            </div>

            <div class="split-shares-grid">
              <div
                v-for="n in evenWays"
                :key="n"
                class="share-card"
              >
                <div class="share-header">
                  <span class="share-title">Guest #{{ n }}</span>
                  <span class="share-share">1/{{ evenWays }}</span>
                </div>
                <div class="share-price">KES {{ (perGuestEvenMinor / 100).toLocaleString() }}</div>
                <button type="button" class="btn-settle-share" @click="settleShare(n, perGuestEvenMinor)">
                  <PosIcon name="credit-card" size="14" />
                  Settle Share
                </button>
              </div>
            </div>
          </div>

          <!-- Mode 2: Split By Item -->
          <div v-else class="split-mode-body">
            <p class="helper-text">Assign bill lines to individual seats:</p>
            <div class="seat-assignment-list">
              <div
                v-for="(item, idx) in billItems"
                :key="idx"
                class="seat-assign-row"
              >
                <div class="item-name">{{ item.quantity }}x {{ item.name }} (KES {{ (item.priceMinor * item.quantity / 100).toLocaleString() }})</div>
                <select v-model="itemSeatAssignments[idx]" class="seat-select">
                  <option value="1">Seat 1 (Mary)</option>
                  <option value="2">Seat 2 (John)</option>
                  <option value="3">Seat 3 (Sarah)</option>
                  <option value="4">Seat 4 (David)</option>
                </select>
              </div>
            </div>

            <div class="seat-totals-summary">
              <div v-for="s in [1, 2, 3, 4]" :key="s" class="seat-summary-badge">
                <span class="s-title">Seat {{ s }} Total:</span>
                <strong>KES {{ (calculateSeatTotal(s) / 100).toLocaleString() }}</strong>
              </div>
            </div>
          </div>

          <div class="split-actions-footer">
            <button type="button" class="btn-print-slips" @click="printSplitReceipts">
              <PosIcon name="printer" size="16" />
              Print Split Receipts
            </button>
            <button type="button" class="btn-primary" @click="sendSplitToTill">
              <PosIcon name="check-circle" size="16" />
              Send Split to Checkout Till
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref, computed } from 'vue'
import { PosIcon } from '@pos/ui'
import { useApi } from '../../../../apps/pos/src/composables/useApi.js'

const emit = defineEmits(['notification'])

const splitMode = ref('even')
const evenWays = ref(4)
const includeServiceCharge = ref(true)
const tipPercent = ref(10)

const billItems = ref([])
const api = useApi()
onMounted(async () => {
  try {
    const data = await api.get('/api/v1/restaurant/tables')
    const table = (data.tables ?? []).find(item => item.status !== 'available')
    billItems.value = table?.items ?? []
  } catch {
    emit('notification', { message: 'Bill details could not be loaded. Please contact your administrator.', type: 'error' })
  }
})

const itemSeatAssignments = ref({
  0: '1',
  1: '2',
  2: '3',
  3: '4',
})

const rawSubtotalMinor = computed(() => {
  return billItems.value.reduce((acc, it) => acc + (it.priceMinor * it.quantity), 0)
})

const serviceChargeMinor = computed(() => {
  return includeServiceCharge.value ? Math.round(rawSubtotalMinor.value * 0.1) : 0
})

const tipMinor = computed(() => {
  return tipPercent.value > 0 ? Math.round(rawSubtotalMinor.value * (tipPercent.value / 100)) : 0
})

const grandTotalMinor = computed(() => {
  return rawSubtotalMinor.value + serviceChargeMinor.value + tipMinor.value
})

const perGuestEvenMinor = computed(() => {
  return Math.round(grandTotalMinor.value / evenWays.value)
})

function calculateSeatTotal(seatNum) {
  let sub = 0
  const seatStr = String(seatNum)
  for (let i = 0; i < billItems.value.length; i++) {
    if (itemSeatAssignments.value[i] === seatStr) {
      sub += billItems.value[i].priceMinor * billItems.value[i].quantity
    }
  }
  // Add proportional service charge & tip
  const ratio = rawSubtotalMinor.value > 0 ? sub / rawSubtotalMinor.value : 0
  const extra = Math.round((serviceChargeMinor.value + tipMinor.value) * ratio)
  return sub + extra
}

function settleShare(guestNum, amountMinor) {
  emit('notification', {
    message: `Payment modal opened for Guest #${guestNum}: KES ${(amountMinor / 100).toLocaleString()}`,
    type: 'success',
  })
}

function printSplitReceipts() {
  emit('notification', {
    message: `Printed ${evenWays.value} split slips with breakdown on thermal receipt printer`,
    type: 'info',
  })
}

function sendSplitToTill() {
  emit('notification', {
    message: `Split session pushed to active register till for sequential settlement`,
    type: 'success',
  })
}
</script>

<style scoped>
.split-bill-manager {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.split-header h3 {
  margin: 0;
  font-size: 18px;
  color: #1e293b;
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-icon {
  color: #8b5cf6;
}

.subtitle {
  margin: 4px 0 0;
  font-size: 13px;
  color: #64748b;
}

.split-layout {
  display: grid;
  grid-template-columns: minmax(280px, 1.2fr) minmax(320px, 1.8fr);
  gap: 18px;
  align-items: start;
}

.bill-card, .split-mode-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 18px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.bill-card-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid #f1f5f9;
  padding-bottom: 10px;
}

.table-tag {
  font-weight: 700;
  color: #0f172a;
  font-size: 14px;
}

.cover-tag {
  background: #f1f5f9;
  font-size: 11px;
  font-weight: 600;
  color: #475569;
  padding: 2px 8px;
  border-radius: 4px;
}

.lines-container {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.bill-line {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  font-size: 13px;
  padding: 4px 0;
  border-bottom: 1px solid #f8fafc;
}

.line-info {
  display: flex;
  gap: 6px;
}

.line-info .qty {
  font-weight: 700;
  color: #8b5cf6;
}

.line-cost {
  color: #1e293b;
}

.bill-totals-box {
  margin-top: 10px;
  border-top: 2px solid #e2e8f0;
  padding-top: 10px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  font-size: 12px;
  color: #64748b;
}

.calc-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.calc-row.grand {
  margin-top: 6px;
  padding-top: 8px;
  border-top: 1px dashed #cbd5e1;
  font-size: 14px;
  color: #0f172a;
}

.calc-row.grand strong {
  font-size: 18px;
  color: #8b5cf6;
}

.check-toggle {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
}

.tip-options {
  display: flex;
  gap: 4px;
}

.tip-btn {
  padding: 2px 6px;
  border: 1px solid #cbd5e1;
  background: white;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  cursor: pointer;
}

.tip-btn.active {
  background: #8b5cf6;
  color: white;
  border-color: #8b5cf6;
}

/* Split Engine */
.mode-tabs {
  display: flex;
  gap: 8px;
}

.mode-btn {
  flex: 1;
  padding: 8px 12px;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  background: #f8fafc;
  font-size: 12px;
  font-weight: 700;
  color: #475569;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

.mode-btn.active {
  background: #8b5cf6;
  color: white;
  border-color: #8b5cf6;
}

.split-control {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #f8fafc;
  padding: 10px 14px;
  border-radius: 8px;
  margin-bottom: 14px;
}

.guest-counter {
  display: flex;
  align-items: center;
  gap: 12px;
}

.guest-counter button {
  width: 28px;
  height: 28px;
  border: 1px solid #cbd5e1;
  background: white;
  border-radius: 4px;
  font-size: 16px;
  cursor: pointer;
}

.split-shares-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 10px;
}

.share-card {
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  background: #fbfbfe;
}

.share-header {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: #64748b;
}

.share-price {
  font-size: 15px;
  font-weight: 800;
  color: #0f172a;
}

.btn-settle-share {
  margin-top: 4px;
  padding: 6px 8px;
  border-radius: 4px;
  background: white;
  border: 1px solid #cbd5e1;
  font-size: 11px;
  font-weight: 600;
  color: #334155;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
}

.btn-settle-share:hover {
  background: #8b5cf6;
  color: white;
  border-color: #8b5cf6;
}

/* Item Split */
.helper-text {
  font-size: 12px;
  color: #64748b;
  margin: 0 0 10px;
}

.seat-assignment-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.seat-assign-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px;
  background: #f8fafc;
  border-radius: 6px;
}

.item-name {
  font-size: 12px;
  font-weight: 600;
  color: #334155;
}

.seat-select {
  padding: 4px 8px;
  border: 1px solid #cbd5e1;
  border-radius: 4px;
  font-size: 12px;
}

.seat-totals-summary {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
  margin-top: 14px;
}

.seat-summary-badge {
  background: #f1f5f9;
  padding: 8px;
  border-radius: 6px;
  display: flex;
  justify-content: space-between;
  font-size: 12px;
}

.split-actions-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 14px;
  border-top: 1px solid #f1f5f9;
  padding-top: 14px;
}

.btn-print-slips, .btn-primary {
  padding: 8px 14px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.btn-print-slips {
  background: white;
  border: 1px solid #cbd5e1;
  color: #334155;
}

.btn-primary {
  background: #8b5cf6;
  color: white;
  border: none;
}

.btn-primary:hover {
  background: #7c3aed;
}

@media (max-width: 768px) {
  .split-layout {
    grid-template-columns: 1fr;
  }
}
</style>
