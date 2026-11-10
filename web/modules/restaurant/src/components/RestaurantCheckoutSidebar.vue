<template>
  <aside class="restaurant-checkout-sidebar">
    <div class="sidebar-header">
      <div class="title-wrap">
        <PosIcon name="utensils" size="18" class="side-icon" />
        <h4>Restaurant Service</h4>
      </div>
      <span class="badge-active">Dining</span>
    </div>

    <!-- Active Dining Table Selector -->
    <div class="sidebar-block">
      <label class="block-label">Assigned Table / Order Type:</label>
      <select v-model="selectedTableNumber" class="table-select" @change="onTableChange">
        <option value="1">Table 1 &mdash; Main (Mary W.)</option>
        <option value="2">Table 2 &mdash; Main (Kevin O.)</option>
        <option value="5">Table 5 &mdash; Terrace (Mary W.)</option>
        <option value="VIP 1">VIP 1 &mdash; Suite (Alice N.)</option>
        <option value="Bar 1">Bar 1 &mdash; Lounge (Brian M.)</option>
        <option value="Takeaway">Takeaway / Counter Delivery</option>
      </select>
    </div>

    <!-- Course Firing Quick Actions -->
    <div class="sidebar-block">
      <label class="block-label">Kitchen Course Firing:</label>
      <div class="course-buttons">
        <button
          type="button"
          class="course-btn"
          :class="{ 'active': firedCourse === 'Starters' }"
          @click="fireCourse('Starters')"
        >
          <PosIcon name="flame" size="13" />
          Fire Starters
        </button>
        <button
          type="button"
          class="course-btn"
          :class="{ 'active': firedCourse === 'Mains' }"
          @click="fireCourse('Mains')"
        >
          <PosIcon name="flame" size="13" />
          Fire Mains
        </button>
        <button
          type="button"
          class="course-btn"
          :class="{ 'active': firedCourse === 'Desserts' }"
          @click="fireCourse('Desserts')"
        >
          <PosIcon name="coffee" size="13" />
          Fire Desserts
        </button>
      </div>
    </div>

    <!-- Table Service & Tip Add-ons -->
    <div class="sidebar-block">
      <label class="block-label">Service Fee &amp; Gratuity:</label>
      <div class="levy-toggle">
        <label class="check-box-label">
          <input type="checkbox" v-model="applyServiceCharge" @change="onServiceChargeToggle" />
          <span>Apply 10% Catering Levy</span>
        </label>
      </div>
      <div class="tip-pills">
        <button
          v-for="tip in [0, 5, 10, 15]"
          :key="tip"
          type="button"
          class="tip-pill"
          :class="{ 'active': activeTip === tip }"
          @click="setTip(tip)"
        >
          {{ tip === 0 ? 'No Tip' : `${tip}% Tip` }}
        </button>
      </div>
    </div>

    <!-- Quick Table Actions -->
    <div class="sidebar-actions">
      <button type="button" class="btn-side-action" @click="printGuestCheck">
        <PosIcon name="printer" size="14" />
        Print Guest Check
      </button>
      <button type="button" class="btn-side-action" @click="openSplitPrompt">
        <PosIcon name="split" size="14" />
        Split Bill
      </button>
    </div>
  </aside>
</template>

<script setup>
import { ref } from 'vue'
import { PosIcon } from '@pos/ui'

const props = defineProps({
  cart: {
    type: Array,
    default: () => [],
  },
})

const selectedTableNumber = ref('1')
const firedCourse = ref('Mains')
const applyServiceCharge = ref(true)
const activeTip = ref(10)

function onTableChange() {
  window.dispatchEvent(new CustomEvent('pos:notification', {
    detail: { message: `Active Till context set to Table ${selectedTableNumber.value}`, type: 'info' }
  }))
}

function fireCourse(courseName) {
  firedCourse.value = courseName
  window.dispatchEvent(new CustomEvent('pos:notification', {
    detail: { message: `KOT Alert: Course "${courseName}" fired to kitchen for Table ${selectedTableNumber.value}!`, type: 'success' }
  }))
}

function onServiceChargeToggle() {
  if (applyServiceCharge.value) {
    window.dispatchEvent(new CustomEvent('pos:notification', {
      detail: { message: '10% Catering Levy added to dining bill', type: 'info' }
    }))
  }
}

function setTip(tipPct) {
  activeTip.value = tipPct
  window.dispatchEvent(new CustomEvent('pos:notification', {
    detail: { message: tipPct > 0 ? `${tipPct}% Server gratuity recorded` : 'Gratuity removed', type: 'info' }
  }))
}

function printGuestCheck() {
  window.dispatchEvent(new CustomEvent('pos:notification', {
    detail: { message: `Pro-forma guest check slip printed for Table ${selectedTableNumber.value}`, type: 'info' }
  }))
}

function openSplitPrompt() {
  window.dispatchEvent(new CustomEvent('pos:notification', {
    detail: { message: `Split bill assistant activated for Table ${selectedTableNumber.value}`, type: 'info' }
  }))
}
</script>

<style scoped>
.restaurant-checkout-sidebar {
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
  color: #ea580c;
}

.sidebar-header h4 {
  margin: 0;
  font-size: 13px;
  color: #0f172a;
}

.badge-active {
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  background: #ffedd5;
  color: #c2410c;
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

.table-select {
  padding: 8px;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 12px;
  background: #f8fafc;
}

.course-buttons {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 4px;
}

.course-btn {
  padding: 6px 4px;
  border: 1px solid #cbd5e1;
  background: white;
  border-radius: 6px;
  font-size: 11px;
  font-weight: 600;
  color: #334155;
  cursor: pointer;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  transition: all 0.15s ease;
}

.course-btn:hover {
  background: #f1f5f9;
}

.course-btn.active {
  background: #ea580c;
  color: white;
  border-color: #ea580c;
}

.levy-toggle {
  font-size: 12px;
  color: #334155;
}

.check-box-label {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
}

.tip-pills {
  display: flex;
  gap: 4px;
  margin-top: 4px;
}

.tip-pill {
  flex: 1;
  padding: 4px 6px;
  border: 1px solid #cbd5e1;
  background: #f8fafc;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  color: #475569;
  cursor: pointer;
}

.tip-pill.active {
  background: #8b5cf6;
  color: white;
  border-color: #8b5cf6;
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
