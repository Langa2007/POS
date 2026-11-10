<template>
  <aside class="hardware-credit-slot">
    <h3>{{ $t('checkoutSidebar') }}</h3>
    <div class="credit-sections">
      <section v-if="hasCutToLength" class="credit-section"><h4>{{ $t('cutToLength') }}</h4><PosButton variant="secondary" @click="cutMaterial">{{ $t('cutMaterial') }}</PosButton></section>
      <section v-if="hasContractPricing" class="credit-section"><h4>{{ $t('contractPricing') }}</h4><PosButton variant="secondary" @click="applyContract">{{ $t('applyContract') }}</PosButton></section>
      <section v-if="hasCreditAccounts" class="credit-section"><h4>{{ $t('creditAccounts') }}</h4><PosButton variant="secondary" @click="checkCredit">{{ creditLimit }} available</PosButton></section>
      <section v-if="hasQuotes" class="credit-section"><h4>{{ $t('quotes') }}</h4><PosButton variant="secondary" @click="createQuote">{{ $t('createQuote') }}</PosButton></section>
      <section v-if="hasDelivery" class="credit-section"><h4>{{ $t('delivery') }}</h4><PosButton variant="secondary" @click="scheduleDelivery">{{ $t('scheduleDelivery') }}</PosButton></section>
    </div>
  </aside>
</template>
<script setup>
import { computed } from 'vue'; import { useI18n } from 'vue-i18n'; import { PosButton } from '@pos/ui'; import { useManifestStore } from '@/stores/manifest.js'
const { t } = useI18n(); const manifest = useManifestStore()
const hasCutToLength = computed(() => manifest.has('cut_to_length'))
const hasContractPricing = computed(() => manifest.has('contract_pricing'))
const hasCreditAccounts = computed(() => manifest.has('credit_accounts'))
const hasQuotes = computed(() => manifest.has('quotes'))
const hasDelivery = computed(() => manifest.has('delivery'))
const creditLimit = 'KES 50,000'
function cutMaterial(){console.log('Cut material')} function applyContract(){console.log('Apply contract price')} function checkCredit(){console.log('Check credit limit')} function createQuote(){console.log('Create quote')} function scheduleDelivery(){console.log('Schedule delivery')}
</script>
<style scoped>
.hardware-credit-slot{padding:1rem}.hardware-credit-slot h3{margin:0 0 1rem;font-size:1rem;border-bottom:1px solid var(--color-border);padding-bottom:.5rem}.credit-section{margin-bottom:1rem}.credit-section h4{margin:0 0 .5rem;font-size:.85rem;color:var(--color-muted);text-transform:uppercase;letter-spacing:.05em}
</style>