<script setup>
/**
 * BackofficeView — daily report summary + reconciliation panel.
 * Extracted from App.vue monolith.
 */
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { PosButton } from '../../../../packages/ui/src/index.js'
import { useConnectivityStore } from '../stores/connectivity.js'
import { useManifestStore } from '../stores/manifest.js'
import { useToastStore } from '../stores/toast.js'
import { useApi } from '../composables/useApi.js'

const { t } = useI18n()
const connectivity = useConnectivityStore()
const manifest     = useManifestStore()
const toast        = useToastStore()
const api          = useApi()

const payments              = ref([])
const unallocatedPayments   = ref([])
const reconciliationRecords = ref([])
const reconciliationFile    = ref(null)
const reconciliationSummary = ref(null)
const busy = ref(false)

onMounted(async () => {
  if (!connectivity.online) return
  const [payRes, unallocRes, reconRes] = await Promise.allSettled([
    api.get('/api/v1/payments'),
    api.get('/api/v1/payments/unallocated'),
    api.get('/api/v1/reconciliation'),
  ])
  if (payRes.status === 'fulfilled')      payments.value            = payRes.value.payments ?? []
  if (unallocRes.status === 'fulfilled')  unallocatedPayments.value = unallocRes.value.payments ?? []
  if (reconRes.status === 'fulfilled')    reconciliationRecords.value = reconRes.value.records ?? []
})

async function chooseStatement(event) {
  const file = event.target.files?.[0]
  reconciliationFile.value = file ? await file.text() : null
}

async function importStatement() {
  if (!reconciliationFile.value) return
  busy.value = true
  try {
    const res  = await fetch('/api/v1/reconciliation/import', {
      method: 'POST',
      headers: { Authorization: `Bearer ${(await import('../stores/session.js')).useSessionStore().accessToken}`, 'Content-Type': 'text/csv' },
      body: reconciliationFile.value,
    })
    const data = await res.json().catch(() => ({}))
    if (!res.ok) throw new Error(data.error?.message ?? 'Statement import failed')
    reconciliationSummary.value = data
    reconciliationFile.value    = null
    toast.success(`Import complete: ${data.matched} matched, ${data.variances} variance, ${data.unresolved} unresolved.`)
  } catch (err) { toast.error(err.message) }
  finally { busy.value = false }
}

async function allocateUnallocated(payment) {
  const saleId = window.prompt(`Enter the sale UUID for M-Pesa receipt ${payment.provider_reference}`)
  if (!saleId) return
  try {
    await api.post(`/api/v1/payments/unallocated/${payment.id}/allocate`, { sale_id: saleId })
    toast.success(`Receipt ${payment.provider_reference} allocated.`)
    unallocatedPayments.value = unallocatedPayments.value.filter(p => p.id !== payment.id)
  } catch (err) { toast.error(err.message) }
}
</script>

<template>
  <div class="backoffice-view">
    <div class="page-heading">
      <p class="eyebrow">PAYMENTS BACK OFFICE</p>
      <h1>Reconciliation</h1>
      <p>Recent statement matches and unmatched receipts.</p>
    </div>

    <section class="recon-section">
      <form class="statement-import" @submit.prevent="importStatement">
        <label class="file-label">
          Import M-Pesa CSV statement
          <input type="file" accept=".csv,text/csv" @change="chooseStatement" />
        </label>
        <PosButton type="submit" :disabled="!reconciliationFile || busy">
          {{ busy ? 'Importing…' : 'Import statement' }}
        </PosButton>
        <small>CSV headers: provider_reference, amount_minor, statement_date (optional).</small>
      </form>

      <div class="operations-grid">
        <section>
          <h3>Unallocated receipts</h3>
          <p v-if="!unallocatedPayments.length" class="empty-state">No unallocated receipts, or your role cannot view this queue.</p>
          <article
            v-for="payment in unallocatedPayments.filter(p => p.status === 'unallocated')"
            :key="payment.id"
            class="operation-row"
          >
            <div>
              <strong>{{ payment.provider_reference }}</strong>
              <small>{{ payment.account_reference || 'No account reference' }} · {{ payment.amount_minor }} {{ manifest.currency }}</small>
            </div>
            <PosButton variant="secondary" @click="allocateUnallocated(payment)">Allocate</PosButton>
          </article>
        </section>

        <section>
          <h3>Latest statement records</h3>
          <p v-if="!reconciliationRecords.length" class="empty-state">No statement entries, or your role cannot view this report.</p>
          <article
            v-for="entry in reconciliationRecords.slice(0, 10)"
            :key="`${entry.provider_reference}-${entry.created_at}`"
            class="operation-row"
          >
            <div>
              <strong>{{ entry.provider_reference }}</strong>
              <small>Expected {{ entry.expected_minor ?? 'unmatched' }} · actual {{ entry.actual_minor }} {{ manifest.currency }}</small>
            </div>
            <span class="op-status" :data-status="entry.status">{{ entry.status }}</span>
          </article>
        </section>
      </div>
    </section>
  </div>
</template>

<style scoped>
.backoffice-view { display: grid; gap: 24px; }
.page-heading p:first-child { font-size:10px; letter-spacing:1.4px; font-weight:700; color:#73818b; }
.page-heading h1 { font: 700 32px/1.1 inherit; letter-spacing:-1px; margin:8px 0 6px; }
.page-heading p:last-child { font-size:13px; color:#84909a; margin:0; }
.recon-section { background:white; border:1px solid #e5e9ef; border-radius:16px; padding:24px; display:grid; gap:20px; }
.statement-import { display:flex; align-items:flex-end; gap:16px; flex-wrap:wrap; }
.statement-import small { flex-basis:100%; color:#697586; font-size:11px; }
.file-label { display:grid; gap:6px; font-size:11px; font-weight:650; }
.operations-grid { display:grid; grid-template-columns:repeat(2, minmax(0,1fr)); gap:24px; }
.operations-grid h3 { margin:0 0 12px; font-size:14px; }
.operation-row { display:flex; justify-content:space-between; align-items:center; gap:12px; padding:12px 0; border-top:1px solid #edf0f4; }
.operation-row div { display:grid; gap:3px; }
.operation-row small { color:#697586; font-size:11px; }
.op-status { font-size:11px; font-weight:700; text-transform:uppercase; }
.op-status[data-status=unresolved],.op-status[data-status=variance] { color:#a44716; }
.op-status[data-status=matched],.op-status[data-status=allocated]   { color:#168252; }
.empty-state { padding:20px 0; text-align:center; color:#8a969d; font-size:12px; }
@media (max-width: 700px) { .operations-grid { grid-template-columns:1fr; } }
</style>

