/**
 * Manifest store — /me/manifest response, capabilities, permissions, nav, and active shop mode.
 *
 * Supports single backend powering multiple shops (e.g. Retail Supermarket and Pharmacy Chemist).
 */
import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { getMetadata, setMetadata } from '../../../../packages/offline/src/index.js'
import { useSessionStore } from './session.js'

export const AVAILABLE_SHOPS = [
  {
    id: 'retail',
    name: 'Retail Supermarket',
    badge: ' Retail Till',
    subtitle: 'Weighed goods, promotions, customer loyalty, and shelf labels',
    icon: '',
    capabilities: ['weighed_items', 'promotions', 'loyalty', 'perishables'],
    route: '/retail',
  },
  {
    id: 'pharmacy',
    name: 'Pharmacy Dispensary',
    badge: ' Pharmacy Counter',
    subtitle: 'Prescriptions, FEFO batches, controlled drugs, and clinical labels',
    icon: '',
    capabilities: ['batch_tracking', 'expiry_tracking', 'prescriptions', 'controlled_substances', 'multi_uom'],
    route: '/pharmacy',
  },
  {
    id: 'hardware',
    name: 'Hardware & Building Supplies',
    badge: ' Hardware Counter',
    subtitle: 'Multi-unit sales, cut-to-length materials, trade credit, and delivery',
    icon: '',
    capabilities: ['multi_uom', 'cut_to_length', 'credit_accounts', 'contract_pricing', 'quotes', 'delivery'],
    route: '/hardware',
  },
  {
    id: 'restaurant',
    name: 'Restaurant & Bar',
    badge: ' Restaurant Till',
    subtitle: 'Table service, kitchen orders, menu modifiers, and split bills',
    icon: '',
    capabilities: ['table_service', 'kitchen_display', 'menu_modifiers', 'split_bills', 'recipes'],
    route: '/restaurant',
  },
]

export const useManifestStore = defineStore('manifest', () => {
  const session  = useSessionStore()

  // ── state ─────────────────────────────────────────────────────────────────
  const manifest    = ref(null)
  const loading     = ref(false)
  const loadError   = ref('')
  const activeShop  = ref(localStorage.getItem('pos.activeShop') || 'retail')

  // ── getters ───────────────────────────────────────────────────────────────
  const businessName  = computed(() => {
    if (manifest.value?.business_name) return manifest.value.business_name
    const shop = AVAILABLE_SHOPS.find((s) => s.id === activeShop.value)
    return shop ? shop.name : 'POS Platform'
  })

  const currentShopInfo = computed(() => {
    return AVAILABLE_SHOPS.find((s) => s.id === activeShop.value) || AVAILABLE_SHOPS[0]
  })

  const capabilities  = computed(() => {
    const list = new Set(manifest.value?.capabilities ?? [])
    const shop = AVAILABLE_SHOPS.find((s) => s.id === activeShop.value)
    if (shop) {
      shop.capabilities.forEach((c) => list.add(c))
    }
    return Array.from(list)
  })

  const permissions   = computed(() => manifest.value?.permissions ?? [
    'retail.read', 'retail.manage',
    'pharmacy.read', 'pharmacy.dispense', 'pharmacy.manage',
    'hardware.read', 'hardware.manage', 'hardware.credit',
    'restaurant.read', 'restaurant.order', 'restaurant.manage',
  ])
  const currency      = computed(() => manifest.value?.branch?.currency ?? 'KES')

  function can(permission) { return permissions.value.includes(permission) || permissions.value.includes('*') }
  function has(capability) { return capabilities.value.includes(capability) }
  function isModuleEnabled(id) {
    if (activeShop.value === id) return true
    const mods = manifest.value?.modules ?? []
    return mods.includes(id) || mods.includes('*')
  }

  function setActiveShop(shopId) {
    activeShop.value = shopId
    localStorage.setItem('pos.activeShop', shopId)
  }

  // ── actions ───────────────────────────────────────────────────────────────
  async function load() {
    loading.value   = true
    loadError.value = ''
    // Try the cache first so the app feels instant when offline
    const cached = await getMetadata('manifest')
    if (cached) manifest.value = cached

    if (navigator.onLine) {
      try {
        const res  = await fetch('/api/v1/me/manifest', {
          headers: { Authorization: `Bearer ${session.accessToken}` },
        })
        if (res.status === 401) { session.clear(); throw new Error('session-expired') }
        const data = await res.json()
        manifest.value = data
        await setMetadata('manifest', data)
      } catch (err) {
        if (!manifest.value) { loadError.value = err.message; throw err }
        // Offline with cached manifest — degrade gracefully
      }
    } else if (!manifest.value) {
      loadError.value = 'No cached workspace. Please connect and sign in first.'
      throw new Error(loadError.value)
    }

    loading.value = false
    return manifest.value
  }

  function clear() {
    manifest.value  = null
    loadError.value = ''
  }

  return {
    manifest,
    loading,
    loadError,
    activeShop,
    currentShopInfo,
    availableShops: AVAILABLE_SHOPS,
    businessName,
    capabilities,
    permissions,
    currency,
    can,
    has,
    isModuleEnabled,
    setActiveShop,
    load,
    clear,
  }
})
