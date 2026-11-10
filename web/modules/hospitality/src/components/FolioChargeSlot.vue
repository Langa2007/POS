<template>
  <aside class="hospitality-folio-slot">
    <h3>{{ $t('folioCharge') }}</h3>
    <div class="folio-sections">
      <section v-if="hasRoomBooking" class="folio-section">
        <h4>{{ $t('roomBooking') }}</h4>
        <PosButton variant="secondary" @click="checkIn">{{ $t('checkIn') }}</PosButton>
        <PosButton variant="secondary" @click="checkOut">{{ $t('checkOut') }}</PosButton>
      </section>
      <section v-if="hasGuestFolio" class="folio-section">
        <h4>{{ $t('guestFolio') }}</h4>
        <PosButton variant="secondary" @click="viewFolio">{{ $t('viewFolio') }}</PosButton>
        <PosButton variant="secondary" @click="addCharge">{{ $t('addCharge') }}</PosButton>
      </section>
      <section v-if="hasHousekeeping" class="folio-section">
        <h4>{{ $t('housekeeping') }}</h4>
        <p class="room-status">Room status: {{ roomStatus }}</p>
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

const hasRoomBooking = computed(() => manifest.has('room_booking'))
const hasGuestFolio = computed(() => manifest.has('guest_folio'))
const hasHousekeeping = computed(() => manifest.has('housekeeping'))

const roomStatus = 'Clean - Ready'

function checkIn() { console.log('Check in') }
function checkOut() { console.log('Check out') }
function viewFolio() { console.log('View folio') }
function addCharge() { console.log('Add charge to folio') }
</script>

<style scoped>
.hospitality-folio-slot { padding: 1rem; }
.hospitality-folio-slot h3 { margin: 0 0 1rem; font-size: 1rem; border-bottom: 1px solid var(--color-border); padding-bottom: 0.5rem; }
.folio-section { margin-bottom: 1rem; }
.folio-section h4 { margin: 0 0 0.5rem; font-size: 0.85rem; color: var(--color-muted); text-transform: uppercase; letter-spacing: 0.05em; }
.room-status { margin: 0; font-size: 0.85rem; color: var(--color-success); }
</style>