/**
 * Session store — authentication tokens, user identity, sign-in/out.
 *
 * Owns: access_token, refresh_token, tenant_id, user_id.
 * Does not hold manifest data (see useManifestStore).
 */
import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

export const useSessionStore = defineStore('session', () => {
  // ── state ─────────────────────────────────────────────────────────────────
  const accessToken  = ref(localStorage.getItem('pos.accessToken')  ?? '')
  const refreshToken = ref(localStorage.getItem('pos.refreshToken') ?? '')
  const tenantId     = ref(localStorage.getItem('pos.tenantId')     ?? '')
  const userId       = ref(localStorage.getItem('pos.userId')       ?? '')

  // ── getters ───────────────────────────────────────────────────────────────
  const isAuthenticated = computed(() => Boolean(accessToken.value))

  // ── actions ───────────────────────────────────────────────────────────────
  function setTokens({ access_token, refresh_token, tenant_id, user_id }) {
    accessToken.value  = access_token
    refreshToken.value = refresh_token
    tenantId.value     = tenant_id ?? tenantId.value
    userId.value       = user_id   ?? userId.value

    localStorage.setItem('pos.accessToken',  access_token)
    localStorage.setItem('pos.refreshToken', refresh_token)
    if (tenant_id) localStorage.setItem('pos.tenantId', tenant_id)
    if (user_id)   localStorage.setItem('pos.userId',   user_id)
  }

  async function refresh() {
    if (!refreshToken.value) throw new Error('No refresh token')
    const res  = await fetch('/api/v1/auth/refresh', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken.value }),
    })
    const data = await res.json().catch(() => ({}))
    if (!res.ok) throw new Error(data.error?.message ?? 'Session refresh failed')
    setTokens(data)
  }

  function clear() {
    accessToken.value  = ''
    refreshToken.value = ''
    userId.value       = ''
    localStorage.removeItem('pos.accessToken')
    localStorage.removeItem('pos.refreshToken')
    localStorage.removeItem('pos.userId')
  }

  return { accessToken, refreshToken, tenantId, userId, isAuthenticated, setTokens, refresh, clear }
})

