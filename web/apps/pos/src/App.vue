<script setup>
import { computed, ref } from 'vue'
import { PosButton } from '@pos/ui'

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
const signedIn = computed(() => Boolean(token.value && manifest.value))
const total = computed(() => cart.value.reduce((sum, line) => sum + line.price_minor * line.quantity, 0))
const change = computed(() => Math.max(0, Number(received.value || 0) - total.value))

async function request(path, options = {}) {
  const response = await fetch(path, { ...options, headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token.value}`, ...(options.headers ?? {}) } })
  const data = await response.json().catch(() => ({}))
  if (!response.ok) throw Object.assign(new Error(data.error?.message ?? 'Request could not be completed'), { status: response.status })
  return data
}
async function login() {
  error.value = ''; busy.value = true
  try {
    const response = await fetch('/api/v1/auth/login', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ tenant_id: tenantId.value, email: email.value, password: password.value }) })
    const data = await response.json()
    if (!response.ok) throw new Error(data.error?.message ?? 'Unable to sign in')
    token.value = data.access_token
    localStorage.setItem('pos.accessToken', token.value)
    localStorage.setItem('pos.tenantId', tenantId.value)
    await loadWorkspace()
  } catch (cause) { error.value = cause instanceof Error ? cause.message : 'Unable to sign in' }
  finally { busy.value = false }
}
async function loadWorkspace() {
  manifest.value = await request('/api/v1/me/manifest')
  items.value = (await request('/api/v1/catalog/items')).items
  report.value = await request('/api/v1/reports/daily')
  payments.value = (await request('/api/v1/payments')).payments
  await loadPaymentOperations()
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
  try {
    const data = await request('/api/v1/shifts', { method: 'POST', body: JSON.stringify({ opening_float_minor: 0 }) })
    shiftId.value = data.shift_id
    localStorage.setItem('pos.shiftId', shiftId.value)
    notice.value = 'Shift opened. You can start selling.'
  } catch (cause) { error.value = cause.message }
}
async function checkout() {
  if (!shiftId.value) { error.value = 'Open a shift before completing a sale.'; return }
  if (paymentMethod.value === 'cash' && Number(received.value) < total.value) { error.value = 'Cash received must cover the sale total.'; return }
  if (paymentMethod.value === 'mpesa_stk' && !/^\+?\d{9,15}$/.test(phone.value.replace(/[\s-]/g, ''))) { error.value = 'Enter a valid mobile phone number.'; return }
  try {
    if (!saleRequestKey.value) saleRequestKey.value = crypto.randomUUID()
    const data = await request('/api/v1/sales', { method: 'POST', headers: { 'Idempotency-Key': saleRequestKey.value }, body: JSON.stringify({ shift_id: shiftId.value, payment_method: paymentMethod.value, received_minor: paymentMethod.value === 'cash' ? Number(received.value) : 0, lines: cart.value.map((line) => ({ item_id: line.id, quantity: line.quantity })) }) })
    if (paymentMethod.value === 'cash') {
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
}
if (token.value) void loadWorkspace().catch((cause) => { signOut(); error.value = cause instanceof Error ? cause.message : 'Could not load your workspace' })
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
      <div class="workspace-heading"><div><p class="eyebrow">CASHIER MODE</p><h1>Run your day.</h1><p>Catalog, checkout, stock, and daily performance in one view.</p></div><div class="heading-actions"><PosButton v-if="!shiftId" @click="openShift">Open shift</PosButton><span v-else class="shift-badge">● Shift open</span><PosButton variant="secondary" @click="signOut">Sign out</PosButton></div></div>
      <p v-if="error" class="error" role="alert">{{ error }}</p><p v-if="notice" class="notice" role="status">{{ notice }}</p>
      <div class="summary-grid"><article class="summary-card"><span class="summary-label">TODAY'S SALES</span><strong>{{ report?.sales_count ?? 0 }}</strong><span class="summary-foot">{{ report?.total_minor ?? 0 }} KES collected</span></article><article class="summary-card"><span class="summary-label">ITEMS SOLD</span><strong>{{ report?.items_sold ?? 0 }}</strong><span class="summary-foot">Across completed sales</span></article><article class="summary-card"><span class="summary-label">PAYMENT INTENTS</span><strong>{{ payments.length }}</strong><span class="summary-foot">Cash, M-Pesa, and credit</span></article></div>
      <div class="commerce-grid"><section class="catalog-panel"><div class="panel-heading"><div><p class="eyebrow">CATALOG</p><h2>Choose an item</h2></div><input v-model="search" aria-label="Search catalog" placeholder="Search name or SKU"></div><div class="product-grid"><button v-for="item in items.filter((entry) => !search || `${entry.name} ${entry.sku}`.toLowerCase().includes(search.toLowerCase()))" :key="item.id" class="product-card" :disabled="item.kind === 'stocked' && item.stock_on_hand < 1" @click="addToCart(item)"><strong>{{ item.name }}</strong><span>{{ item.sku }}</span><b>{{ item.price_minor }} KES</b><small>{{ item.stock_on_hand }} in stock</small></button></div></section>
        <aside class="cart-panel"><div class="panel-heading"><div><p class="eyebrow">CURRENT SALE</p><h2>Cart <span>({{ cart.length }})</span></h2></div><button class="clear-button" @click="cart = []">Clear</button></div><div v-if="!cart.length" class="cart-empty">Select items from the catalog to begin.</div><div v-for="line in cart" :key="line.id" class="cart-line"><div><strong>{{ line.name }}</strong><small>{{ line.price_minor }} KES each</small></div><div class="quantity"><button @click="removeFromCart(line)" aria-label="Decrease quantity">−</button><b>{{ line.quantity }}</b><button @click="addToCart(line)" aria-label="Increase quantity">+</button></div></div><div class="cart-total"><span>Total</span><strong>{{ total }} KES</strong></div><label class="cash-label">Payment method<select v-model="paymentMethod"><option value="cash">Cash</option><option value="mpesa_stk">M-Pesa STK Push</option></select></label><template v-if="paymentMethod === 'cash'"><label class="cash-label">Cash received<input v-model="received" type="number" min="0" :placeholder="String(total)"></label><p class="change-line">Change <strong>{{ change }} KES</strong></p></template><label v-else class="cash-label">Customer M-Pesa number<input v-model="phone" type="tel" inputmode="tel" placeholder="07XXXXXXXX"></label><p v-if="pendingPayment" class="change-line" role="status">Latest M-Pesa payment: <strong>{{ pendingPayment.status }}</strong></p><PosButton class="checkout-button" :disabled="!cart.length || !shiftId" @click="checkout">{{ paymentMethod === 'cash' ? 'Complete cash sale' : 'Send M-Pesa prompt' }}</PosButton></aside>
      </div>
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
