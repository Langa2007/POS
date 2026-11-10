<template>
  <aside class="hardware-checkout-sidebar">
    <div class="sidebar-header">
      <div class="title-wrap">
        <PosIcon name="wrench" size="18" class="side-icon" />
        <h4>Hardware &amp; Trade Counter</h4>
      </div>
      <span class="badge-trade">Trade Desk</span>
    </div>

    <!-- Contractor Account Selector -->
    <div class="sidebar-block">
      <label class="block-label">Customer / Contractor Account:</label>
      <select v-model="selectedContractorId" class="contractor-select" @change="onContractorChange">
        <option value="walk-in">Retail Walk-in Customer (Standard Cash)</option>
        <option value="cnt-1">Apex Building Contractors (8% Trade &bull; Limit KES 1M)</option>
        <option value="cnt-2">Fundi Juma Plumbing (5% Trade &bull; Limit KES 250k)</option>
        <option value="cnt-3">Coast Highrise Civil (10% Trade &bull; Limit KES 800k)</option>
      </select>
    </div>

    <!-- Trade Account Status -->
    <div v-if="activeContractor" class="contractor-status-box">
      <div class="c-line">
        <span>Available Credit:</span>
        <strong>KES {{ (activeContractor.availMinor / 100).toLocaleString() }}</strong>
      </div>
      <div class="c-line">
        <span>Trade Discount:</span>
        <strong class="trade-save">{{ activeContractor.discountPct }}% Off Retail</strong>
      </div>
    </div>

    <!-- Quick Tools -->
    <div class="sidebar-block">
      <label class="block-label">Counter Tools:</label>
      <div class="tool-buttons">
        <button type="button" class="tool-btn" @click="openCutTool">
          <PosIcon name="scissors" size="14" />
          Cut Pipe / Cable
        </button>
        <button type="button" class="tool-btn" @click="openMultiUom">
          <PosIcon name="layers" size="14" />
          Bulk UOM Convert
        </button>
      </div>
    </div>

    <!-- Delivery Option -->
    <div class="sidebar-block">
      <label class="block-label">Site Logistics:</label>
      <label class="delivery-toggle-label">
        <input type="checkbox" v-model="scheduleDelivery" @change="onDeliveryToggle" />
        <span>Schedule Truck Site Delivery</span>
      </label>
      <div v-if="scheduleDelivery" class="delivery-zone-picker">
        <select v-model="selectedZoneFee" class="zone-select" @change="applyDeliveryFee">
          <option :value="150000">Zone 1: Local Municipal (&lt; 10km &bull; KES 1,500)</option>
          <option :value="350000">Zone 2: Suburbs (10-30km &bull; KES 3,500)</option>
          <option :value="750000">Zone 3: Remote Site (&gt; 30km &bull; KES 7,500)</option>
        </select>
      </div>
    </div>

    <div class="sidebar-actions">
      <button type="button" class="btn-side-action" @click="printQuote">
        <PosIcon name="file-text" size="14" />
        Print Quotation
      </button>
      <button type="button" class="btn-side-action" @click="printDeliveryNote">
        <PosIcon name="truck" size="14" />
        Delivery Note
      </button>
    </div>
  </aside>
</template>

<script setup>
import { ref, computed } from 'vue'
import { PosIcon } from '@pos/ui'

const props = defineProps({
  cart: {
    type: Array,
    default: () => [],
  },
})

const selectedContractorId = ref('walk-in')
const scheduleDelivery = ref(false)
const selectedZoneFee = ref(150000)

const contractorsMap = {
  'cnt-1': { name: 'Apex Building Contractors', availMinor: 58000000, discountPct: 8 },
  'cnt-2': { name: 'Fundi Juma Plumbing', availMinor: 6500000, discountPct: 5 },
  'cnt-3': { name: 'Coast Highrise Civil', availMinor: 68000000, discountPct: 10 },
}

const activeContractor = computed(() => {
  return contractorsMap[selectedContractorId.value] || null
})

