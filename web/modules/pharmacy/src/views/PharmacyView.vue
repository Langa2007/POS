<template>
  <div class="pharmacy-view">
    <!-- Top Header -->
    <header class="view-header">
      <div class="header-main">
        <span class="shop-badge">💊 PHARMACY DISPENSARY SUITE</span>
        <h1>Pharmacy Dispensary & Clinical Counter</h1>
        <p class="view-description">
          Prescription intake, First-Expired First-Out (FEFO) batch traceability, statutory DDA Poison Book register, generic bioequivalence substitution, and clinical label generation.
        </p>
      </div>

      <!-- Quick clinical stats -->
      <div class="quick-kpis">
        <div class="kpi-card">
          <span class="kpi-label">Rx Intake Queue</span>
          <span class="kpi-val">2 Pending</span>
        </div>
        <div class="kpi-card">
          <span class="kpi-label">FEFO Near-Expiry</span>
          <span class="kpi-val warning">3 Batches</span>
        </div>
        <div class="kpi-card">
          <span class="kpi-label">Controlled Book</span>
          <span class="kpi-val">3 DDA Entries</span>
        </div>
      </div>
    </header>

    <!-- Notification Toast if any -->
    <div v-if="notification" class="inline-notice" :class="notification.type">
      <span>{{ notification.message }}</span>
      <button class="notice-close" @click="notification = null">✕</button>
    </div>

    <!-- Navigation Tabs -->
    <nav class="pharmacy-tabs">
      <button
        type="button"
        class="tab-btn"
        :class="{ 'active': activeTab === 'prescriptions' }"
        @click="activeTab = 'prescriptions'"
      >
        📋 Prescriptions & Intake
      </button>
      <button
        type="button"
        class="tab-btn"
        :class="{ 'active': activeTab === 'batches' }"
        @click="activeTab = 'batches'"
      >
        🔬 Batch Tracking & FEFO
      </button>
      <button
        type="button"
        class="tab-btn"
        :class="{ 'active': activeTab === 'controlled' }"
        @click="activeTab = 'controlled'"
      >
        🔒 Controlled Drugs Register
      </button>
      <button
        type="button"
        class="tab-btn"
        :class="{ 'active': activeTab === 'generics' }"
        @click="activeTab = 'generics'"
      >
        💊 Generic Substitution Guide
      </button>
      <button
        type="button"
        class="tab-btn"
        :class="{ 'active': activeTab === 'labels' }"
        @click="activeTab = 'labels'"
      >
        🏷️ Clinical Dispensing Labels
      </button>
    </nav>

    <!-- Tab Panels -->
    <main class="tab-content-panel">
      <!-- 1. Prescriptions & Dispensing -->
      <section v-if="activeTab === 'prescriptions'">
        <PrescriptionStation
          @add-item="handleAddItem"
          @notification="showNotice"
        />
      </section>

      <!-- 2. Batch & FEFO Engine -->
      <section v-if="activeTab === 'batches'">
        <BatchFefoManager
          @notification="showNotice"
        />
      </section>

      <!-- 3. Controlled Substances Register -->
      <section v-if="activeTab === 'controlled'">
        <ControlledDrugsRegister
          @add-item="handleAddItem"
          @notification="showNotice"
        />
      </section>

      <!-- 4. Generic Substitution -->
      <section v-if="activeTab === 'generics'">
        <GenericSubstitutionFinder
          @add-item="handleAddItem"
          @notification="showNotice"
        />
      </section>

      <!-- 5. Clinical Labels -->
      <section v-if="activeTab === 'labels'">
        <DispensingLabelGenerator
          @notification="showNotice"
        />
      </section>
    </main>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import PrescriptionStation from '../components/PrescriptionStation.vue'
import BatchFefoManager from '../components/BatchFefoManager.vue'
import ControlledDrugsRegister from '../components/ControlledDrugsRegister.vue'
import GenericSubstitutionFinder from '../components/GenericSubstitutionFinder.vue'
import DispensingLabelGenerator from '../components/DispensingLabelGenerator.vue'

const activeTab = ref('prescriptions')
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
  try {
    const existingCart = JSON.parse(localStorage.getItem('pos.retailCartPending') || '[]')
    existingCart.push(item)
    localStorage.setItem('pos.retailCartPending', JSON.stringify(existingCart))
    window.dispatchEvent(new CustomEvent('pos:add-item', { detail: item }))
  } catch (e) {
    console.error(e)
  }
  showNotice({
    message: `Dispensed "${item.name}" (KES ${(item.price_minor / 100).toFixed(2)}) to active cart.`,
    type: 'success',
  })
}
</script>

<style scoped>
.pharmacy-view {
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
  color: #0369a1;
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
  max-width: 680px;
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
.kpi-val.warning {
  color: #d97706;
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
.pharmacy-tabs {
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
