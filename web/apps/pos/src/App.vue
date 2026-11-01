<script setup>
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { PosButton } from '@pos/ui'
import { applyPulledChanges, cacheCatalog, createConnectivityMonitor, enqueueSale, flush, getCatalog, getMetadata, listPending, listRejected, setMetadata, retryRejected, nextReceiptNumber } from '../../../packages/offline/src/index.js'

const tenantId = ref(localStorage.getItem('pos.tenantId') ?? '')
const email = ref('')
const password = ref('')
const token = ref(localStorage.getItem('pos.accessToken') ?? '')
const manifest = ref(null)
const items = ref([])
const cart = ref([])
const shiftId = ref(localStorage.getItem('pos.shiftId') ?? '')
const search = ref('')
const received = ref('')
const paymentMethod = ref('cash')
const phone = ref('')
const pendingPayment = ref(null)
const lastReceipt = ref(null)
const printEndpoint = ref(localStorage.getItem('pos.printEndpoint') ?? 'http://127.0.0.1:8765')
const printToken = ref(sessionStorage.getItem('pos.printToken') ?? '')
const saleRequestKey = ref('')
const paymentRequestKey = ref('')
const report = ref(null)
const payments = ref([])
const unallocatedPayments = ref([])
const reconciliationRecords = ref([])
const reconciliationFile = ref(null)
const reconciliationSummary = ref(null)
const busy = ref(false)
const error = ref('')
const notice = ref('')
const online = ref(typeof navigator === 'undefined' ? true : navigator.onLine)
const pendingSync = ref(0)
const rejectedSync = ref([])
const syncCursor = ref(0)
const offlinePolicy = ref({ max_offline_hours: 24, max_sale_minor: 500000, allow_negative_stock: false })
const policyHours = ref(24)
const policyMaxMinor = ref(500000)
const policyNegativeStock = ref(false)
const canManageOfflinePolicy = ref(false)
let refreshInFlight
let cartRestored = false
const signedIn = computed(() => Boolean(token.value && manifest.value))
const total = computed(() => cart.value.reduce((sum, line) => sum + line.price_minor * line.quantity, 0))
const change = computed(() => Math.max(0, Number(received.value || 0) - total.value))

