<script setup>
/**
 * CheckoutView — cashier mode: catalog grid + cart + tender + receipt.
 * Extracted from the App.vue monolith and wired to stores.
 */
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { PosButton, PosInput, PosSelect, PosDialog } from '../../../../packages/ui/src/index.js'
import { useSessionStore } from '../stores/session.js'
import { useConnectivityStore } from '../stores/connectivity.js'
import { useManifestStore } from '../stores/manifest.js'
import { useToastStore } from '../stores/toast.js'
import { useApi } from '../composables/useApi.js'
import {
  cacheCatalog, enqueueSale, getCatalog, getMetadata, nextReceiptNumber, setMetadata,
} from '../../../../packages/offline/src/index.js'
import SyncPanel from '../components/SyncPanel.vue'

const { t } = useI18n()
const session     = useSessionStore()
const connectivity= useConnectivityStore()
const manifest    = useManifestStore()
const toast       = useToastStore()
const api         = useApi()

// ── Catalog & cart ─────────────────────────────────────────────────────────
const items       = ref([])
const cart        = ref([])
const search      = ref('')
const shiftId     = ref(localStorage.getItem('pos.shiftId') ?? '')

// ── Tender ──────────────────────────────────────────────────────────────────
const paymentMethod = ref('cash')
const received      = ref('')
const phone         = ref('')
const pendingPayment= ref(null)
let   saleKey       = ''
let   paymentKey    = ''
let scanTimer

// ── Receipt & printer ───────────────────────────────────────────────────────
const lastReceipt   = ref(null)
const printEndpoint = ref(localStorage.getItem('pos.printEndpoint') ?? 'http://127.0.0.1:8765')
const printToken    = ref(sessionStorage.getItem('pos.printToken')  ?? '')

// ── Reports ─────────────────────────────────────────────────────────────────
const report = ref(null)
const productDialog = ref(false)
const manageDialog = ref(false)
const sudoDialog = ref(false)
const sudoEmail = ref('')
const sudoPassword = ref('')
const sudoToken = ref('')
const editingItem = ref(null)
const productBusy = ref(false)
const productForm = ref({ name: '', sku: '', barcode: '', price_minor: '', cost_minor: '', stock_on_hand: '', reorder_level: '' })

// ── Computed ─────────────────────────────────────────────────────────────────
const total  = computed(() => cart.value.reduce((s, l) => s + l.price_minor * l.quantity, 0))
const change = computed(() => Math.max(0, Number(received.value || 0) - total.value))
const filteredItems = computed(() => {
  const q = search.value.toLowerCase()
  return q ? items.value.filter(i => `${i.name} ${i.sku} ${i.barcode ?? ''}`.toLowerCase().includes(q)) : items.value
})
function formatMoney(minor) {
  return (Number(minor || 0) / 100).toFixed(2)
}

// ── Lifecycle ────────────────────────────────────────────────────────────────
onMounted(async () => {
  // Load catalog from cache first
  const cached = await getCatalog()
  if (cached.length) items.value = cached
  const cachedCart = await getMetadata('current_cart', [])
  if (cachedCart.length) cart.value = cachedCart

  if (connectivity.online) {
    try {
      const [catalogRes, reportRes] = await Promise.allSettled([
        api.get('/api/v1/catalog/items'),
        api.get('/api/v1/reports/daily'),
      ])
      if (catalogRes.status === 'fulfilled') {
        items.value = catalogRes.value.items
        await cacheCatalog(catalogRes.value.items)
      }
      if (reportRes.status === 'fulfilled') report.value = reportRes.value
    } catch { /* use cached */ }
  }
})

// ── Cart helpers ─────────────────────────────────────────────────────────────
function addToCart(item) {
  const existing = cart.value.find(l => l.id === item.id)
  const nextQuantity = (existing?.quantity ?? 0) + 1
  if (item.kind === 'stocked' && nextQuantity > item.stock_on_hand) {
    toast.warning(`Only ${item.stock_on_hand} ${item.name} available.`)
    return
  }
  if (existing) existing.quantity = nextQuantity
  else cart.value.push({ ...item, quantity: 1 })
  void setMetadata('current_cart', cart.value)
}
function removeFromCart(item) {
  const existing = cart.value.find(l => l.id === item.id)
  if (!existing) return
  if (existing.quantity === 1) cart.value = cart.value.filter(l => l.id !== item.id)
  else existing.quantity -= 1
  void setMetadata('current_cart', cart.value)
}
function clearCart() { cart.value = []; void setMetadata('current_cart', []) }

