<template>
  <div class="restaurant-view">
    <!-- Top KPI Bar -->
    <header class="restaurant-topbar">
      <div class="title-group">
        <div class="main-title">
          <PosIcon name="utensils" size="24" class="title-icon" />
          <h2>Restaurant &amp; Bar Workstation</h2>
        </div>
        <p class="view-subtitle">Floor management, kitchen order routing, table tabs, and bills</p>
      </div>

      <div class="quick-kpis">
        <div class="kpi-card">
          <span class="kpi-lbl">Table Occupancy</span>
          <strong class="kpi-val highlight">{{ summary.tables }}</strong>
          <small class="kpi-sub">Tables in service</small>
        </div>
        <div class="kpi-card">
          <span class="kpi-lbl">Active Covers</span>
          <strong class="kpi-val">{{ summary.covers }} Guests</strong>
          <small class="kpi-sub">Active covers</small>
        </div>
        <div class="kpi-card">
          <span class="kpi-lbl">Kitchen Queue</span>
          <strong class="kpi-val warning">{{ summary.kitchen_tickets }} KOTs</strong>
          <small class="kpi-sub">Awaiting preparation</small>
        </div>
        <div class="kpi-card">
          <span class="kpi-lbl">Today's Dining Revenue</span>
          <strong class="kpi-val success">KES {{ (Number(summary.total_minor) / 100).toLocaleString() }}</strong>
          <small class="kpi-sub">{{ summary.sales_count }} settled sales</small>
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
    <nav class="restaurant-tabs">
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
      <TableServiceManager
        v-if="activeTab === 'tables'"
        @notification="showNotification"
      />

      <KitchenDisplayManager
        v-else-if="activeTab === 'kitchen'"
        @notification="showNotification"
      />

      <MenuModifiersManager
        v-else-if="activeTab === 'modifiers'"
        @notification="showNotification"
      />

      <SplitBillManager
        v-else-if="activeTab === 'split'"
        @notification="showNotification"
      />
    </main>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { PosIcon } from '@pos/ui'
import { useApi } from '../../../../apps/pos/src/composables/useApi.js'
import TableServiceManager from '../components/TableServiceManager.vue'
import KitchenDisplayManager from '../components/KitchenDisplayManager.vue'
import MenuModifiersManager from '../components/MenuModifiersManager.vue'
import SplitBillManager from '../components/SplitBillManager.vue'

const activeTab = ref('tables')
const notification = ref(null)

const api = useApi()
const summary = ref({ tables: 0, covers: 0, kitchen_tickets: 0, total_minor: 0, sales_count: 0 })
const tabs = [
  { id: 'tables', label: 'Floor Plan & Tables', icon: 'layout-grid' },
  { id: 'kitchen', label: 'Kitchen Display (KDS)', icon: 'flame' },
  { id: 'modifiers', label: 'Dish Modifiers & Combos', icon: 'chef-hat' },
  { id: 'split', label: 'Split Bills & Gratuity', icon: 'split' },
]
onMounted(async () => {
  try {
    summary.value = await api.get('/api/v1/restaurant/summary')
  } catch {
    notification.value = { message: 'Restaurant summary is unavailable. Please contact your administrator.', type: 'error' }
  }
})

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
.restaurant-view {
  display: flex;
  flex-direction: column;
  gap: 18px;
  padding: 8px 0;
}

.restaurant-topbar {
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
  color: #ea580c;
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

.kpi-val.highlight { color: #ea580c; }
.kpi-val.warning { color: #dc2626; }
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

.restaurant-tabs {
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
  color: #ea580c;
  border-bottom-color: #ea580c;
  font-weight: 700;
}

.tab-badge {
  font-size: 10px;
  padding: 2px 6px;
  border-radius: 10px;
  font-weight: 700;
}

.badge-blue { background: #e0f2fe; color: #0369a1; }
.badge-red { background: #fee2e2; color: #b91c1c; }

.tab-content {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 20px;
}
</style>
