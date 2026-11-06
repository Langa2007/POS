/**
 * Manifest store — /me/manifest response, capabilities, permissions, nav.
 *
 * Owns: manifest data fetched after login.
 */
import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { getMetadata, setMetadata } from '../../../../packages/offline/src/index.js'
import { useSessionStore } from './session.js'

export const useManifestStore = defineStore('manifest', () => {
  const session  = useSessionStore()

  // ── state ─────────────────────────────────────────────────────────────────
  const manifest    = ref(null)
  const loading     = ref(false)
  const loadError   = ref('')

  // ── getters ───────────────────────────────────────────────────────────────
  const businessName  = computed(() => manifest.value?.business_name ?? '')
  const capabilities  = computed(() => manifest.value?.capabilities  ?? [])
  const permissions   = computed(() => manifest.value?.permissions   ?? [])
  const currency      = computed(() => manifest.value?.branch?.currency ?? 'KES')

  function can(permission) { return permissions.value.includes(permission) }
  function has(capability) { return capabilities.value.includes(capability) }

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

  return { manifest, loading, loadError, businessName, capabilities, permissions, currency, can, has, load, clear }
})