function openProductDialog() {
  productForm.value = { name: '', sku: '', barcode: '', price_minor: '', cost_minor: '', stock_on_hand: '', reorder_level: '' }
  productDialog.value = true
}
function openManageDialog() { sudoEmail.value = ''; sudoPassword.value = ''; sudoDialog.value = true }
async function unlockCatalog() {
  try {
    const response = await fetch('/api/v1/auth/login', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ tenant_id: session.tenantId, email: sudoEmail.value, password: sudoPassword.value }) })
    const data = await response.json()
    if (!response.ok) throw new Error(data.error?.message ?? 'Owner verification failed')
    sudoToken.value = data.access_token
    sudoDialog.value = false
    manageDialog.value = true
  } catch (err) { toast.error(err.message) }
}
function editItem(item) {
  editingItem.value = { ...item, price_minor: (item.price_minor / 100).toFixed(2), cost_minor: (item.cost_minor / 100).toFixed(2) }
}
async function saveItem() {
  try {
    const form = editingItem.value
    const data = await api.put(`/api/v1/catalog/items/${form.id}`, { name: form.name.trim(), sku: form.sku.trim(), barcode: form.barcode?.trim() ?? '', price_minor: Math.round(Number(form.price_minor) * 100), cost_minor: Math.round(Number(form.cost_minor || 0) * 100), reorder_level: Number(form.reorder_level), kind: form.kind }, { headers: { Authorization: `Bearer ${sudoToken.value}` } })
    const index = items.value.findIndex(item => item.id === form.id)
    if (index >= 0) items.value[index] = { ...items.value[index], ...data, barcode: form.barcode, cost_minor: Math.round(Number(form.cost_minor || 0) * 100) }
    await cacheCatalog(items.value)
    editingItem.value = null
    toast.success('Product updated.')
  } catch (err) { toast.error(err.message) }
}
async function deleteItem(item) {
  if (!window.confirm(`Delete ${item.name} from the catalog? Existing sales are preserved.`)) return
  try {
    await api.delete(`/api/v1/catalog/items/${item.id}`, { headers: { Authorization: `Bearer ${sudoToken.value}` } })
    items.value = items.value.filter(entry => entry.id !== item.id)
    await cacheCatalog(items.value)
    toast.success('Product removed from the catalog.')
  } catch (err) { toast.error(err.message) }
}

async function createProduct() {
  productBusy.value = true
  try {
    const sellingPrice = Number(productForm.value.price_minor)
    const costPrice = Number(productForm.value.cost_minor || 0)
    if (!Number.isFinite(sellingPrice) || sellingPrice < 0) throw new Error('Enter a valid selling price.')
    if (!Number.isFinite(costPrice) || costPrice < 0) throw new Error('Enter a valid item cost.')
    const data = await api.post('/api/v1/catalog/items', {
      name: productForm.value.name.trim(),
      sku: productForm.value.sku.trim(),
      barcode: productForm.value.barcode.trim(),
      price_minor: Math.round(sellingPrice * 100),
      cost_minor: Math.round(costPrice * 100),
      stock_on_hand: Number(productForm.value.stock_on_hand || 0),
      reorder_level: Number(productForm.value.reorder_level || 0),
      kind: 'stocked',
    })
    const primaryBarcode = productForm.value.barcode.trim()
    const item = { ...data, barcode: primaryBarcode, barcodes: primaryBarcode ? [primaryBarcode] : [], currency: manifest.currency, kind: 'stocked', reorder_level: Number(productForm.value.reorder_level || 0) }
    items.value = [...items.value, item].sort((a, b) => a.name.localeCompare(b.name))
    await cacheCatalog(items.value)
    productDialog.value = false
    toast.success(`${item.name} added to the catalog.`)
  } catch (err) {
    toast.error(err.message)
  } finally {
    productBusy.value = false
  }
}