async function request(path, options = {}) {
  const send = () => fetch(path, { ...options, headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token.value}`, ...(options.headers ?? {}) } })
  let response = await send()
  if (response.status === 401 && navigator.onLine && localStorage.getItem('pos.refreshToken')) {
    try { await refreshSession(); response = await send() } catch { /* Preserve the original authorization response. */ }
  }
  const data = await response.json().catch(() => ({}))
  if (!response.ok) throw Object.assign(new Error(data.error?.message ?? 'Request could not be completed'), { status: response.status })
  return data
}
async function refreshSession() {
  if (!refreshInFlight) refreshInFlight = (async () => {
    const response = await fetch('/api/v1/auth/refresh', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ refresh_token: localStorage.getItem('pos.refreshToken') }) })
    const data = await response.json().catch(() => ({}))
    if (!response.ok) throw new Error(data.error?.message ?? 'Session refresh failed')
    token.value = data.access_token
    localStorage.setItem('pos.accessToken', data.access_token)
    localStorage.setItem('pos.refreshToken', data.refresh_token)
  })().finally(() => { refreshInFlight = null })
  return refreshInFlight
}
async function refreshSyncCount() {
  pendingSync.value = (await listPending()).length
  rejectedSync.value = await listRejected()
}
async function flushSync() {
  if (!token.value || !online.value) return
  try {
    let accepted = 0
    let rejected = 0
    for (let batch = 0; batch < 10; batch += 1) {
      const result = await flush(async (operations) => request('/api/v1/sync/push', { method: 'POST', body: JSON.stringify({ operations }) }), { batchSize: 100 })
      accepted += result.accepted
      rejected += result.rejected
      if (!result.pending || result.accepted === 0) break
    }
    if (rejected) notice.value = `${rejected} offline sale(s) need attention. Review the sync errors below.`
    else if (accepted) notice.value = `${accepted} offline sale(s) synced successfully.`
    await pullSync()
    await refreshSyncCount()
  } catch {
    notice.value = 'Offline changes are saved locally and will retry when the connection returns.'
  }
}
async function pullSync() {
  let cursor = Number(await getMetadata('sync_cursor', 0)) || 0
  for (let page = 0; page < 20; page += 1) {
    const batch = await request(`/api/v1/sync/pull?cursor=${cursor}&limit=500`)
    await applyPulledChanges(batch.changes, batch.next_cursor)
    cursor = batch.next_cursor
    if (batch.changes.length < 500) break
  }
  syncCursor.value = cursor
  items.value = await getCatalog()
}
async function downloadSyncBundle() {
  const data = { created_at: new Date().toISOString(), pending: await listPending(), rejected: await listRejected(), cursor: await getMetadata('sync_cursor', 0) }
  const link = document.createElement('a')
  link.href = URL.createObjectURL(new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' }))
  link.download = `pos-sync-support-${Date.now()}.json`
  link.click()
  URL.revokeObjectURL(link.href)
}
async function retrySyncError(operation) {
  try {
    await retryRejected(operation.client_id)
  await refreshSyncCount()
    if (online.value) await flushSync()
  } catch (cause) { error.value = cause.message }
}
function scanBarcode() {
  const value = search.value.trim()
  if (!value) return
  const item = items.value.find((entry) => entry.barcode === value || entry.sku.toLowerCase() === value.toLowerCase())
  if (!item) { error.value = `No catalog item matches barcode or SKU ${value}.`; return }
  addToCart(item)
  search.value = ''
  error.value = ''
}
async function printLastReceipt() {
  if (!lastReceipt.value) return
  localStorage.setItem('pos.printEndpoint', printEndpoint.value.trim().replace(/\/$/, ''))
  sessionStorage.setItem('pos.printToken', printToken.value)
  try {
    const response = await fetch(`${printEndpoint.value.trim().replace(/\/$/, '')}/print`, { method: 'POST', headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${printToken.value}` }, body: JSON.stringify({ header: manifest.value?.business_name || 'POS', lines: lastReceipt.value.lines.map((line) => `${line.quantity} x ${line.name}  ${line.quantity * line.unit_price_minor} KES`), total: `${lastReceipt.value.total_minor} KES` }) })
    if (!response.ok) throw new Error(response.status === 401 ? 'Printer token was rejected.' : `Local printer returned HTTP ${response.status}.`)
    notice.value = `Receipt ${lastReceipt.value.number} sent to the local printer.`
  } catch (cause) { error.value = `Local printer unavailable: ${cause.message}. Use browser printing or check the local print agent.` }
}
function browserPrintReceipt() { window.print() }
async function saveOfflinePolicy() {
  try {
    offlinePolicy.value = await request('/api/v1/sync/policy', { method: 'PUT', body: JSON.stringify({ max_offline_hours: Number(policyHours.value), max_sale_minor: Number(policyMaxMinor.value), allow_negative_stock: policyNegativeStock.value }) })
    await setMetadata('offline_policy', offlinePolicy.value)
    notice.value = 'Offline sale limits saved.'
  } catch (cause) { error.value = cause.message }
}
async function login() {
  error.value = ''; busy.value = true
  try {
    const response = await fetch('/api/v1/auth/login', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ tenant_id: tenantId.value, email: email.value, password: password.value }) })
    const data = await response.json()
    if (!response.ok) throw new Error(data.error?.message ?? 'Unable to sign in')
    token.value = data.access_token
    localStorage.setItem('pos.accessToken', token.value)
    localStorage.setItem('pos.refreshToken', data.refresh_token)
    localStorage.setItem('pos.tenantId', tenantId.value)
    localStorage.setItem('pos.userId', data.user_id)
    cartRestored = false
    await loadWorkspace()
  } catch (cause) { error.value = cause instanceof Error ? cause.message : 'Unable to sign in' }
  finally { busy.value = false }
}
async function loadWorkspace() {
  const cachedManifest = await getMetadata('manifest')
  const cachedItems = await getCatalog()
  if (!cartRestored) { cart.value = await getMetadata('current_cart', []); cartRestored = true }
  if (cachedManifest) manifest.value = cachedManifest
  items.value = cachedItems
  if (online.value) {
    try {
      manifest.value = await request('/api/v1/me/manifest')
      await setMetadata('manifest', manifest.value)
      const [catalogResult, policyResult] = await Promise.allSettled([request('/api/v1/catalog/items'), request('/api/v1/sync/policy')])
      if (catalogResult.status === 'fulfilled') { items.value = catalogResult.value.items; await cacheCatalog(catalogResult.value.items) }
      else if (!items.value.length) throw catalogResult.reason
      if (policyResult.status === 'fulfilled') {
        canManageOfflinePolicy.value = true
        await setMetadata('offline_policy', policyResult.value)
        offlinePolicy.value = policyResult.value
        policyHours.value = policyResult.value.max_offline_hours
        policyMaxMinor.value = policyResult.value.max_sale_minor
        policyNegativeStock.value = policyResult.value.allow_negative_stock
      }
      await setMetadata('last_online_at', new Date().toISOString())
      await pullSync()
      const [daily, paymentRows] = await Promise.allSettled([request('/api/v1/reports/daily'), request('/api/v1/payments')])
      if (daily.status === 'fulfilled') report.value = daily.value
      if (paymentRows.status === 'fulfilled') payments.value = paymentRows.value.payments
      await loadPaymentOperations()
    } catch (cause) {
      if (cause.status === 401 || cause.status === 403) throw cause
      if (!items.value.length || !manifest.value) throw cause
      notice.value = 'Using the last saved catalog. Some live data may be out of date.'
    }
  } else if (!items.value.length || !manifest.value) {
    throw new Error('Sign in while online and load the catalog before using this terminal offline.')
  }
  await refreshSyncCount()
}
async function loadPaymentOperations() {
  const [unallocated, reconciliation] = await Promise.allSettled([
    request('/api/v1/payments/unallocated'),
    request('/api/v1/reconciliation'),
  ])
  if (unallocated.status === 'fulfilled') unallocatedPayments.value = unallocated.value.payments
  if (reconciliation.status === 'fulfilled') reconciliationRecords.value = reconciliation.value.records
}
async function importStatement() {
  if (!reconciliationFile.value) return
  error.value = ''; busy.value = true
  try {
    const response = await fetch('/api/v1/reconciliation/import', {
      method: 'POST',
      headers: { Authorization: `Bearer ${token.value}`, 'Content-Type': 'text/csv' },
      body: reconciliationFile.value,
    })
    const data = await response.json().catch(() => ({}))
    if (!response.ok) throw new Error(data.error?.message ?? 'Statement import failed')
    reconciliationSummary.value = data
    reconciliationFile.value = null
    await loadPaymentOperations()
  } catch (cause) { error.value = cause instanceof Error ? cause.message : 'Statement import failed' }
  finally { busy.value = false }
}
async function chooseStatement(event) {
  const file = event.target.files?.[0]
  if (!file) { reconciliationFile.value = null; return }
  reconciliationFile.value = await file.text()
}
async function allocateUnallocated(payment) {
  const saleId = window.prompt(`Enter the sale UUID for M-Pesa receipt ${payment.provider_reference}`)
  if (!saleId) return
  try {
    await request(`/api/v1/payments/unallocated/${payment.id}/allocate`, { method: 'POST', body: JSON.stringify({ sale_id: saleId }) })
    notice.value = `Receipt ${payment.provider_reference} allocated.`
    await loadPaymentOperations()
  } catch (cause) { error.value = cause.message }
}
function addToCart(item) {
  const existing = cart.value.find((line) => line.id === item.id)
  if (existing) existing.quantity += 1
  else cart.value.push({ ...item, quantity: 1 })
}
function removeFromCart(item) {
  const existing = cart.value.find((line) => line.id === item.id)
  if (!existing) return
  if (existing.quantity === 1) cart.value = cart.value.filter((line) => line.id !== item.id)
  else existing.quantity -= 1
}
async function openShift() {
  if (!online.value) { error.value = 'Open a shift while connected before working offline.'; return }
  try {
    const data = await request('/api/v1/shifts', { method: 'POST', body: JSON.stringify({ opening_float_minor: 0 }) })
    shiftId.value = data.shift_id
    localStorage.setItem('pos.shiftId', shiftId.value)
    localStorage.setItem('pos.branchId', data.branch_id)
    notice.value = 'Shift opened. You can start selling.'
  } catch (cause) { error.value = cause.message }
}
async function checkout() {
  if (!shiftId.value) { error.value = 'Open a shift before completing a sale.'; return }
  if (!online.value) { await checkoutOffline(); return }
  if (paymentMethod.value === 'cash' && Number(received.value) < total.value) { error.value = 'Cash received must cover the sale total.'; return }
  if (paymentMethod.value === 'mpesa_stk' && !/^\+?\d{9,15}$/.test(phone.value.replace(/[\s-]/g, ''))) { error.value = 'Enter a valid mobile phone number.'; return }
  try {
    if (!saleRequestKey.value) saleRequestKey.value = crypto.randomUUID()
    const data = await request('/api/v1/sales', { method: 'POST', headers: { 'Idempotency-Key': saleRequestKey.value }, body: JSON.stringify({ shift_id: shiftId.value, payment_method: paymentMethod.value, received_minor: paymentMethod.value === 'cash' ? Number(received.value) : 0, lines: cart.value.map((line) => ({ item_id: line.id, quantity: line.quantity })) }) })
    if (paymentMethod.value === 'cash') {
      lastReceipt.value = { number: data.receipt_number || data.sale_id.slice(0, 8).toUpperCase(), lines: cart.value.map((line) => ({ name: line.name, quantity: line.quantity, unit_price_minor: line.price_minor })), total_minor: total.value, received_minor: Number(received.value), change_minor: change.value }
      notice.value = `Sale ${data.sale_id.slice(0, 8)} completed. Change: ${change.value} KES`
      cart.value = []; received.value = ''; saleRequestKey.value = ''; await loadWorkspace(); return
    }
    if (!paymentRequestKey.value) paymentRequestKey.value = crypto.randomUUID()
    const intent = await request('/api/v1/payments/intents', { method: 'POST', headers: { 'Idempotency-Key': paymentRequestKey.value }, body: JSON.stringify({ sale_id: data.sale_id, amount_minor: total.value, phone: phone.value.replace(/[+\s-]/g, ''), account_reference: data.payment_reference }) })
    paymentRequestKey.value = ''
    pendingPayment.value = { id: intent.intent_id, status: intent.status, saleId: data.sale_id }
    notice.value = 'STK prompt sent. Ask the customer to approve it on their phone.'
    cart.value = []; received.value = ''; saleRequestKey.value = ''; await loadWorkspace(); void watchPayment(intent.intent_id)
  } catch (cause) { if (cause.status === 502) paymentRequestKey.value = ''; error.value = cause.message }
}
async function checkoutOffline() {
  if (paymentMethod.value !== 'cash') { error.value = 'Only cash sales are available while offline. M-Pesa requires a live connection.'; return }
  if (!cart.value.length || Number(received.value) < total.value) { error.value = 'Cash received must cover the sale total.'; return }
  const branchId = localStorage.getItem('pos.branchId') ?? ''
  if (!branchId) { error.value = 'Open a shift while online before using offline checkout.'; return }
  const policy = await getMetadata('offline_policy', { max_offline_hours: 24, max_sale_minor: 500000, allow_negative_stock: false })
  const offlineSince = await getMetadata('offline_since')
  const disconnectedAt = offlineSince ? new Date(offlineSince) : new Date(await getMetadata('last_online_at', new Date().toISOString()))
  if (!Number.isFinite(disconnectedAt.getTime()) || Date.now() - disconnectedAt.getTime() > policy.max_offline_hours * 60 * 60 * 1000) {
    error.value = 'Offline time limit exceeded. Reconnect and sync before selling.'; return
  }
  if (total.value <= 0 || total.value > policy.max_sale_minor) { error.value = 'This sale exceeds the offline sale limit.'; return }
  const insufficient = cart.value.find((line) => line.kind === 'stocked' && line.stock_on_hand < line.quantity)
  if (insufficient && !policy.allow_negative_stock) { error.value = `Not enough cached stock for ${insufficient.name}. Reconnect to refresh inventory.`; return }
  const occurredAt = new Date().toISOString()
  const saleId = crypto.randomUUID()
  const clientId = crypto.randomUUID()
  const terminalPrefix = (localStorage.getItem('pos.terminalId') ?? crypto.randomUUID()).replaceAll('-', '').slice(0, 10)
  const receiptPrefix = await getMetadata('receipt_prefix', null) ?? `T-${terminalPrefix}`.toUpperCase()
  await setMetadata('receipt_prefix', receiptPrefix)
  const receiptNumber = await nextReceiptNumber(receiptPrefix)
  const operation = {
    client_id: clientId,
    terminal_id: localStorage.getItem('pos.terminalId') ?? '',
    entity_type: 'sale', entity_id: saleId, operation: 'create',
    client_occurred_at: occurredAt,
    payload: { branch_id: branchId, shift_id: shiftId.value, receipt_number: receiptNumber, offline_since: disconnectedAt.toISOString(), received_minor: Number(received.value), lines: cart.value.map((line) => ({ item_id: line.id, quantity: line.quantity, discount_minor: 0, unit_price_minor: line.price_minor })) },
  }
  const saleTotal = total.value
  const receiptLines = cart.value.map((line) => ({ name: line.name, quantity: line.quantity, unit_price_minor: line.price_minor }))
  await enqueueSale(operation, cart.value.filter((line) => line.kind === 'stocked').map((line) => ({ id: line.id, quantity: line.quantity })))
  lastReceipt.value = { number: receiptNumber, lines: receiptLines, total_minor: saleTotal, received_minor: Number(received.value), change_minor: Number(received.value) - saleTotal }
  items.value = await getCatalog()
  cart.value = []; received.value = ''
  await refreshSyncCount()
  notice.value = `Offline cash sale saved on this terminal. Receipt ${receiptNumber} · ${saleTotal} KES. It will sync when online.`
}
async function watchPayment(intentId) {
  for (let attempt = 0; attempt < 90; attempt += 1) {
    await new Promise((resolve) => setTimeout(resolve, 2000))
    try {
      const status = await request(`/api/v1/payments/intents/${intentId}`)
      pendingPayment.value = { ...pendingPayment.value, status: status.status }
      if (status.status === 'captured') { paymentRequestKey.value = ''; notice.value = 'M-Pesa payment confirmed and sale completed.'; await loadWorkspace(); return }
      if (['failed', 'cancelled'].includes(status.status)) { paymentRequestKey.value = ''; error.value = `M-Pesa payment ${status.status}. Sale remains unpaid.`; await loadWorkspace(); return }
    } catch { /* A brief network interruption is retried on the next poll. */ }
  }
  notice.value = 'Payment is still pending. Check the payment list before retrying.'
}
function signOut() {
  token.value = ''; manifest.value = null; items.value = []; cart.value = []
  localStorage.removeItem('pos.accessToken')
  localStorage.removeItem('pos.refreshToken')
}
if (token.value) void loadWorkspace().catch((cause) => { if (cause.status === 401 || cause.status === 403) signOut(); error.value = cause instanceof Error ? cause.message : 'Could not load your workspace' })
onMounted(() => {
  const stop = createConnectivityMonitor((connected) => {
    online.value = connected
    if (connected) {
      void setMetadata('offline_since', null)
      void setMetadata('last_online_at', new Date().toISOString())
      void flushSync()
    } else {
      void getMetadata('offline_since').then(async (started) => {
        if (!started) await setMetadata('offline_since', await getMetadata('last_online_at', new Date().toISOString()))
      })
    }
  })
  const retryTimer = window.setInterval(() => { if (online.value && pendingSync.value) void flushSync() }, 30000)
  void Promise.all([refreshSyncCount(), getCatalog().then((cached) => { if (!items.value.length) items.value = cached }), getMetadata('manifest').then((cached) => { if (!manifest.value && cached) manifest.value = cached })])
  onUnmounted(() => { stop(); window.clearInterval(retryTimer) })
})
watch(cart, (value) => { if (cartRestored) void setMetadata('current_cart', value) }, { deep: true })
</script>

