<template>
  <aside class="bookshop-sidebar">
    <header><div class="title"><PosIcon name="file-text" size="18" /><strong>Bookshop Counter</strong></div><span class="badge">Books</span></header>
    <label>Customer order type
      <select v-model="orderType" @change="notify">
        <option>Walk-in sale</option><option>School order</option><option>Special order</option><option>Second-hand purchase</option>
      </select>
    </label>
    <label class="check"><input v-model="reserve" type="checkbox" @change="notify" /> Reserve unavailable titles</label>
    <div class="actions">
      <button type="button" @click="notify('ISBN lookup ready')"><PosIcon name="search" size="14" /> ISBN lookup</button>
      <button type="button" @click="notify('Order slip prepared')"><PosIcon name="file-text" size="14" /> Order slip</button>
    </div>
  </aside>
</template>

<script setup>
import { ref } from 'vue'
import { PosIcon } from '@pos/ui'

defineProps({ cart: { type: Array, default: () => [] } })
const orderType = ref('Walk-in sale')
const reserve = ref(false)
function notify(message = `${orderType.value} selected`) {
  window.dispatchEvent(new CustomEvent('pos:notification', { detail: { message, type: 'info' } }))
}
</script>

<style scoped>
.bookshop-sidebar { display: flex; flex-direction: column; gap: 14px; padding: 16px; background: #fff; border: 1px solid #e2e8f0; border-radius: 12px; }
header, .title, .actions, .check { display: flex; align-items: center; gap: 8px; } header { justify-content: space-between; border-bottom: 1px solid #f1f5f9; padding-bottom: 8px; }
.title { color: #334155; font-size: 13px; } .badge { padding: 3px 7px; border-radius: 5px; background: #ede9fe; color: #6d28d9; font-size: 10px; font-weight: 700; }
label { display: flex; flex-direction: column; gap: 6px; color: #64748b; font-size: 11px; font-weight: 700; text-transform: uppercase; } select { padding: 8px; border: 1px solid #cbd5e1; border-radius: 6px; background: #f8fafc; font-size: 12px; }
.check { flex-direction: row; color: #475569; text-transform: none; } .actions { border-top: 1px solid #f1f5f9; padding-top: 10px; }
button { flex: 1; display: inline-flex; align-items: center; justify-content: center; gap: 5px; padding: 8px; border: 1px solid #cbd5e1; border-radius: 6px; background: #fff; color: #334155; font-size: 11px; font-weight: 600; cursor: pointer; } button:hover { background: #f8fafc; }
</style>
