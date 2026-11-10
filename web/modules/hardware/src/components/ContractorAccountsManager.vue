<template>
  <div class="contractor-accounts-manager">
    <div class="contractor-header">
      <div>
        <h3>
          <PosIcon name="hard-hat" size="20" class="header-icon" />
          Contractor Trade Accounts &amp; Credit
        </h3>
        <p class="subtitle">Manage builder credit lines, 30/60/90-day aging, trade discounts, and account checkout authorization</p>
      </div>

      <button type="button" class="btn-new-contractor" @click="openEnrollModal = true">
        <PosIcon name="plus" size="16" />
        New Contractor Account
      </button>
    </div>

    <!-- Accounts Grid / Table -->
    <div class="accounts-grid">
      <div
        v-for="acc in contractors"
        :key="acc.id"
        class="contractor-card"
        :class="[acc.status, { 'selected': selectedContractor?.id === acc.id }]"
        @click="selectedContractor = acc"
      >
        <div class="card-top">
          <div>
            <span class="company-name">{{ acc.companyName }}</span>
            <span class="reg-number">PIN: {{ acc.kraPin }} &bull; Account #{{ acc.accountNo }}</span>
          </div>
          <span class="status-badge" :class="acc.status">{{ formatStatus(acc.status) }}</span>
        </div>

        <div class="credit-bar-wrap">
          <div class="bar-labels">
            <span>Used: KES {{ (acc.balanceMinor / 100).toLocaleString() }}</span>
            <span>Limit: KES {{ (acc.limitMinor / 100).toLocaleString() }}</span>
          </div>
          <div class="progress-track">
            <div
              class="progress-fill"
              :class="getProgressColor(acc.balanceMinor, acc.limitMinor)"
              :style="{ width: `${Math.min(100, Math.round((acc.balanceMinor / acc.limitMinor) * 100))}%` }"
            ></div>
          </div>
          <span class="avail-text">
            Available Credit: <strong>KES {{ (Math.max(0, acc.limitMinor - acc.balanceMinor) / 100).toLocaleString() }}</strong>
          </span>
        </div>

        <!-- Aging Breakdown Strip -->
        <div class="aging-strip">
          <div class="age-col">
            <span class="a-lbl">Current</span>
            <strong>KES {{ (acc.aging.current / 100).toLocaleString() }}</strong>
          </div>
          <div class="age-col">
            <span class="a-lbl">30 Days</span>
            <strong>KES {{ (acc.aging.d30 / 100).toLocaleString() }}</strong>
          </div>
          <div class="age-col" :class="{ 'alert': acc.aging.d60 > 0 }">
            <span class="a-lbl">60 Days</span>
            <strong>KES {{ (acc.aging.d60 / 100).toLocaleString() }}</strong>
          </div>
          <div class="age-col" :class="{ 'danger': acc.aging.d90 > 0 }">
            <span class="a-lbl">90+ Days</span>
            <strong>KES {{ (acc.aging.d90 / 100).toLocaleString() }}</strong>
          </div>
        </div>

        <div class="card-footer">
          <span class="trade-tier">Trade Discount: <strong>{{ acc.tradeDiscountPercent }}% Off</strong></span>
          <button
            type="button"
            class="btn-charge"
            :disabled="acc.status === 'suspended'"
            @click.stop="openChargeModal(acc)"
          >
            <PosIcon name="credit-card" size="14" />
            Charge to Account
          </button>
        </div>
      </div>
    </div>

    <!-- Charge Authorization Modal -->
    <div v-if="authModalOpen" class="modal-backdrop" @click.self="authModalOpen = false">
      <div class="modal-card">
        <div class="modal-header">
          <h4>Authorize Trade Credit Charge</h4>
          <button class="close-btn" @click="authModalOpen = false"><PosIcon name="x" size="16" /></button>
        </div>
        <p class="modal-desc">
          Charge customer sale to <strong>{{ targetedContractor?.companyName }}</strong>'s credit facility.
        </p>

        <div class="charge-summary-box">
          <div class="cs-row">
            <span>Available Facility:</span>
            <strong>KES {{ (Math.max(0, targetedContractor.limitMinor - targetedContractor.balanceMinor) / 100).toLocaleString() }}</strong>
          </div>
          <div class="cs-row">
            <span>Applicable Trade Discount:</span>
            <strong class="trade-save">{{ targetedContractor.tradeDiscountPercent }}% Trade Pricing Applied</strong>
          </div>
        </div>

        <div class="form-row">
          <label>Authorized Site Representative Name:</label>
          <input type="text" v-model="authForm.repName" placeholder="e.g. Foreman Charles Mwangi" class="form-input" />
        </div>
        <div class="form-row">
          <label>National ID / Passport Number:</label>
          <input type="text" v-model="authForm.repId" placeholder="e.g. ID-24819201" class="form-input" />
        </div>
        <div class="form-row">
          <label>Site Delivery Address / LPO Number:</label>
          <input type="text" v-model="authForm.lpoNumber" placeholder="e.g. LPO-APEX-2026-88" class="form-input" />
        </div>
        <div class="form-row">
          <label>Store Manager Approval PIN (Use: 9988):</label>
          <input type="password" v-model="authForm.pin" placeholder="Enter PIN" class="form-input" />
        </div>

        <div class="modal-actions">
          <button type="button" class="btn-secondary" @click="authModalOpen = false">Cancel</button>
          <button type="button" class="btn-primary" @click="authorizeChargeAction">
            <PosIcon name="check-circle" size="16" />
            Authorize Credit Charge
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { PosIcon } from '@pos/ui'

