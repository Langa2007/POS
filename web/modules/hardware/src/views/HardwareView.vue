<template>
  <div class="hardware-view">
    <header class="view-header"><h1>{{ $t('hardware') }}</h1><p class="view-description">{{ $t('multiUomDesc') }}</p></header>
    <div class="feature-grid">
      <section class="feature-card" v-for="f in features" :key="f.key">
        <div class="feature-icon">{{ f.icon }}</div><h3>{{ $t(f.titleKey) }}</h3><p>{{ $t(f.descKey) }}</p>
      </section>
    </div>
  </div>
</template>
<script setup>
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
const features = [
  { key: 'uom', icon: '📏', titleKey: 'multiUom', descKey: 'multiUomDesc' },
  { key: 'cut', icon: '✂️', titleKey: 'cutToLength', descKey: 'cutToLengthDesc' },
  { key: 'credit', icon: '💳', titleKey: 'creditAccounts', descKey: 'creditAccountsDesc' },
  { key: 'contract', icon: '📋', titleKey: 'contractPricing', descKey: 'contractPricingDesc' },
  { key: 'quote', icon: '📝', titleKey: 'quotes', descKey: 'quotesDesc' },
  { key: 'delivery', icon: '🚚', titleKey: 'delivery', descKey: 'deliveryDesc' },
]
</script>
<style scoped>
.hardware-view{padding:1.5rem}.view-header{margin-bottom:1.5rem}.view-header h1{margin:0 0 .5rem;font-size:1.75rem}.view-description{color:var(--color-muted);margin:0}.feature-grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(250px,1fr));gap:1rem}.feature-card{padding:1.25rem;border:1px solid var(--color-border);border-radius:8px;background:var(--color-surface)}.feature-icon{font-size:2rem;margin-bottom:.5rem}.feature-card h3{margin:0 0 .5rem;font-size:1.1rem}.feature-card p{margin:0;color:var(--color-muted);font-size:.9rem}
</style>