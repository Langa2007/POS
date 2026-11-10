<template>
  <div class="hospitality-view">
    <header class="view-header">
      <h1>{{ $t('hospitality') }}</h1>
      <p class="view-description">{{ $t('roomBookingDesc') }}</p>
    </header>
    <div class="feature-grid">
      <section class="feature-card" v-for="feature in features" :key="feature.key">
        <div class="feature-icon">{{ feature.icon }}</div>
        <h3>{{ $t(feature.titleKey) }}</h3>
        <p>{{ $t(feature.descKey) }}</p>
      </section>
    </div>
  </div>
</template>

<script setup>
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const features = [
  { key: 'booking', icon: '🛏️', titleKey: 'roomBooking', descKey: 'roomBookingDesc' },
  { key: 'rates', icon: '💰', titleKey: 'ratePlans', descKey: 'ratePlansDesc' },
  { key: 'folio', icon: '📄', titleKey: 'guestFolio', descKey: 'guestFolioDesc' },
  { key: 'housekeeping', icon: '🧹', titleKey: 'housekeeping', descKey: 'housekeepingDesc' },
  { key: 'audit', icon: '🌙', titleKey: 'nightAudit', descKey: 'nightAuditDesc' },
  { key: 'group', icon: '👥', titleKey: 'groupBilling', descKey: 'groupBillingDesc' },
  { key: 'deposits', icon: '💵', titleKey: 'deposits', descKey: 'depositsDesc' },
]
</script>

<style scoped>
.hospitality-view { padding: 1.5rem; }
.view-header { margin-bottom: 1.5rem; }
.view-header h1 { margin: 0 0 0.5rem; font-size: 1.75rem; }
.view-description { color: var(--color-muted); margin: 0; }
.feature-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 1rem; }
.feature-card { padding: 1.25rem; border: 1px solid var(--color-border); border-radius: 8px; background: var(--color-surface); }
.feature-icon { font-size: 2rem; margin-bottom: 0.5rem; }
.feature-card h3 { margin: 0 0 0.5rem; font-size: 1.1rem; }
.feature-card p { margin: 0; color: var(--color-muted); font-size: 0.9rem; }
</style>