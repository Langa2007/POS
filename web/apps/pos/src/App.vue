<script setup>
import { computed, ref } from 'vue'
import { PosButton } from '@pos/ui'

const tenantId = ref(localStorage.getItem('pos.tenantId') ?? '')
const email = ref('')
const password = ref('')
const token = ref(localStorage.getItem('pos.accessToken') ?? '')
const manifest = ref(null)
const busy = ref(false)
const error = ref('')
const signedIn = computed(() => Boolean(token.value && manifest.value))

async function login() {
  error.value = ''; busy.value = true
  try {
    const response = await fetch('/api/v1/auth/login', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ tenant_id: tenantId.value, email: email.value, password: password.value }) })
    const data = await response.json()
    if (!response.ok) throw new Error(data.error?.message ?? 'Unable to sign in')
    token.value = data.access_token
    localStorage.setItem('pos.accessToken', token.value)
    localStorage.setItem('pos.tenantId', tenantId.value)
    await loadManifest()
  } catch (cause) { error.value = cause instanceof Error ? cause.message : 'Unable to sign in' }
  finally { busy.value = false }
}
async function loadManifest() {
  const response = await fetch('/api/v1/me/manifest', { headers: { Authorization: `Bearer ${token.value}` } })
  if (!response.ok) { signOut(); throw new Error('Session expired. Sign in again.') }
  manifest.value = await response.json()
}
function signOut() {
  token.value = ''; manifest.value = null
  localStorage.removeItem('pos.accessToken')
}
if (token.value) void loadManifest().catch((cause) => { error.value = cause instanceof Error ? cause.message : 'Could not load your workspace' })
</script>

<template>
  <main class="app-frame">
    <header class="topbar"><a class="brand" href="/" aria-label="POS platform home"><span class="brand-mark">P</span><span>platform<span class="brand-light">POS</span></span></a><span class="environment">FOUNDATION PREVIEW</span></header>
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
    <section v-else class="workspace"><p class="eyebrow">YOUR WORKSPACE</p><div class="workspace-heading"><div><h1>Welcome in.</h1><p>Your platform foundation is connected and ready to grow.</p></div><PosButton variant="secondary" @click="signOut">Sign out</PosButton></div><div class="summary-grid"><article class="summary-card"><span class="summary-label">TENANT</span><strong>{{ manifest?.tenant_id }}</strong><span class="summary-foot">Your isolated business workspace</span></article><article class="summary-card"><span class="summary-label">ACTIVE MODULES</span><strong>{{ manifest?.modules.length ?? 0 }}</strong><span class="summary-foot">Modules load as capabilities are enabled</span></article><article class="summary-card"><span class="summary-label">CAPABILITIES</span><strong>{{ manifest?.capabilities.length ?? 0 }}</strong><span class="summary-foot">Configured for your operation</span></article></div><div class="empty-state"><div class="empty-icon">✳</div><h2>Your platform starts here</h2><p>The foundation is in place. Commerce modules will appear here as they are built and enabled for your tenant.</p></div></section>
    <footer>Built for the way your business works <span>•</span> English <span>•</span> Kiswahili coming soon</footer>
  </main>
</template>