function scanBarcode() {
  const val = search.value.trim()
  if (!val) return
  const item = items.value.find(i => i.barcode === val || i.barcodes?.includes(val) || i.sku?.toLowerCase() === val.toLowerCase())
  if (!item) { toast.error(t('catalog.noMatch', { value: val })); return }
  addToCart(item)
  search.value = ''
}
async function addBarcode(item) {
  const barcode = window.prompt(`Add another barcode for ${item.name}`)
  if (!barcode?.trim()) return
  try {
    await api.post(`/api/v1/catalog/items/${item.id}/barcodes`, { barcode: barcode.trim() }, { headers: { Authorization: `Bearer ${sudoToken.value}` } })
    item.barcodes = [...new Set([...(item.barcodes ?? []), barcode.trim()])]
    await cacheCatalog(items.value)
    toast.success('Barcode linked to this product.')
  } catch (err) { toast.error(err.message) }
}
function handleSearchInput() {
  clearTimeout(scanTimer)
  const value = search.value.trim()
  if (!value) return
  const item = items.value.find(entry => entry.barcode === value)
  if (item) scanTimer = setTimeout(() => scanBarcode(), 120)
}

// ── Shift ────────────────────────────────────────────────────────────────────
async function openShift() {
  if (!connectivity.online) { toast.warning(t('shift.openWhileOnline')); return }
  try {
    const data = await api.post('/api/v1/shifts', { opening_float_minor: 0 })
    shiftId.value = data.shift_id
    localStorage.setItem('pos.shiftId', data.shift_id)
    localStorage.setItem('pos.branchId', data.branch_id)
    toast.success(t('shift.opened'))
  } catch (err) { toast.error(err.message) }
}

// ── Checkout ─────────────────────────────────────────────────────────────────
async function checkout() {
  if (!shiftId.value) { toast.warning(t('shift.openFirst')); return }
  if (!connectivity.online) { await checkoutOffline(); return }
  if (paymentMethod.value === 'cash' && Number(received.value) < total.value) {
    toast.warning(t('checkout.cashMustCover')); return
  }
  if (paymentMethod.value === 'mpesa_stk' && !/^\+?\d{9,15}$/.test(phone.value.replace(/[\s-]/g, ''))) {
    toast.warning(t('checkout.invalidPhone')); return
  }
  try {
    const soldLines = cart.value.map(line => ({ ...line }))
    if (!saleKey) saleKey = crypto.randomUUID()
    const data = await api.post('/api/v1/sales',
      { shift_id: shiftId.value, payment_method: paymentMethod.value,
        received_minor: paymentMethod.value === 'cash' ? Number(received.value) : 0,
        lines: cart.value.map(l => ({ item_id: l.id, quantity: l.quantity })) },
      { headers: { 'Idempotency-Key': saleKey } },
    )
    saleKey = ''
    if (paymentMethod.value === 'cash') {
      lastReceipt.value = buildReceipt(data.receipt_number || data.sale_id.slice(0, 8).toUpperCase())
      await applyLocalStockUpdate(soldLines)
      toast.success(t('checkout.completed', { id: data.sale_id.slice(0, 8), change: change.value }))
      clearCart(); received.value = ''
      return
    }
    if (!paymentKey) paymentKey = crypto.randomUUID()
    const intent = await api.post('/api/v1/payments/intents',
      { sale_id: data.sale_id, amount_minor: total.value, phone: phone.value.replace(/[+\s-]/g, ''), account_reference: data.payment_reference },
      { headers: { 'Idempotency-Key': paymentKey } },
    )
    paymentKey = ''
    pendingPayment.value = { id: intent.intent_id, status: intent.status }
    await applyLocalStockUpdate(soldLines)
    toast.info(t('checkout.promptSent'))
    clearCart(); received.value = ''
    void watchPayment(intent.intent_id)
  } catch (err) {
    if (err.status === 502) paymentKey = ''
    toast.error(err.message)
  }

  async function applyLocalStockUpdate(soldLines) {
    const sold = new Map(soldLines.map(line => [line.id, line.quantity]))
    items.value = items.value.map(item => sold.has(item.id)
      ? { ...item, stock_on_hand: Math.max(0, item.stock_on_hand - sold.get(item.id)) }
      : item)
    await cacheCatalog(items.value)
  }
}

function buildReceipt(number) {
  return {
    number,
    lines: cart.value.map(l => ({ name: l.name, quantity: l.quantity, unit_price_minor: l.price_minor })),
    total_minor:    total.value,
    received_minor: Number(received.value),
    change_minor:   change.value,
  }
}

