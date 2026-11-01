const DB_NAME = 'pos-terminal'
const DB_VERSION = 2
const OUTBOX = 'outbox'
const CATALOG = 'catalog'
const META = 'metadata'

const memoryScopes = new Map()

function currentScope() {
  try { return `${localStorage.getItem('pos.tenantId') ?? 'unassigned'}:${localStorage.getItem('pos.userId') ?? 'anonymous'}` }
  catch { return 'unassigned:anonymous' }
}

function memoryStore() {
  const scope = currentScope()
  if (!memoryScopes.has(scope)) memoryScopes.set(scope, { outbox: new Map(), catalog: new Map(), metadata: new Map() })
  return memoryScopes.get(scope)
}

function openDatabase() {
  if (typeof indexedDB === 'undefined') return Promise.resolve(null)
  return new Promise((resolve, reject) => {
    const request = indexedDB.open(`${DB_NAME}:${currentScope()}`, DB_VERSION)
    request.onupgradeneeded = () => {
      const db = request.result
      if (!db.objectStoreNames.contains(OUTBOX)) {
        const store = db.createObjectStore(OUTBOX, { keyPath: 'client_id' })
        store.createIndex('created_at', 'created_at')
      }
      if (!db.objectStoreNames.contains(CATALOG)) db.createObjectStore(CATALOG, { keyPath: 'id' })
      if (!db.objectStoreNames.contains(META)) db.createObjectStore(META, { keyPath: 'key' })
    }
    request.onsuccess = () => resolve(request.result)
    request.onerror = () => reject(request.error ?? new Error('offline database could not be opened'))
    request.onblocked = () => reject(new Error('offline database upgrade is blocked by another open tab'))
  })
}

function runTransaction(db, stores, mode, action) {
  return new Promise((resolve, reject) => {
    const tx = db.transaction(stores, mode)
    let result
    try { result = action(tx) } catch (error) { tx.abort(); reject(error); return }
    tx.oncomplete = () => resolve(result)
    tx.onabort = () => reject(tx.error ?? new Error('offline database transaction aborted'))
    tx.onerror = () => reject(tx.error ?? new Error('offline database transaction failed'))
  })
}

function requestResult(request) {
  return new Promise((resolve, reject) => {
    request.onsuccess = () => resolve(request.result)
    request.onerror = () => reject(request.error ?? new Error('offline database request failed'))
  })
}

export async function enqueue(operation) {
  if (!operation?.client_id || !operation.entity_type || !operation.entity_id || !operation.operation || !operation.payload) {
    throw new Error('client_id, entity_type, entity_id, operation, and payload are required')
  }
  const entry = { ...operation, created_at: operation.created_at ?? new Date().toISOString(), client_occurred_at: operation.client_occurred_at ?? new Date().toISOString(), status: 'pending', attempts: 0 }
  const db = await openDatabase()
  if (!db) { memoryStore().outbox.set(entry.client_id, entry); return entry }
  await runTransaction(db, OUTBOX, 'readwrite', (tx) => tx.objectStore(OUTBOX).put(entry))
  db.close()
  return entry
}

export async function enqueueSale(operation, stockDeltas = []) {
  if (operation?.entity_type !== 'sale' || operation.operation !== 'create') throw new Error('offline operation must create a sale')
  const entry = { ...operation, created_at: operation.created_at ?? new Date().toISOString(), client_occurred_at: operation.client_occurred_at ?? new Date().toISOString(), status: 'pending', attempts: 0 }
  const db = await openDatabase()
  if (!db) {
    memoryStore().outbox.set(entry.client_id, entry)
    for (const delta of stockDeltas) { const item = memoryStore().catalog.get(delta.id); if (item) memoryStore().catalog.set(delta.id, { ...item, stock_on_hand: item.stock_on_hand - delta.quantity }) }
    return entry
  }
  await runTransaction(db, [OUTBOX, CATALOG], 'readwrite', (tx) => {
    tx.objectStore(OUTBOX).add(entry)
    const catalog = tx.objectStore(CATALOG)
    for (const delta of stockDeltas) {
      const get = catalog.get(delta.id)
      get.onsuccess = () => { if (get.result) catalog.put({ ...get.result, stock_on_hand: get.result.stock_on_hand - delta.quantity }) }
    }
  })
  db.close()
  return entry
}

async function listByStatus(status) {
  const db = await openDatabase()
  if (!db) return [...memoryStore().outbox.values()].filter((item) => item.status === status).sort((a, b) => a.created_at.localeCompare(b.created_at))
  const tx = db.transaction(OUTBOX, 'readonly')
  const entries = await requestResult(tx.objectStore(OUTBOX).index('created_at').getAll())
  db.close()
  return entries.filter((item) => item.status === status)
}

export const listPending = () => listByStatus('pending')
export const listRejected = () => listByStatus('rejected')

export async function remove(clientIds) {
  const ids = [...new Set(clientIds)]
  const db = await openDatabase()
  if (!db) { ids.forEach((id) => memoryStore().outbox.delete(id)); return }
  await runTransaction(db, OUTBOX, 'readwrite', (tx) => ids.forEach((id) => tx.objectStore(OUTBOX).delete(id)))
  db.close()
}

export const clearRejected = (clientId) => remove([clientId])

export async function retryRejected(clientId) {
  const rejected = await listRejected()
  const original = rejected.find((item) => item.client_id === clientId)
  if (!original) throw new Error('rejected operation was not found')
  const replacement = { ...original, client_id: crypto.randomUUID(), entity_id: crypto.randomUUID(), client_occurred_at: original.client_occurred_at, status: 'pending', attempts: 0 }
  const db = await openDatabase()
  if (!db) { memoryStore().outbox.set(replacement.client_id, replacement); memoryStore().outbox.delete(clientId); return replacement }
  await runTransaction(db, OUTBOX, 'readwrite', (tx) => {
    tx.objectStore(OUTBOX).add(replacement)
    tx.objectStore(OUTBOX).delete(clientId)
  })
  db.close()
  return replacement
}

