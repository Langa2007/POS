<template>
  <aside class="restaurant-checkout-sidebar">
    <h3>{{ $t('checkoutSidebar') }}</h3>
    <div class="sidebar-sections">
      <section v-if="hasTableService" class="sidebar-section">
        <h4>{{ $t('tableService') }}</h4>
        <PosButton variant="secondary" @click="openTable">{{ $t('openTable') }}</PosButton>
      </section>
      <section v-if="hasKitchenDisplay" class="sidebar-section">
        <h4>{{ $t('kitchenDisplay') }}</h4>
        <PosButton variant="secondary" @click="sendToKitchen">{{ $t('sendToKitchen') }}</PosButton>
      </section>
      <section v-if="hasMenuModifiers" class="sidebar-section">
        <h4>{{ $t('menuModifiers') }}</h4>
        <PosButton variant="secondary" @click="modifyOrder">{{ $t('modifyOrder') }}</PosButton>
      </section>
      <section v-if="hasSplitBills" class="sidebar-section">
        <h4>{{ $t('splitBills') }}</h4>
        <PosButton variant="secondary" @click="splitBill">{{ $t('splitBill') }}</PosButton>
      </section>
    </div>
  </aside>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { PosButton } from '@pos/ui'
import { useManifestStore } from '@/stores/manifest.js'

const { t } = useI18n()
const manifest = useManifestStore()

const hasTableService = computed(() => manifest.has('table_service'))
const hasKitchenDisplay = computed(() => manifest.has('kitchen_display'))
const hasMenuModifiers = computed(() => manifest.has('menu_modifiers'))
const hasSplitBills = computed(() => manifest.has('split_bills'))

function openTable() {
  console.log('Open table')
}

function sendToKitchen() {
  console.log('Send to kitchen')
}

function modifyOrder() {
  console.log('Modify order')
}

function splitBill() {
  console.log('Split bill')
}
</script>

<style scoped>
.restaurant-checkout-sidebar {
  padding: 1rem;
}
.restaurant-checkout-sidebar h3 {
  margin: 0 0 1rem;
  font-size: 1rem;
  border-bottom: 1px solid var(--color-border);
  padding-bottom: 0.5rem;
}
.sidebar-section {
  margin-bottom: 1rem;
}
.sidebar-section h4 {
  margin: 0 0 0.5rem;
  font-size: 0.85rem;
  color: var(--color-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
</style>