async function checkoutOffline() {
  if (paymentMethod.value !== 'cash') { toast.warning(t('checkout.offlineOnly')); return }
  if (!cart.value.length || Number(received.value) < total.value) { toast.warning(t('checkout.cashMustCover')); return }
  const branchId = localStorage.getItem('pos.branchId') ?? ''
  if (!branchId) { toast.warning(t('shift.openWhileOnline')); return }

  const policy       = await getMetadata('offline_policy', { max_offline_hours: 24, max_sale_minor: 500000, allow_negative_stock: false })
  const offlineSince = await getMetadata('offline_since')
  const disconnectedAt = offlineSince ? new Date(offlineSince) : new Date(await getMetadata('last_online_at', new Date().toISOString()))
  if (!Number.isFinite(disconnectedAt.getTime()) || Date.now() - disconnectedAt.getTime() > policy.max_offline_hours * 3_600_000) {
    toast.error(t('checkout.offlineLimit')); return
  }
  if (total.value <= 0 || total.value > policy.max_sale_minor) { toast.error(t('checkout.overLimit')); return }
  const insufficient = cart.value.find(l => l.kind === 'stocked' && l.stock_on_hand < l.quantity)
  if (insufficient && !policy.allow_negative_stock) { toast.error(`Not enough cached stock for ${insufficient.name}.`); return }

  const saleId     = crypto.randomUUID()
  const clientId   = crypto.randomUUID()
  const termPrefix = (localStorage.getItem('pos.terminalId') ?? crypto.randomUUID()).replaceAll('-', '').slice(0, 10)
  const rcptPrefix = await getMetadata('receipt_prefix', null) ?? `T-${termPrefix}`.toUpperCase()
  await setMetadata('receipt_prefix', rcptPrefix)
  const rcptNumber = await nextReceiptNumber(rcptPrefix)

  await enqueueSale({
    client_id: clientId, terminal_id: localStorage.getItem('pos.terminalId') ?? '',
    entity_type: 'sale', entity_id: saleId, operation: 'create',
    client_occurred_at: new Date().toISOString(),
    payload: { branch_id: branchId, shift_id: shiftId.value, receipt_number: rcptNumber,
      offline_since: disconnectedAt.toISOString(), received_minor: Number(received.value),
      lines: cart.value.map(l => ({ item_id: l.id, quantity: l.quantity, discount_minor: 0, unit_price_minor: l.price_minor })) },
  }, cart.value.filter(l => l.kind === 'stocked').map(l => ({ id: l.id, quantity: l.quantity })))

  lastReceipt.value = buildReceipt(rcptNumber)
  items.value = await getCatalog()
  clearCart(); received.value = ''
  await connectivity.refreshCounts()
  toast.success(t('checkout.offlineSaved', { number: rcptNumber, total: total.value }))
}

async function watchPayment(intentId) {
  for (let i = 0; i < 90; i++) {
    await new Promise(r => setTimeout(r, 2000))
    try {
      const status = await api.get(`/api/v1/payments/intents/${intentId}`)
      pendingPayment.value = { ...pendingPayment.value, status: status.status }
      if (status.status === 'captured') { toast.success(t('checkout.promptSent')); return }
      if (['failed', 'cancelled'].includes(status.status)) { toast.error(`M-Pesa payment ${status.status}.`); return }
    } catch { /* retry */ }
  }
  toast.warning('Payment still pending. Check the payment list before retrying.')
}

