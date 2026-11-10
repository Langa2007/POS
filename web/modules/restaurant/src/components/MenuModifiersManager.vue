<template>
  <section class="menu-manager">
    <header><div><h3><PosIcon name="chef-hat" size="20" /> Menu &amp; Modifiers</h3><p>Menu items and prices are supplied by the catalog service.</p></div><span class="admin-note"><PosIcon name="lock" size="13" /> Admin changes require owner verification</span></header>
    <p v-if="loading" class="state">Loading menu…</p><p v-else-if="error" class="state error">{{ error }}</p><p v-else-if="!dishes.length" class="state">No menu items are configured. Ask an administrator to add them.</p>
    <div v-else class="builder-layout"><div class="dishes-list"><button v-for="dish in dishes" :key="dish.id" type="button" class="dish-card" :class="{ active: selectedDish?.id === dish.id }" @click="selectDish(dish)"><strong>{{ dish.name }}</strong><span>{{ dish.category || 'Menu item' }}</span><b>KES {{ formatMoney(dish.price_minor) }}</b></button></div><div v-if="selectedDish" class="dish-config-card"><h4>{{ selectedDish.name }}</h4><p>Price: KES {{ formatMoney(selectedDish.price_minor) }}</p><label>Preparation notes<input v-model="notes" type="text" placeholder="Optional kitchen note" /></label><strong>Configured price: KES {{ formatMoney(selectedDish.price_minor) }}</strong><button type="button" class="btn-add-cart" @click="addToCart"><PosIcon name="plus" size="16" /> Add to active POS cart</button></div></div>
  </section>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { PosIcon } from '@pos/ui'
import { useApi } from '../../../../apps/pos/src/composables/useApi.js'
const emit = defineEmits(['add-item', 'notification'])
const api = useApi()
const dishes = ref([])
const selectedDish = ref(null)
const notes = ref('')
const loading = ref(true)
const error = ref('')
onMounted(async () => { try { const data = await api.get('/api/v1/restaurant/menu'); dishes.value = (data.items ?? []).filter(item => item.kind === 'service' || item.kind === 'non_inventory'); selectedDish.value = dishes.value[0] ?? null } catch { error.value = 'Menu could not be loaded. Please contact your administrator.' } finally { loading.value = false } })
function selectDish(dish) { selectedDish.value = dish; notes.value = '' }
function formatMoney(minor) { return (Number(minor) / 100).toLocaleString() }
function addToCart() { if (!selectedDish.value) return; const item = { id: selectedDish.value.id, name: notes.value ? `${selectedDish.value.name} (${notes.value})` : selectedDish.value.name, sku: selectedDish.value.sku, price_minor: Number(selectedDish.value.price_minor), unit_price_minor: Number(selectedDish.value.price_minor), quantity: 1, kind: selectedDish.value.kind }; window.dispatchEvent(new CustomEvent('pos:add-item', { detail: item })); emit('add-item', item); emit('notification', { message: 'Menu item added to the active sale', type: 'success' }) }
</script>

<style scoped>
.menu-manager{display:flex;flex-direction:column;gap:16px}header{display:flex;justify-content:space-between;align-items:flex-start;gap:12px}h3{display:flex;align-items:center;gap:8px;margin:0;color:#1e293b;font-size:18px}header p{margin:4px 0 0;color:#64748b;font-size:13px}.admin-note{display:flex;align-items:center;gap:5px;padding:7px 9px;border-radius:6px;background:#f1f5f9;color:#475569;font-size:11px}.state{padding:28px;border:1px dashed #cbd5e1;border-radius:8px;color:#64748b;text-align:center;font-size:12px}.state.error{border-color:#fecaca;color:#b91c1c}.builder-layout{display:grid;grid-template-columns:minmax(220px,1fr) minmax(280px,1.8fr);gap:18px}.dishes-list{display:flex;flex-direction:column;gap:8px}.dish-card{display:flex;flex-direction:column;align-items:flex-start;gap:4px;padding:12px;border:1px solid #e2e8f0;border-radius:8px;background:#fff;text-align:left;cursor:pointer}.dish-card.active{border-color:#ea580c;background:#fff7ed}.dish-card span{color:#64748b;font-size:11px}.dish-card b{color:#ea580c;font-size:13px}.dish-config-card{display:flex;flex-direction:column;gap:14px;padding:18px;border:1px solid #e2e8f0;border-radius:9px}.dish-config-card h4{margin:0;color:#1e293b}.dish-config-card p{margin:0;color:#64748b;font-size:12px}.dish-config-card label{display:flex;flex-direction:column;gap:6px;color:#475569;font-size:11px;font-weight:700}.dish-config-card input{padding:9px;border:1px solid #cbd5e1;border-radius:6px}.btn-add-cart{display:inline-flex;align-items:center;justify-content:center;gap:6px;padding:9px;border:0;border-radius:6px;background:#ea580c;color:#fff;font-size:12px;font-weight:700;cursor:pointer}
</style>