const emit = defineEmits(['notification'])

const selectedContractor = ref(null)
const authModalOpen = ref(false)
const targetedContractor = ref(null)
const openEnrollModal = ref(false)

const authForm = ref({
  repName: 'Foreman Charles Mwangi',
  repId: 'ID-24819201',
  lpoNumber: 'LPO-APEX-2026-88',
  pin: '',
})

const contractors = ref([
  {
    id: 'cnt-1',
    companyName: 'Apex Building Contractors Ltd',
    kraPin: 'P051284920K',
    accountNo: 'TRD-1001',
    limitMinor: 100000000, // KES 1,000,000
    balanceMinor: 42000000, // KES 420,000
    tradeDiscountPercent: 8,
    status: 'active',
    aging: {
      current: 25000000,
      d30: 17000000,
      d60: 0,
      d90: 0,
    },
  },
  {
    id: 'cnt-2',
    companyName: 'Fundi Juma Plumbing & Civil Works',
    kraPin: 'A009182341M',
    accountNo: 'TRD-1002',
    limitMinor: 25000000, // KES 250,000
    balanceMinor: 18500000, // KES 185,000
    tradeDiscountPercent: 5,
    status: 'warning',
    aging: {
      current: 5000000,
      d30: 8500000,
      d60: 5000000,
      d90: 0,
    },
  },
  {
    id: 'cnt-3',
    companyName: 'Coast Highrise Civil Works',
    kraPin: 'P059182741Z',
    accountNo: 'TRD-1003',
    limitMinor: 80000000, // KES 800,000
    balanceMinor: 12000000, // KES 120,000
    tradeDiscountPercent: 10,
    status: 'active',
    aging: {
      current: 12000000,
      d30: 0,
      d60: 0,
      d90: 0,
    },
  },
  {
    id: 'cnt-4',
    companyName: 'Skyline Interiors & Drywall',
    kraPin: 'P051184910Y',
    accountNo: 'TRD-1004',
    limitMinor: 50000000, // KES 500,000
    balanceMinor: 49500000, // KES 495,000
    tradeDiscountPercent: 5,
    status: 'suspended',
    aging: {
      current: 9500000,
      d30: 15000000,
      d60: 10000000,
      d90: 15000000,
    },
  },
])

function formatStatus(status) {
  switch (status) {
    case 'active': return 'Good Standing'
    case 'warning': return 'Review (Over 60%)'
    case 'suspended': return 'Suspended (Overdue)'
    default: return status
  }
}

function getProgressColor(balance, limit) {
  const pct = (balance / limit) * 100
  if (pct > 90) return 'danger'
  if (pct > 60) return 'warning'
  return 'good'
}

function openChargeModal(contractor) {
  targetedContractor.value = contractor
  authModalOpen.value = true
  authForm.value.pin = ''
}