export async function acknowledge(acks = []) {
  const db = await openDatabase()
  if (!db) {
    for (const ack of acks) {
      if (ack.status === 'accepted') memoryStore().outbox.delete(ack.client_id)
      else if (ack.status === 'rejected' && memoryStore().outbox.has(ack.client_id)) memoryStore().outbox.set(ack.client_id, { ...memoryStore().outbox.get(ack.client_id), status: 'rejected', sync_error: ack.error ?? null })
    }
    return
  }
  await runTransaction(db, OUTBOX, 'readwrite', (tx) => {
    const store = tx.objectStore(OUTBOX)
    for (const ack of acks) {
      if (ack.status === 'accepted') store.delete(ack.client_id)
      else if (ack.status === 'rejected') {
        const get = store.get(ack.client_id)
        get.onsuccess = () => { if (get.result) store.put({ ...get.result, status: 'rejected', sync_error: ack.error ?? null }) }
      }
    }
  })
  db.close()
}

export async function flush(push, { batchSize = 100 } = {}) {
  if (typeof push !== 'function') throw new Error('push callback is required')
  if (!Number.isInteger(batchSize) || batchSize < 1 || batchSize > 1000) throw new Error('batchSize must be between 1 and 1000')
  const pending = (await listPending()).slice(0, batchSize)
  if (!pending.length) return { accepted: 0, rejected: 0, pending: 0 }
  const response = await push(pending)
  const acks = Array.isArray(response?.acks) ? response.acks : []
  await acknowledge(acks)
  return { accepted: acks.filter((ack) => ack.status === 'accepted').length, rejected: acks.filter((ack) => ack.status === 'rejected').length, pending: (await listPending()).length }
}

export async function cacheCatalog(items) {
  if (!Array.isArray(items)) throw new Error('catalog items must be an array')
  const db = await openDatabase()
  if (!db) { memoryStore().catalog.clear(); items.forEach((item) => memoryStore().catalog.set(item.id, item)); return }
  await runTransaction(db, [CATALOG, META], 'readwrite', (tx) => {
    const catalog = tx.objectStore(CATALOG)
    catalog.clear()
    items.forEach((item) => catalog.put(item))
    tx.objectStore(META).put({ key: 'catalog_updated_at', value: new Date().toISOString() })
  })
  db.close()
}

export async function getCatalog() {
  const db = await openDatabase()
  if (!db) return [...memoryStore().catalog.values()]
  const tx = db.transaction(CATALOG, 'readonly')
  const items = await requestResult(tx.objectStore(CATALOG).getAll())
  db.close()
  return items
}

export async function applyPulledChanges(changes, cursor) {
  const db = await openDatabase()
  if (!db) {
    for (const change of changes) {
      if (change.entity_type !== 'master.catalog_item') continue
      if (change.operation === 'delete') memoryStore().catalog.delete(change.entity_id)
      else if (change.payload?.id) memoryStore().catalog.set(change.payload.id, change.payload)
    }
    memoryStore().metadata.set('sync_cursor', Math.max(Number(memoryStore().metadata.get('sync_cursor') ?? 0), Number(cursor ?? 0)))
    return
  }
  await runTransaction(db, [CATALOG, META], 'readwrite', (tx) => {
    const store = tx.objectStore(CATALOG)
    for (const change of changes) {
      if (change.entity_type !== 'master.catalog_item') continue
      if (change.operation === 'delete') store.delete(change.entity_id)
      else if (change.payload?.id) store.put(change.payload)
    }
    const meta = tx.objectStore(META)
    const get = meta.get('sync_cursor')
    get.onsuccess = () => meta.put({ key: 'sync_cursor', value: Math.max(Number(get.result?.value ?? 0), Number(cursor ?? 0)) })
  })
  db.close()
}

export async function getMetadata(key, fallback = null) {
  const db = await openDatabase()
  if (!db) return memoryStore().metadata.has(key) ? memoryStore().metadata.get(key) : fallback
  const tx = db.transaction(META, 'readonly')
  const row = await requestResult(tx.objectStore(META).get(key))
  db.close()
  return row?.value ?? fallback
}

export async function setMetadata(key, value) {
  const db = await openDatabase()
  if (!db) { memoryStore().metadata.set(key, value); return }
  await runTransaction(db, META, 'readwrite', (tx) => tx.objectStore(META).put({ key, value }))
  db.close()
}

export async function nextReceiptNumber(prefix) {
  const safePrefix = String(prefix ?? '').toUpperCase().replace(/[^A-Z0-9-]/g, '').slice(0, 12)
  if (!safePrefix) throw new Error('terminal receipt prefix is required')
  const key = `receipt_sequence:${safePrefix}`
  const db = await openDatabase()
  let sequence
  if (!db) {
    sequence = (memoryStore().metadata.get(key) ?? 0) + 1
    memoryStore().metadata.set(key, sequence)
  } else {
    await runTransaction(db, META, 'readwrite', (tx) => {
      const store = tx.objectStore(META)
      const get = store.get(key)
      get.onsuccess = () => { sequence = (get.result?.value ?? 0) + 1; store.put({ key, value: sequence }) }
    })
    db.close()
  }
  return `${safePrefix}-${String(sequence).padStart(8, '0')}`
}

export function createConnectivityMonitor(onChange) {
  const listener = () => onChange(navigator.onLine)
  window.addEventListener('online', listener)
  window.addEventListener('offline', listener)
  listener()
  return () => { window.removeEventListener('online', listener); window.removeEventListener('offline', listener) }
}