function onContractorChange() {
  if (activeContractor.value) {
    window.dispatchEvent(new CustomEvent('pos:apply-discount', {
      detail: {
        type: 'percent',
        value: activeContractor.value.discountPct,
        description: `Trade Discount (${activeContractor.value.discountPct}% - ${activeContractor.value.name})`,
      }
    }))
    window.dispatchEvent(new CustomEvent('pos:notification', {
      detail: { message: `Contractor trade pricing active: ${activeContractor.value.discountPct}% discount applied`, type: 'success' }
    }))
  } else {
    window.dispatchEvent(new CustomEvent('pos:notification', {
      detail: { message: 'Reset to standard walk-in retail pricing', type: 'info' }
    }))
  }
}

function onDeliveryToggle() {
  if (scheduleDelivery.value) {
    applyDeliveryFee()
  }
}

function applyDeliveryFee() {
  const feeItem = {
    id: `del-${Date.now()}`,
    name: 'Site Truck Delivery Dispatch',
    sku: 'SRV-DEL-FEE',
    price_minor: selectedZoneFee.value,
    unit_price_minor: selectedZoneFee.value,
    quantity: 1,
    kind: 'service',
  }

  try {
    const existing = JSON.parse(localStorage.getItem('pos.retailCartPending') || '[]')
    localStorage.setItem('pos.retailCartPending', JSON.stringify([...existing, feeItem]))
  } catch (e) {}

  window.dispatchEvent(new CustomEvent('pos:add-item', { detail: feeItem }))
  window.dispatchEvent(new CustomEvent('pos:notification', {
    detail: { message: `Added Delivery Fee (KES ${(selectedZoneFee.value / 100).toLocaleString()}) to cart`, type: 'info' }
  }))
}

function openCutTool() {
  window.dispatchEvent(new CustomEvent('pos:notification', {
    detail: { message: 'Cut-to-Length station available in Hardware Workstation tab', type: 'info' }
  }))
}

function openMultiUom() {
  window.dispatchEvent(new CustomEvent('pos:notification', {
    detail: { message: 'Multi-UOM pack converter available in Hardware Workstation tab', type: 'info' }
  }))
}

function printQuote() {
  window.dispatchEvent(new CustomEvent('pos:notification', {
    detail: { message: 'Generated printable Pro-Forma Construction Quotation', type: 'info' }
  }))
}

function printDeliveryNote() {
  window.dispatchEvent(new CustomEvent('pos:notification', {
    detail: { message: 'Printed Gate Pass & Delivery Dispatch Manifest', type: 'info' }
  }))
}
</script>

<style scoped>
.hardware-checkout-sidebar {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.sidebar-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid #f1f5f9;
  padding-bottom: 8px;
}

.title-wrap {
  display: flex;
  align-items: center;
  gap: 6px;
}

.side-icon {
  color: #0284c7;
}

.sidebar-header h4 {
  margin: 0;
  font-size: 13px;
  color: #0f172a;
}

.badge-trade {
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  background: #e0f2fe;
  color: #0369a1;
  padding: 2px 6px;
  border-radius: 4px;
}

.sidebar-block {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.block-label {
  font-size: 11px;
  font-weight: 700;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.contractor-select, .zone-select {
  padding: 8px;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 12px;
  background: #f8fafc;
}

.contractor-status-box {
  background: #f0fdf4;
  border: 1px solid #bbf7d0;
  border-radius: 8px;
  padding: 10px 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 11px;
}

.c-line {
  display: flex;
  justify-content: space-between;
}

.trade-save {
  color: #15803d;
}

.tool-buttons {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 6px;
}

.tool-btn {
  padding: 8px;
  border: 1px solid #cbd5e1;
  background: #f8fafc;
  border-radius: 6px;
  font-size: 11px;
  font-weight: 600;
  color: #334155;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  transition: all 0.15s ease;
}

.tool-btn:hover {
  background: #f1f5f9;
}

.delivery-toggle-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #334155;
  cursor: pointer;
}

.delivery-zone-picker {
  margin-top: 4px;
}

.sidebar-actions {
  display: flex;
  gap: 8px;
  margin-top: 4px;
  border-top: 1px solid #f1f5f9;
  padding-top: 10px;
}

.btn-side-action {
  flex: 1;
  padding: 8px;
  border: 1px solid #cbd5e1;
  background: white;
  border-radius: 6px;
  font-size: 11px;
  font-weight: 600;
  color: #334155;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
}

.btn-side-action:hover {
  background: #f1f5f9;
}
</style>