<template>
  <main class="app-frame">
    <header class="topbar"><a class="brand" href="/" aria-label="POS platform home"><span class="brand-mark">P</span><span>platform<span class="brand-light">POS</span></span></a><span class="environment">CORE COMMERCE MVP</span></header>
    <section v-if="!signedIn" class="auth-layout">
      <div class="intro"><p class="eyebrow">ONE PLATFORM. EVERY COUNTER.</p><h1>Good business<br><span>starts here.</span></h1><p class="intro-copy">A clear view of your shop, your people, and every sale — all in one place.</p><div class="intro-note"><span class="status-dot"></span> Your workspace is ready when you are.</div></div>
      <form class="login-card" @submit.prevent="login"><div class="card-heading"><span class="eyebrow">WELCOME BACK</span><h2>Sign in to your workspace</h2><p>Use the owner account created during setup.</p></div>
        <label>Tenant ID<input v-model="tenantId" autocomplete="organization" placeholder="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" required></label>
        <label>Email address<input v-model="email" type="email" autocomplete="username" placeholder="you@business.com" required></label>
        <label>Password<input v-model="password" type="password" autocomplete="current-password" placeholder="Enter your password" required></label>
        <p v-if="error" class="error" role="alert">{{ error }}</p><PosButton type="submit" :disabled="busy">{{ busy ? 'Signing in…' : 'Sign in' }} <span aria-hidden="true">→</span></PosButton>
        <p class="secure-note"><span aria-hidden="true">◈</span> Protected with tenant-scoped access</p>
      </form>
    </section>
    <section v-else class="workspace">
      <div class="workspace-heading"><div><p class="eyebrow">CASHIER MODE</p><h1>Run your day.</h1><p>Catalog, checkout, stock, and daily performance in one view.</p></div><div class="heading-actions"><span class="shift-badge" :data-offline="!online">● {{ online ? 'Online' : 'Offline' }}<template v-if="pendingSync"> · {{ pendingSync }} pending</template></span><PosButton v-if="!shiftId" @click="openShift">Open shift</PosButton><span v-else class="shift-badge">● Shift open</span><PosButton variant="secondary" @click="signOut">Sign out</PosButton></div></div>
      <p v-if="error" class="error" role="alert">{{ error }}</p><p v-if="notice" class="notice" role="status">{{ notice }}</p>
      <section v-if="pendingSync || rejectedSync.length" class="sync-panel" aria-label="Offline sync status">
        <div><strong>{{ pendingSync }} pending · {{ rejectedSync.length }} need attention</strong><small>Offline cash sales are stored on this browser until the server acknowledges them.</small></div>
        <div class="sync-actions"><PosButton variant="secondary" :disabled="!online" @click="flushSync">Sync now</PosButton><PosButton variant="secondary" @click="downloadSyncBundle">Download support bundle</PosButton></div>
        <article v-for="operation in rejectedSync" :key="operation.client_id" class="sync-error"><div><strong>Receipt {{ operation.payload?.receipt_number || operation.entity_id }}</strong><small>{{ operation.sync_error?.message || 'The server rejected this sale.' }} Sale details remain saved locally.</small></div><PosButton variant="secondary" :disabled="!online" @click="retrySyncError(operation)">Retry sync</PosButton></article>
      </section>
      <details v-if="online && canManageOfflinePolicy" class="offline-policy"><summary>Offline terminal limits</summary><form @submit.prevent="saveOfflinePolicy"><label>Maximum disconnected hours<input v-model.number="policyHours" type="number" min="1" max="168"></label><label>Maximum sale amount (minor units)<input v-model.number="policyMaxMinor" type="number" min="1"></label><label><input v-model="policyNegativeStock" type="checkbox"> Allow stock below zero during disconnected sales</label><PosButton type="submit">Save limits</PosButton></form></details>
      <div class="summary-grid"><article class="summary-card"><span class="summary-label">TODAY'S SALES</span><strong>{{ report?.sales_count ?? 0 }}</strong><span class="summary-foot">{{ report?.total_minor ?? 0 }} KES collected</span></article><article class="summary-card"><span class="summary-label">ITEMS SOLD</span><strong>{{ report?.items_sold ?? 0 }}</strong><span class="summary-foot">Across completed sales</span></article><article class="summary-card"><span class="summary-label">PAYMENT INTENTS</span><strong>{{ payments.length }}</strong><span class="summary-foot">Cash, M-Pesa, and credit</span></article></div>
      <div class="commerce-grid"><section class="catalog-panel"><div class="panel-heading"><div><p class="eyebrow">CATALOG</p><h2>Choose an item</h2></div><input v-model="search" aria-label="Scan barcode or search catalog" placeholder="Scan barcode or search" @keydown.enter.prevent="scanBarcode"></div><div class="product-grid"><button v-for="item in items.filter((entry) => !search || `${entry.name} ${entry.sku} ${entry.barcode}`.toLowerCase().includes(search.toLowerCase()))" :key="item.id" class="product-card" :disabled="item.kind === 'stocked' && item.stock_on_hand < 1" @click="addToCart(item)"><strong>{{ item.name }}</strong><span>{{ item.sku }}</span><b>{{ item.price_minor }} KES</b><small>{{ item.stock_on_hand }} in stock</small></button></div></section>
        <aside class="cart-panel"><div class="panel-heading"><div><p class="eyebrow">CURRENT SALE</p><h2>Cart <span>({{ cart.length }})</span></h2></div><button class="clear-button" @click="cart = []">Clear</button></div><div v-if="!cart.length" class="cart-empty">Select items from the catalog to begin.</div><div v-for="line in cart" :key="line.id" class="cart-line"><div><strong>{{ line.name }}</strong><small>{{ line.price_minor }} KES each</small></div><div class="quantity"><button @click="removeFromCart(line)" aria-label="Decrease quantity">−</button><b>{{ line.quantity }}</b><button @click="addToCart(line)" aria-label="Increase quantity">+</button></div></div><div class="cart-total"><span>Total</span><strong>{{ total }} KES</strong></div><label class="cash-label">Payment method<select v-model="paymentMethod"><option value="cash">Cash</option><option value="mpesa_stk">M-Pesa STK Push</option></select></label><template v-if="paymentMethod === 'cash'"><label class="cash-label">Cash received<input v-model="received" type="number" min="0" :placeholder="String(total)"></label><p class="change-line">Change <strong>{{ change }} KES</strong></p></template><label v-else class="cash-label">Customer M-Pesa number<input v-model="phone" type="tel" inputmode="tel" placeholder="07XXXXXXXX"></label><p v-if="pendingPayment" class="change-line" role="status">Latest M-Pesa payment: <strong>{{ pendingPayment.status }}</strong></p><PosButton class="checkout-button" :disabled="!cart.length || !shiftId" @click="checkout">{{ paymentMethod === 'cash' ? 'Complete cash sale' : 'Send M-Pesa prompt' }}</PosButton></aside>
      </div>
      <section v-if="lastReceipt" class="receipt-preview"><div><p class="eyebrow">LATEST RECEIPT</p><h2>{{ lastReceipt.number }}</h2></div><div class="print-controls"><label>Local printer URL<input v-model="printEndpoint" autocomplete="url" placeholder="http://127.0.0.1:8765"></label><label>Printer token<input v-model="printToken" type="password" autocomplete="off" placeholder="Terminal print-agent token"></label><PosButton variant="secondary" @click="printLastReceipt">Print on local device</PosButton><PosButton variant="secondary" @click="browserPrintReceipt">Browser print</PosButton></div><div class="receipt-lines"><div v-for="(line, index) in lastReceipt.lines" :key="`${line.name}-${index}`"><span>{{ line.quantity }} × {{ line.name }}</span><span>{{ line.quantity * line.unit_price_minor }} KES</span></div></div><div class="receipt-totals"><span>Total {{ lastReceipt.total_minor }} KES</span><span>Received {{ lastReceipt.received_minor }} KES</span><strong>Change {{ lastReceipt.change_minor }} KES</strong></div></section>
      <section class="payments-ops">
        <div class="panel-heading"><div><p class="eyebrow">PAYMENTS BACK OFFICE</p><h2>Reconciliation</h2></div><span class="summary-foot">Recent statement matches and unmatched receipts</span></div>
        <form class="statement-import" @submit.prevent="importStatement"><label class="cash-label">Import M-Pesa CSV statement<input type="file" accept=".csv,text/csv" @change="chooseStatement"></label><PosButton type="submit" :disabled="!reconciliationFile || busy">{{ busy ? 'Importing…' : 'Import statement' }}</PosButton><small>CSV headers: provider_reference, amount_minor, statement_date (optional).</small></form>
        <p v-if="reconciliationSummary" class="notice" role="status">Import result: {{ reconciliationSummary.matched }} matched, {{ reconciliationSummary.variances }} variance, {{ reconciliationSummary.unresolved }} unresolved.</p>
        <div class="operations-grid"><section><h3>Unallocated receipts</h3><p v-if="!unallocatedPayments.length" class="cart-empty">No unallocated receipts found, or your role cannot view this queue.</p><article v-for="payment in unallocatedPayments.filter((entry) => entry.status === 'unallocated')" :key="payment.id" class="operation-row"><div><strong>{{ payment.provider_reference }}</strong><small>{{ payment.account_reference || 'No account reference' }} · {{ payment.amount_minor }} KES</small></div><PosButton variant="secondary" @click="allocateUnallocated(payment)">Allocate</PosButton></article></section><section><h3>Latest statement records</h3><p v-if="!reconciliationRecords.length" class="cart-empty">No statement entries found, or your role cannot view this report.</p><article v-for="entry in reconciliationRecords.slice(0, 8)" :key="`${entry.provider_reference}-${entry.created_at}`" class="operation-row"><div><strong>{{ entry.provider_reference }}</strong><small>Expected {{ entry.expected_minor ?? 'unmatched' }} · actual {{ entry.actual_minor }} KES</small></div><span class="operation-status" :data-status="entry.status">{{ entry.status }}</span></article></section></div>
      </section>
    </section>
    <footer>Built for the way your business works <span>•</span> English <span>•</span> Kiswahili coming soon</footer>
  </main>
</template>
