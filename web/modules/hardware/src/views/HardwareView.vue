<template>
  <div class="hardware-view">
    <!-- Top KPI Bar -->
    <header class="hardware-topbar">
      <div class="title-group">
        <div class="main-title">
          <PosIcon name="wrench" size="24" class="title-icon" />
          <h2>Hardware &amp; Building Supplies Workstation</h2>
        </div>
        <p class="view-subtitle">Bulk multi-UOM packaging, precision cut-to-length, contractor credit, and site delivery</p>
      </div>

      <div class="quick-kpis">
        <div class="kpi-card">
          <span class="kpi-lbl">Active Trade Credit</span>
          <strong class="kpi-val highlight">KES 1.22M</strong>
          <small class="kpi-sub">Across 4 contractor accounts</small>
        </div>
        <div class="kpi-card">
          <span class="kpi-lbl">Offcut Remnants</span>
          <strong class="kpi-val">18 Pieces</strong>
          <small class="kpi-sub">Pipes, timber, cables in yard</small>
        </div>
        <div class="kpi-card">
          <span class="kpi-lbl">Dispatches Today</span>
          <strong class="kpi-val warning">3 Trucks</strong>
          <small class="kpi-sub">12.9 Tonnes payload</small>
        </div>
        <div class="kpi-card">
          <span class="kpi-lbl">Today's Trade Sales</span>
          <strong class="kpi-val success">KES 485,000</strong>
          <small class="kpi-sub">Cement &amp; timber leading</small>
        </div>
      </div>
    </header>

    <!-- Notification Toast Banner -->
    <div v-if="notification" class="notification-banner" :class="notification.type">
      <PosIcon :name="notification.type === 'success' ? 'check-circle' : 'info'" size="16" />
      <span>{{ notification.message }}</span>
      <button class="banner-close" @click="notification = null">✕</button>
    </div>

    <!-- Navigation Tabs -->
    <nav class="hardware-tabs">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        type="button"
        class="tab-btn"
        :class="{ 'active': activeTab === tab.id }"
        @click="activeTab = tab.id"
      >
        <PosIcon :name="tab.icon" size="16" />
        <span>{{ tab.label }}</span>
        <span v-if="tab.badge" class="tab-badge" :class="tab.badgeClass">{{ tab.badge }}</span>
      </button>
    </nav>

    <!-- Tab View Components -->
    <main class="tab-content">
      <MultiUomConverter
        v-if="activeTab === 'uom'"
        @notification="showNotification"
      />

      <CutToLengthManager
        v-else-if="activeTab === 'ctl'"
        @notification="showNotification"
      />

      <ContractorAccountsManager
        v-else-if="activeTab === 'contractors'"
        @notification="showNotification"
      />

      <DeliveryDispatchManager
        v-else-if="activeTab === 'delivery'"
        @notification="showNotification"
      />
    </main>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { PosIcon } from '@pos/ui'
import MultiUomConverter from '../components/MultiUomConverter.vue'
import CutToLengthManager from '../components/CutToLengthManager.vue'
import ContractorAccountsManager from '../components/ContractorAccountsManager.vue'
import DeliveryDispatchManager from '../components/DeliveryDispatchManager.vue'

const activeTab = ref('uom')
const notification = ref(null)

const tabs = [
  { id: 'uom', label: 'Multi-Unit Conversions & Bulk', icon: 'layers' },
  { id: 'ctl', label: 'Cut-to-Length & Offcuts', icon: 'scissors', badge: 'Workshop Active', badgeClass: 'badge-blue' },
  { id: 'contractors', label: 'Contractor Trade Credit', icon: 'hard-hat', badge: 'KES 1.2M Used', badgeClass: 'badge-orange' },
  { id: 'delivery', label: 'Heavy Site Dispatch', icon: 'truck', badge: '3 Scheduled', badgeClass: 'badge-green' },
]

function showNotification({ message, type = 'info' }) {
  notification.value = { message, type }
  setTimeout(() => {
    if (notification.value?.message === message) {
      notification.value = null
    }
  }, 4000)
}
</script>

<style scoped>
.hardware-view {
  display: flex;
  flex-direction: column;
  gap: 18px;
  padding: 8px 0;
}

.hardware-topbar {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  flex-wrap: wrap;
  gap: 16px;
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 18px 20px;
}

.main-title {
  display: flex;
  align-items: center;
  gap: 10px;
}

.title-icon {
  color: #0284c7;
}

.main-title h2 {
  margin: 0;
  font-size: 20px;
  color: #0f172a;
}

.view-subtitle {
  margin: 4px 0 0;
  font-size: 13px;
  color: #64748b;
}

.quick-kpis {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.kpi-card {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 10px 14px;
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 120px;
}

.kpi-lbl {
  font-size: 10px;
  font-weight: 700;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.kpi-val {
  font-size: 16px;
  font-weight: 800;
  color: #0f172a;
}

.kpi-val.highlight { color: #0284c7; }
.kpi-val.warning { color: #d97706; }
.kpi-val.success { color: #16a34a; }

.kpi-sub {
  font-size: 10px;
  color: #94a3b8;
}

.notification-banner {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 16px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
}

.notification-banner.success {
  background: #f0fdf4;
  border: 1px solid #bbf7d0;
  color: #166534;
}

.notification-banner.info {
  background: #f0f9ff;
  border: 1px solid #bae6fd;
  color: #0369a1;
}

.banner-close {
  margin-left: auto;
  background: none;
  border: none;
  cursor: pointer;
  color: inherit;
  font-size: 14px;
}

.hardware-tabs {
  display: flex;
  gap: 8px;
  border-bottom: 2px solid #e2e8f0;
  padding-bottom: 2px;
  overflow-x: auto;
}

.tab-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  background: transparent;
  border: none;
  border-bottom: 2px solid transparent;
  margin-bottom: -2px;
  font-size: 13px;
  font-weight: 600;
  color: #64748b;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.15s ease;
}

.tab-btn:hover {
  color: #0f172a;
}

.tab-btn.active {
  color: #0284c7;
  border-bottom-color: #0284c7;
  font-weight: 700;
}

.tab-badge {
  font-size: 10px;
  padding: 2px 6px;
  border-radius: 10px;
  font-weight: 700;
}

.badge-blue { background: #e0f2fe; color: #0369a1; }
.badge-orange { background: #ffedd5; color: #c2410c; }
.badge-green { background: #dcfce7; color: #15803d; }

.tab-content {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 20px;
}
</style>