// ── Printing ─────────────────────────────────────────────────────────────────
async function printLastReceipt() {
  if (!lastReceipt.value) return
  localStorage.setItem('pos.printEndpoint', printEndpoint.value.trim().replace(/\/$/, ''))
  sessionStorage.setItem('pos.printToken', printToken.value)
  try {
    const res = await fetch(`${printEndpoint.value.trim().replace(/\/$/, '')}/print`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${printToken.value}` },
      body: JSON.stringify({
        header: manifest.businessName || 'POS',
        lines: lastReceipt.value.lines.map(l => `${l.quantity} x ${l.name}  ${l.quantity * l.unit_price_minor} ${manifest.currency}`),
        total: `${lastReceipt.value.total_minor} ${manifest.currency}`,
      }),
    })
    if (!res.ok) throw new Error(res.status === 401 ? t('receipt.printerRejected') : `HTTP ${res.status}`)
    toast.success(t('receipt.sent', { number: lastReceipt.value.number }))
  } catch (err) {
    toast.error(t('receipt.printerUnavailable', { reason: err.message }))
  }
}
function browserPrint() { window.print() }
</script>

<template>
  <div class="checkout-view">
    <!-- Summary strip -->
    <div class="summary-grid">
      <article class="summary-card">
        <span class="summary-label">{{ t('reports.dailySales').toUpperCase() }}</span>
        <strong>{{ report?.sales_count ?? 0 }}</strong>
        <span class="summary-foot">{{ report?.total_minor ?? 0 }} {{ manifest.currency }} collected</span>
      </article>
      <article class="summary-card">
        <span class="summary-label">{{ t('reports.itemsSold').toUpperCase() }}</span>
        <strong>{{ report?.items_sold ?? 0 }}</strong>
        <span class="summary-foot">{{ t('reports.acrossCompleted') }}</span>
      </article>
      <article class="summary-card">
        <span class="summary-label">{{ t('reports.paymentIntents').toUpperCase() }}</span>
        <strong>{{ pendingPayment ? 1 : 0 }}</strong>
        <span class="summary-foot">{{ pendingPayment ? pendingPayment.status : t('reports.cashMpesaCredit') }}</span>
      </article>
    </div>

    <!-- Shift bar -->
    <div class="shift-bar">
      <PosButton v-if="!shiftId" @click="openShift">{{ t('shift.open') }}</PosButton>
      <span v-else class="shift-open-badge">● {{ t('shift.alreadyOpen') }}</span>
    </div>

    <!-- Sync panel -->
    <SyncPanel />

    <!-- Commerce grid -->
    <div class="commerce-grid">
      <!-- Catalog -->
      <section class="catalog-panel" aria-label="t('catalog.title')">
        <div class="panel-heading">
          <div>
            <p class="eyebrow">{{ t('catalog.title').toUpperCase() }}</p>
            <h2>{{ t('catalog.chooseItem') }}</h2>
          </div>
          <div class="catalog-actions">
            <input
              v-model="search"
              class="search-input"
              :aria-label="t('catalog.scanSearch')"
              :placeholder="t('catalog.scanSearch')"
              @input="handleSearchInput"
              @keydown.enter.prevent="scanBarcode"
            />
            <PosButton variant="secondary" @click="openProductDialog">+ Product</PosButton>
            <PosButton variant="secondary" @click="openManageDialog">Manage</PosButton>
          </div>
        </div>
        <div class="product-grid" role="list">
          <article
            v-for="item in filteredItems" :key="item.id"
            class="product-card" role="listitem"
          >
            <strong>{{ item.name }}</strong>
            <span>{{ item.sku }}<template v-if="item.barcode"> · {{ item.barcode }}</template></span>
            <b>{{ formatMoney(item.price_minor) }} {{ manifest.currency }}</b>
            <small>{{ t('catalog.inStock', { count: item.stock_on_hand }) }}</small>
            <small v-if="item.low_stock" class="low-stock">Low stock · reorder soon</small>
          </article>
        </div>
        <p v-if="!filteredItems.length" class="catalog-empty">No products match this search. Scan a barcode or add the product.</p>
      </section>

      <!-- Cart -->
      <aside class="cart-panel">
        <div class="panel-heading">
          <div>
            <p class="eyebrow">{{ t('checkout.title').toUpperCase() }}</p>
            <h2>{{ t('checkout.cart', { count: cart.length }) }}</h2>
          </div>
          <button class="clear-button" :disabled="!cart.length" @click="clearCart">{{ t('checkout.clear') }}</button>
        </div>

        <div v-if="!cart.length" class="cart-empty">{{ t('checkout.empty') }}</div>
        <div v-for="line in cart" :key="line.id" class="cart-line">
          <div>
            <strong>{{ line.name }}</strong>
            <small>{{ formatMoney(line.price_minor) }} {{ manifest.currency }} each</small>
          </div>
          <div class="quantity">
            <button @click="removeFromCart(line)" :aria-label="`Decrease ${line.name} quantity`">−</button>
            <b>{{ line.quantity }}</b>
            <button :disabled="line.kind === 'stocked' && line.quantity >= line.stock_on_hand" @click="addToCart(line)" :aria-label="`Increase ${line.name} quantity`">+</button>
          </div>
          <strong class="line-total">{{ formatMoney(line.price_minor * line.quantity) }} {{ manifest.currency }}</strong>
        </div>

        <div class="cart-total">
          <span>{{ t('checkout.total') }}</span>
          <strong>{{ formatMoney(total) }} {{ manifest.currency }}</strong>
        </div>

        <PosSelect
          v-model="paymentMethod"
          :label="t('checkout.paymentMethod')"
          :options="[{ value: 'cash', label: t('checkout.cash') }, { value: 'mpesa_stk', label: t('checkout.mpesaStk') }]"
          style="margin-top:14px"
        />

        <template v-if="paymentMethod === 'cash'">
          <PosInput
            v-model="received"
            :label="t('checkout.cashReceived')"
            type="number" min="0"
            :placeholder="String(total)"
            style="margin-top:12px"
          />
          <p class="change-line">{{ t('checkout.change') }} <strong>{{ formatMoney(change) }} {{ manifest.currency }}</strong></p>
        </template>
        <PosInput
          v-else
          v-model="phone"
          :label="t('checkout.mpesaPhone')"
          type="tel" inputmode="tel"
          :placeholder="t('checkout.mpesaPhonePlaceholder')"
          style="margin-top:12px"
        />

        <p v-if="pendingPayment" class="change-line" role="status">
          {{ t('checkout.paymentStatus', { status: pendingPayment.status }) }}
        </p>

        <PosButton
          class="checkout-button"
          :disabled="!cart.length || !shiftId"
          @click="checkout"
        >
          {{ paymentMethod === 'cash' ? t('checkout.completeCash') : t('checkout.sendPrompt') }}
        </PosButton>
      </aside>
    </div>

    <!-- Receipt preview -->
    <section v-if="lastReceipt" class="receipt-preview" aria-label="t('receipt.title')">
      <div>
        <p class="eyebrow">{{ t('receipt.title').toUpperCase() }}</p>
        <h2>{{ lastReceipt.number }}</h2>
      </div>
      <div class="print-controls">
        <PosInput v-model="printEndpoint" :label="t('receipt.printerUrl')" autocomplete="url" placeholder="http://127.0.0.1:8765" />
        <PosInput v-model="printToken"    :label="t('receipt.printerToken')" type="password" autocomplete="off" />
        <PosButton variant="secondary" @click="printLastReceipt">{{ t('receipt.printLocal') }}</PosButton>
        <PosButton variant="secondary" @click="browserPrint">{{ t('receipt.printBrowser') }}</PosButton>
      </div>
      <div class="receipt-lines">
        <div v-for="(line, i) in lastReceipt.lines" :key="`${line.name}-${i}`">
          <span>{{ line.quantity }} × {{ line.name }}</span>
          <span>{{ formatMoney(line.quantity * line.unit_price_minor) }} {{ manifest.currency }}</span>
        </div>
      </div>
      <div class="receipt-totals">
        <span class="rt-row"><span>Total</span><span>{{ formatMoney(lastReceipt.total_minor) }} {{ manifest.currency }}</span></span>
        <span class="rt-row"><span>Received</span><span>{{ formatMoney(lastReceipt.received_minor) }} {{ manifest.currency }}</span></span>
        <strong class="rt-row"><span>Change</span><span>{{ formatMoney(lastReceipt.change_minor) }} {{ manifest.currency }}</span></strong>
      </div>
    </section>

    <PosDialog :open="productDialog" title="Capture product" size="md" @close="productDialog = false">
      <form class="product-form" @submit.prevent="createProduct">
        <p class="form-helper">The selling price is fixed at checkout. Enter money in KES; the system stores cents safely in the backend.</p>
        <PosInput v-model="productForm.name" label="Product name" placeholder="e.g. Premium bread" required />
        <div class="form-grid">
          <PosInput v-model="productForm.sku" label="SKU" placeholder="e.g. BREAD-001" required />
          <PosInput v-model="productForm.barcode" label="Barcode" placeholder="Scan or type barcode" />
          <PosInput v-model="productForm.price_minor" label="Selling price (KES)" hint="Fixed customer price, e.g. 125.50" type="number" min="0" step="0.01" required />
          <PosInput v-model="productForm.cost_minor" label="Your cost (KES)" hint="What you paid per item; used for margin and stock value." type="number" min="0" step="0.01" />
          <PosInput v-model="productForm.stock_on_hand" label="Opening stock (items)" hint="How many units are currently in the shop." type="number" min="0" step="1" />
          <PosInput v-model="productForm.reorder_level" label="Reorder when stock reaches (items)" hint="Low-stock threshold; use 0 to disable the reminder." type="number" min="0" step="1" />
        </div>
        <div class="dialog-actions">
          <PosButton type="button" variant="secondary" @click="productDialog = false">Cancel</PosButton>
          <PosButton type="submit" :disabled="productBusy">{{ productBusy ? 'Saving…' : 'Save product' }}</PosButton>
        </div>
      </form>
    </PosDialog>
    <PosDialog :open="sudoDialog" title="Owner verification" size="sm" @close="sudoDialog = false">
      <form class="product-form" @submit.prevent="unlockCatalog">
        <p class="form-helper">Catalog changes require owner/admin verification.</p>
        <PosInput v-model="sudoEmail" label="Owner email" type="email" required />
        <PosInput v-model="sudoPassword" label="Owner password" type="password" required />
        <div class="dialog-actions"><PosButton type="submit">Unlock catalog</PosButton></div>
      </form>
    </PosDialog>
    <PosDialog :open="manageDialog" title="Manage catalog" size="lg" @close="manageDialog = false">
      <div v-if="!editingItem" class="manage-list">
        <p class="form-helper">These actions are separate from the customer checkout cart.</p>
        <article v-for="item in items" :key="item.id" class="manage-row">
          <div><strong>{{ item.name }}</strong><small>{{ item.sku }} · {{ formatMoney(item.price_minor) }} {{ manifest.currency }} · {{ item.stock_on_hand }} in stock</small></div>
        <div class="manage-actions"><PosButton variant="secondary" @click="addBarcode(item)">+ Barcode</PosButton><PosButton variant="secondary" @click="editItem(item)">Edit</PosButton><PosButton variant="secondary" @click="deleteItem(item)">Delete</PosButton></div>
        </article>
      </div>
      <form v-else class="product-form" @submit.prevent="saveItem">
        <p class="form-helper">Changes affect future sales. Existing receipts remain unchanged.</p>
        <PosInput v-model="editingItem.name" label="Product name" required />
        <div class="form-grid"><PosInput v-model="editingItem.sku" label="SKU" required /><PosInput v-model="editingItem.barcode" label="Primary barcode" /><PosInput v-model="editingItem.price_minor" label="Selling price (KES)" type="number" step="0.01" required /><PosInput v-model="editingItem.cost_minor" label="Your cost (KES)" type="number" step="0.01" /><PosInput v-model="editingItem.reorder_level" label="Reorder level (items)" type="number" min="0" /></div>
        <p class="form-helper">Stock is controlled by receiving and sales; it cannot be edited here.</p>
        <div class="dialog-actions"><PosButton type="button" variant="secondary" @click="editingItem = null">Cancel</PosButton><PosButton type="submit">Save changes</PosButton></div>
      </form>
    </PosDialog>
  </div>
</template>

<style scoped>
.checkout-view { display: grid; gap: 20px; }
.summary-grid { display: grid; grid-template-columns: repeat(3,1fr); gap: 15px; }
.summary-card { background:white; border:1px solid #e8ebef; border-radius:12px; padding:20px; display:grid; gap:10px; }
.summary-card strong { font: 700 22px inherit; }
.shift-bar { display: flex; align-items: center; gap: 12px; }
.shift-open-badge { font-size:11px; color:#16765d; background:#e5f5ee; padding:10px 14px; border-radius:9px; }
.commerce-grid { display:grid; grid-template-columns:minmax(0,1.5fr) minmax(300px,.8fr); gap:18px; }
.catalog-panel, .cart-panel { background:white; border:1px solid #e8ebef; border-radius:14px; padding:22px; }
.panel-heading { display:flex; justify-content:space-between; align-items:center; gap:14px; margin-bottom:18px; }
.panel-heading h2 { font-size:18px; margin:7px 0 0; }
.catalog-actions { display:flex; align-items:center; gap:8px; }
.search-input { height:42px; border:1px solid #dfe5e9; border-radius:8px; padding:0 11px; font:inherit; font-size:12px; outline:none; }
.search-input:focus { border-color:#31a88a; }
.product-grid { display:grid; grid-template-columns:repeat(auto-fill,minmax(140px,1fr)); gap:10px; }
.product-card { min-height:120px; text-align:left; background:#fbfcfd; border:1px solid #e4e9ec; border-radius:10px; padding:13px; display:grid; gap:5px; color:var(--color-ink); }
.product-card:hover { border-color:#cfd9d6; }
.product-card span,.product-card small { font-size:10px; color:#8a969d; }
.product-card b { font-size:13px; color:#13765f; }
.product-card span { overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.product-card .low-stock { color:#b45309; font-weight:700; }
.catalog-empty { padding:32px 10px; text-align:center; color:#8a969d; font-size:12px; }
.cart-empty { padding:36px 10px; text-align:center; color:#8a969d; font-size:12px; }
.cart-line { display:flex; justify-content:space-between; align-items:center; border-bottom:1px solid #eef0f2; padding:11px 0; gap:10px; }
.cart-line > div:first-child { display:grid; gap:3px; }
.cart-line small { font-size:10px; color:#8a969d; }
.quantity { display:flex; align-items:center; gap:9px; }
.quantity button,.clear-button { border:0; background:#edf3f1; color:#176c57; border-radius:6px; min-width:28px; height:28px; cursor:pointer; }
.quantity button:disabled { opacity:.4; cursor:not-allowed; }
.line-total { margin-left:auto; white-space:nowrap; font-size:12px; color:#176c57; }
.clear-button { background:transparent; font-size:11px; }
.cart-total { display:flex; justify-content:space-between; border-top:2px solid #dfe9e5; margin-top:16px; padding-top:14px; font-size:14px; }
.cart-total strong { font-size:20px; color:#126b55; }
.change-line { font-size:11px; color:#7d898f; text-align:right; margin-top:8px; }
.change-line strong { color:#126b55; }
.checkout-button { width:100%; margin-top:12px; }
/* Receipt */
.receipt-preview { display:grid; grid-template-columns:1fr auto; gap:12px; padding:20px; border:1px dashed #aab5c4; border-radius:12px; background:white; }
.print-controls { grid-column:1/-1; display:flex; align-items:flex-end; gap:10px; flex-wrap:wrap; }
.receipt-lines,.receipt-totals { grid-column:1/-1; display:grid; gap:8px; }
.receipt-lines > div { display:flex; justify-content:space-between; font-size:13px; }
.receipt-totals { padding-top:10px; border-top:1px solid #e2e8f0; }
.rt-row { display:flex; justify-content:space-between; font-size:13px; }
.product-form { display:grid; gap:16px; }
.form-helper { margin:0; color:#697586; font-size:12px; line-height:1.5; }
.form-grid { display:grid; grid-template-columns:repeat(2,minmax(0,1fr)); gap:14px; }
.dialog-actions { display:flex; justify-content:flex-end; gap:10px; padding-top:4px; }
.manage-list { display:grid; gap:8px; }
.manage-row { display:flex; justify-content:space-between; align-items:center; gap:12px; padding:12px 0; border-bottom:1px solid #edf0f4; }
.manage-row > div:first-child { display:grid; gap:4px; }
.manage-row small { color:#697586; font-size:11px; }
.manage-actions { display:flex; gap:8px; }
/* Responsive */
@media (max-width: 900px) { .commerce-grid { grid-template-columns:1fr; } .cart-panel { order:-1; } }
@media (max-width: 600px) { .summary-grid { grid-template-columns:1fr; } .catalog-actions { align-items:stretch; flex-direction:column; } .form-grid { grid-template-columns:1fr; } .line-total { display:none; } }
@media print {
  body * { visibility:hidden!important; }
  .receipt-preview,.receipt-preview * { visibility:visible!important; }
  .receipt-preview { position:absolute; inset:0 auto auto 0; width:72mm; padding:4mm; border:0; font:12px monospace; }
  .print-controls { display:none!important; }
}
</style>