function authorizeChargeAction() {
  if (authForm.value.pin !== '9988') {
    emit('notification', {
      message: 'Invalid Manager Approval PIN (Correct PIN for testing is 9988)',
      type: 'error',
    })
    return
  }
  if (!authForm.value.repName || !authForm.value.repId) {
    emit('notification', {
      message: 'Representative Name and National ID are mandatory for trade credit charges',
      type: 'error',
    })
    return
  }

  // Push trade discount event
  window.dispatchEvent(new CustomEvent('pos:apply-discount', {
    detail: {
      type: 'percent',
      value: targetedContractor.value.tradeDiscountPercent,
      description: `Contractor Trade Pricing (${targetedContractor.value.tradeDiscountPercent}% - ${targetedContractor.value.companyName})`,
    }
  }))

  authModalOpen.value = false

  emit('notification', {
    message: `Trade account charge authorized for ${targetedContractor.value.companyName}. Applied ${targetedContractor.value.tradeDiscountPercent}% discount.`,
    type: 'success',
  })
}
</script>

<style scoped>
.contractor-accounts-manager {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.contractor-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  flex-wrap: wrap;
  gap: 12px;
}

.contractor-header h3 {
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

.btn-new-contractor {
  background: white;
  border: 1px solid #cbd5e1;
  padding: 8px 14px;
  border-radius: 8px;
  font-size: 12px;
  font-weight: 600;
  color: #334155;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.btn-new-contractor:hover {
  background: #f1f5f9;
}

.accounts-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(310px, 1fr));
  gap: 16px;
}

.contractor-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 18px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.contractor-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
}

.contractor-card.selected {
  outline: 2px solid #0284c7;
}

.card-top {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 8px;
}

.company-name {
  display: block;
  font-weight: 700;
  font-size: 15px;
  color: #0f172a;
}

.reg-number {
  font-size: 11px;
  color: #64748b;
}

.status-badge {
  font-size: 10px;
  font-weight: 700;
  padding: 3px 8px;
  border-radius: 6px;
  text-transform: uppercase;
}

.status-badge.active { background: #dcfce7; color: #15803d; }
.status-badge.warning { background: #fef3c7; color: #b45309; }
.status-badge.suspended { background: #fee2e2; color: #b91c1c; }

.credit-bar-wrap {
  display: flex;
  flex-direction: column;
  gap: 4px;
  background: #f8fafc;
  padding: 10px 12px;
  border-radius: 8px;
}

.bar-labels {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: #475569;
}

.progress-track {
  height: 8px;
  background: #e2e8f0;
  border-radius: 4px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  border-radius: 4px;
  transition: width 0.3s ease;
}

.progress-fill.good { background: #16a34a; }
.progress-fill.warning { background: #f59e0b; }
.progress-fill.danger { background: #ef4444; }

.avail-text {
  font-size: 11px;
  color: #64748b;
  margin-top: 2px;
}

.avail-text strong {
  color: #0f172a;
}

.aging-strip {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 6px;
  border-top: 1px dashed #e2e8f0;
  border-bottom: 1px dashed #e2e8f0;
  padding: 8px 0;
}

.age-col {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.a-lbl {
  font-size: 10px;
  color: #64748b;
  text-transform: uppercase;
}

.age-col strong {
  font-size: 11px;
  color: #0f172a;
}

.age-col.alert strong { color: #d97706; }
.age-col.danger strong { color: #dc2626; }

.card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 4px;
}

.trade-tier {
  font-size: 12px;
  color: #166534;
}

.btn-charge {
  padding: 8px 12px;
  border-radius: 6px;
  background: #0284c7;
  color: white;
  border: none;
  font-size: 11px;
  font-weight: 700;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.btn-charge:hover {
  background: #0369a1;
}

.btn-charge:disabled {
  background: #cbd5e1;
  cursor: not-allowed;
}

/* Modal */
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 999;
}

.modal-card {
  background: white;
  border-radius: 12px;
  padding: 22px;
  width: 90%;
  max-width: 480px;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.modal-desc {
  font-size: 13px;
  color: #64748b;
  margin: 0 0 14px;
}

.charge-summary-box {
  background: #f0fdf4;
  border: 1px solid #bbf7d0;
  border-radius: 8px;
  padding: 10px 14px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  font-size: 12px;
  margin-bottom: 14px;
}

.cs-row {
  display: flex;
  justify-content: space-between;
}

.trade-save {
  color: #15803d;
}

.form-row {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-bottom: 12px;
}

.form-row label {
  font-size: 12px;
  font-weight: 600;
  color: #475569;
}

.form-input {
  padding: 8px 10px;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 13px;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 16px;
}

.btn-secondary, .btn-primary {
  padding: 8px 14px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.btn-secondary { background: white; border: 1px solid #cbd5e1; color: #334155; }
.btn-primary { background: #0284c7; color: white; border: none; }
.btn-primary:hover { background: #0369a1; }
</style>
