import { beforeEach, describe, expect, it } from 'vitest'
import { IDBFactory } from 'fake-indexeddb'
import { acknowledge, applyPulledChanges, cacheCatalog, enqueue, enqueueSale, flush, getCatalog, getMetadata, listPending, listRejected, nextReceiptNumber, retryRejected, setMetadata } from '../../../packages/offline/src/index.js'

let scope = 0
beforeEach(() => {
  globalThis.indexedDB = new IDBFactory()
  localStorage.setItem('pos.tenantId', `tenant-${++scope}`)
  localStorage.setItem('pos.userId', 'cashier-1')
})

describe('IndexedDB terminal storage', () => {
  it('atomically queues offline sale data and updates local stock', async () => {
    const item = { id: 'item-1', name: 'Tea', price_minor: 1250, stock_on_hand: 2, kind: 'stocked' }
    await cacheCatalog([item])
    await enqueueSale({ client_id: 'sale-op-1', entity_type: 'sale', entity_id: 'sale-1', operation: 'create', payload: { receipt_number: 'TILL-00000001' } }, [{ id: item.id, quantity: 1 }])
    expect(await listPending()).toHaveLength(1)
    expect((await getCatalog())[0].stock_on_hand).toBe(1)
  })

  it('keeps rejected operations visible and removes acknowledged accepted operations', async () => {
    await enqueue({ client_id: 'op-ok', entity_type: 'sale', entity_id: 'sale-ok', operation: 'create', payload: {} })
    await enqueue({ client_id: 'op-bad', entity_type: 'sale', entity_id: 'sale-bad', operation: 'create', payload: {} })
    const result = await flush(async () => ({ acks: [
      { client_id: 'op-ok', status: 'accepted' },
      { client_id: 'op-bad', status: 'rejected', error: { code: 'sync.stock', message: 'Stock changed' } },
    ] }))
    expect(result).toEqual({ accepted: 1, rejected: 1, pending: 0 })
    expect(await listPending()).toHaveLength(0)
    expect((await listRejected())[0].sync_error.code).toBe('sync.stock')
  })

  it('applies pulled master-data updates and tombstones with a durable cursor', async () => {
    await cacheCatalog([{ id: 'item-1', name: 'Old' }, { id: 'item-2', name: 'Removed' }])
    await applyPulledChanges([
      { entity_type: 'master.catalog_item', entity_id: 'item-1', operation: 'update', payload: { id: 'item-1', name: 'New' } },
      { entity_type: 'master.catalog_item', entity_id: 'item-2', operation: 'delete', payload: {} },
    ], 27)
    expect(await getCatalog()).toEqual([{ id: 'item-1', name: 'New' }])
    expect(await getMetadata('sync_cursor')).toBe(27)
  })

  it('allocates persistent unique terminal receipt numbers', async () => {
    expect(await nextReceiptNumber('till-a')).toBe('TILL-A-00000001')
    expect(await nextReceiptNumber('TILL-A')).toBe('TILL-A-00000002')
  })

  it('isolates queued sales and cached catalog between tenant/user scopes', async () => {
    await enqueue({ client_id: 'private-op', entity_type: 'sale', entity_id: 'sale-1', operation: 'create', payload: {} })
    await cacheCatalog([{ id: 'private-item', name: 'Private' }])
    localStorage.setItem('pos.tenantId', 'another-tenant')
    expect(await listPending()).toHaveLength(0)
    expect(await getCatalog()).toHaveLength(0)
  })

  it('does not move the sync cursor backwards', async () => {
    await setMetadata('sync_cursor', 20)
    await applyPulledChanges([], 12)
    expect(await getMetadata('sync_cursor')).toBe(20)
  })

  it('retries a rejected sale with fresh idempotency identifiers and preserves its receipt', async () => {
    await enqueue({ client_id: 'old-operation', entity_type: 'sale', entity_id: 'old-sale', operation: 'create', payload: { receipt_number: 'T-0001' } })
    await acknowledge([{ client_id: 'old-operation', status: 'rejected', error: { code: 'sync.stock', message: 'Stock changed' } }])
    const replacement = await retryRejected('old-operation')
    expect(replacement.client_id).not.toBe('old-operation')
    expect(replacement.entity_id).not.toBe('old-sale')
    expect(replacement.payload.receipt_number).toBe('T-0001')
    expect(await listRejected()).toHaveLength(0)
    expect(await listPending()).toHaveLength(1)
  })
})
