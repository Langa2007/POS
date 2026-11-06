
import { defineStore } from 'pinia'
import { ref } from 'vue'
import { createConnectivityMonitor, flush, getMetadata, setMetadata } from '../../../../packages/offline/src/index.js'
import { useSessionStore } from './session.js'

export const useConnectivityStore = defineStore('connectivity', () => {
  const session = useSessionStore()

  // ── state ─────────────────────────────────────────────────────────────────
  const online   = ref(typeof navigator !== 'undefined' ? navigator.onLine : true)
  const pending  = ref(0)
  const rejected = ref([])
  let   _stopMonitor   = null
  let   _retryInterval = null

  // ── actions ───────────────────────────────────────────────────────────────
  async function refreshCounts() {
    const { listPending, listRejected } = await import('../../../../packages/offline/src/index.js')
    pending.value  = (await listPending()).length
    rejected.value = await listRejected()
  }

  async function sync() {
    if (!online.value || !session.accessToken) return
    try {
      let accepted = 0, rejectedCount = 0
      for (let batch = 0; batch < 10; batch++) {
        const result = await flush(
          (operations) => fetch('/api/v1/sync/push', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${session.accessToken}` },
            body: JSON.stringify({ operations }),
          }).then(r => r.json()),
          { batchSize: 100 },
        )
        accepted      += result.accepted
        rejectedCount += result.rejected
        if (!result.pending || result.accepted === 0) break
      }
      await refreshCounts()
      return { accepted, rejected: rejectedCount }
    } catch {
      // Network error — will retry on next interval
    }
  }

  function startMonitor() {
    if (_stopMonitor) return
    _stopMonitor = createConnectivityMonitor(async (connected) => {
      online.value = connected
      if (connected) {
        await setMetadata('offline_since', null)
        await setMetadata('last_online_at', new Date().toISOString())
        await sync()
      } else {
        const started = await getMetadata('offline_since')
        if (!started) await setMetadata('offline_since', await getMetadata('last_online_at', new Date().toISOString()))
      }
    })
    _retryInterval = window.setInterval(() => { if (online.value && pending.value) sync() }, 30_000)
  }

  function stopMonitor() {
    _stopMonitor?.()
    _stopMonitor = null
    if (_retryInterval) { clearInterval(_retryInterval); _retryInterval = null }
  }

  return { online, pending, rejected, refreshCounts, sync, startMonitor, stopMonitor }
})

